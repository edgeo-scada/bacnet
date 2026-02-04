package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/edgeo-scada/bacnet/bacnet"
)

var (
	watchObjectType string
	watchProperty   string
	watchInterval   time.Duration
	watchCOV        bool
	watchCOVLifetime uint32
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch a property for changes",
	Long: `Watch monitors a BACnet property for changes.

Two modes are available:
  - Polling: Periodically reads the property value
  - COV: Subscribes to Change of Value notifications (if supported)

Examples:
  # Poll present value every second
  edgeo-bacnet watch -d 1234 -o analog-input:1 -p present-value --interval 1s

  # Subscribe to COV notifications
  edgeo-bacnet watch -d 1234 -o analog-input:1 --cov

  # COV with custom lifetime
  edgeo-bacnet watch -d 1234 -o analog-input:1 --cov --cov-lifetime 300`,

	RunE: runWatch,
}

func init() {
	watchCmd.Flags().StringVarP(&watchObjectType, "object", "O", "", "Object type and instance (e.g., analog-input:1)")
	watchCmd.Flags().StringVarP(&watchProperty, "property", "P", "present-value", "Property identifier")
	watchCmd.Flags().DurationVar(&watchInterval, "interval", time.Second, "Polling interval")
	watchCmd.Flags().BoolVar(&watchCOV, "cov", false, "Use COV subscription instead of polling")
	watchCmd.Flags().Uint32Var(&watchCOVLifetime, "cov-lifetime", 0, "COV subscription lifetime in seconds (0 = indefinite)")

	watchCmd.MarkFlagRequired("object")
}

func runWatch(cmd *cobra.Command, args []string) error {
	if deviceID == 0 {
		return fmt.Errorf("device ID is required (-d or --device)")
	}

	// Parse object identifier
	objectID, err := parseObjectIdentifier(watchObjectType)
	if err != nil {
		return fmt.Errorf("invalid object: %w", err)
	}

	// Parse property identifier
	propID, err := parsePropertyIdentifier(watchProperty)
	if err != nil {
		return fmt.Errorf("invalid property: %w", err)
	}

	client, err := createClient()
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	// Handle interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "\nStopping watch...")
		cancel()
	}()

	fmt.Printf("Watching %s.%s on device %d\n", objectID.String(), propID.String(), deviceID)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	if watchCOV {
		return runCOVWatch(ctx, client, objectID, propID)
	}
	return runPollingWatch(ctx, client, objectID, propID)
}

func runPollingWatch(ctx context.Context, client *bacnet.Client, objectID bacnet.ObjectIdentifier, propID bacnet.PropertyIdentifier) error {
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	var lastValue interface{}

	// Read initial value
	value, err := client.ReadProperty(ctx, deviceID, objectID, propID)
	if err != nil {
		return fmt.Errorf("initial read: %w", err)
	}

	outputWatchValue(time.Now(), objectID, propID, value, true)
	lastValue = value

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			readCtx, readCancel := context.WithTimeout(ctx, timeout)
			value, err := client.ReadProperty(readCtx, deviceID, objectID, propID)
			readCancel()

			if err != nil {
				fmt.Fprintf(os.Stderr, "[%s] Error: %v\n", time.Now().Format("15:04:05.000"), err)
				continue
			}

			changed := !valuesEqual(lastValue, value)
			if changed || verbose {
				outputWatchValue(time.Now(), objectID, propID, value, changed)
				lastValue = value
			}
		}
	}
}

func runCOVWatch(ctx context.Context, client *bacnet.Client, objectID bacnet.ObjectIdentifier, propID bacnet.PropertyIdentifier) error {
	// Build subscription options
	var subOpts []bacnet.SubscribeOption
	if watchCOVLifetime > 0 {
		subOpts = append(subOpts, bacnet.WithSubscriptionLifetime(watchCOVLifetime))
	}

	// Subscribe to COV
	handler := func(devID uint32, oid bacnet.ObjectIdentifier, values []bacnet.PropertyValue) {
		for _, pv := range values {
			if pv.PropertyID == propID {
				outputWatchValue(time.Now(), oid, pv.PropertyID, pv.Value, true)
			}
		}
	}

	subID, err := client.SubscribeCOV(ctx, deviceID, objectID, handler, subOpts...)
	if err != nil {
		return fmt.Errorf("subscribe COV: %w", err)
	}

	fmt.Printf("Subscribed to COV (subscription ID: %d)\n", subID)

	// Wait for context cancellation
	<-ctx.Done()

	// Unsubscribe
	unsubCtx, unsubCancel := context.WithTimeout(context.Background(), timeout)
	defer unsubCancel()

	if err := client.UnsubscribeCOV(unsubCtx, deviceID, objectID, subID); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to unsubscribe: %v\n", err)
	}

	return nil
}

func outputWatchValue(t time.Time, objectID bacnet.ObjectIdentifier, propID bacnet.PropertyIdentifier, value interface{}, changed bool) {
	changeMarker := " "
	if changed {
		changeMarker = "*"
	}

	switch outputFmt {
	case "json":
		fmt.Printf(`{"time": "%s", "object": "%s", "property": "%s", "value": %s, "changed": %v}`+"\n",
			t.Format(time.RFC3339Nano),
			objectID.String(),
			propID.String(),
			formatValueJSON(value),
			changed,
		)
	case "csv":
		fmt.Printf("%s,%s,%s,%s,%v\n",
			t.Format(time.RFC3339Nano),
			objectID.String(),
			propID.String(),
			formatValue(value),
			changed,
		)
	default:
		fmt.Printf("[%s] %s %s.%s = %s\n",
			t.Format("15:04:05.000"),
			changeMarker,
			objectID.String(),
			propID.String(),
			formatValue(value),
		)
	}
}

func formatValueJSON(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return fmt.Sprintf("%q", v)
	case bacnet.ObjectIdentifier:
		return fmt.Sprintf("%q", v.String())
	default:
		return formatValue(value)
	}
}

func valuesEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}
