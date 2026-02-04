package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/edgeo-scada/bacnet/bacnet"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display device information",
	Long: `Info retrieves and displays detailed information about a BACnet device.

Examples:
  # Get device info
  edgeo-bacnet info -d 1234

  # Get info in JSON format
  edgeo-bacnet info -d 1234 -o json`,

	RunE: runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	if deviceID == 0 {
		return fmt.Errorf("device ID is required (-d or --device)")
	}

	client, err := createClient()
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout*10)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	// Read device properties
	deviceOID := bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID)

	info := make(map[string]interface{})

	// Properties to read
	properties := []struct {
		name string
		prop bacnet.PropertyIdentifier
	}{
		{"Object Name", bacnet.PropertyObjectName},
		{"Vendor Name", bacnet.PropertyVendorName},
		{"Vendor ID", bacnet.PropertyVendorIdentifier},
		{"Model Name", bacnet.PropertyModelName},
		{"Firmware Revision", bacnet.PropertyFirmwareRevision},
		{"Application Software", bacnet.PropertyApplicationSoftwareVersion},
		{"Protocol Version", bacnet.PropertyProtocolVersion},
		{"Protocol Revision", bacnet.PropertyProtocolRevision},
		{"System Status", bacnet.PropertySystemStatus},
		{"Description", bacnet.PropertyDescription},
		{"Location", bacnet.PropertyLocation},
		{"Max APDU Length", bacnet.PropertyMaxApduLengthAccepted},
		{"Segmentation", bacnet.PropertySegmentationSupported},
		{"Database Revision", bacnet.PropertyDatabaseRevision},
	}

	for _, p := range properties {
		readCtx, readCancel := context.WithTimeout(ctx, timeout)
		val, err := client.ReadProperty(readCtx, deviceID, deviceOID, p.prop)
		readCancel()

		if err == nil {
			info[p.name] = val
		}
	}

	// Get object count
	readCtx, readCancel := context.WithTimeout(ctx, timeout)
	objCount, err := client.ReadProperty(readCtx, deviceID, deviceOID, bacnet.PropertyObjectList, bacnet.WithArrayIndex(0))
	readCancel()
	if err == nil {
		info["Object Count"] = objCount
	}

	// Output results
	switch outputFmt {
	case "json":
		return outputInfoJSON(info)
	default:
		return outputInfoTable(info)
	}
}

func outputInfoTable(info map[string]interface{}) error {
	fmt.Printf("\n=== Device %d ===\n\n", deviceID)

	// Ordered output
	order := []string{
		"Object Name",
		"Description",
		"Location",
		"Vendor Name",
		"Vendor ID",
		"Model Name",
		"Firmware Revision",
		"Application Software",
		"Protocol Version",
		"Protocol Revision",
		"System Status",
		"Max APDU Length",
		"Segmentation",
		"Object Count",
		"Database Revision",
	}

	for _, key := range order {
		if val, ok := info[key]; ok {
			fmt.Printf("%-25s: %v\n", key, formatValue(val))
		}
	}

	fmt.Println()
	return nil
}

func outputInfoJSON(info map[string]interface{}) error {
	fmt.Println("{")
	fmt.Printf(`  "device_id": %d,`+"\n", deviceID)
	fmt.Printf(`  "timestamp": "%s",`+"\n", time.Now().Format(time.RFC3339))

	first := true
	for key, val := range info {
		if !first {
			fmt.Println(",")
		}
		first = false

		switch v := val.(type) {
		case string:
			fmt.Printf(`  "%s": "%s"`, key, v)
		case bacnet.ObjectIdentifier:
			fmt.Printf(`  "%s": "%s"`, key, v.String())
		default:
			fmt.Printf(`  "%s": %v`, key, v)
		}
	}
	fmt.Println("\n}")
	return nil
}
