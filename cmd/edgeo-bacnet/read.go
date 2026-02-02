package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edgeo/drivers/bacnet/bacnet"
)

var (
	readObjectType  string
	readObjectInst  uint32
	readProperty    string
	readArrayIndex  int
)

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Read a property from a BACnet object",
	Long: `Read retrieves property values from BACnet objects.

Object types can be specified by name or number:
  analog-input, ai, 0
  analog-output, ao, 1
  analog-value, av, 2
  binary-input, bi, 3
  binary-output, bo, 4
  binary-value, bv, 5
  device, dev, 8
  multi-state-input, msi, 13
  multi-state-output, mso, 14
  multi-state-value, msv, 19

Properties can be specified by name or number:
  present-value, pv, 85
  object-name, name, 77
  description, desc, 28
  status-flags, sf, 111
  units, 117
  out-of-service, oos, 81

Examples:
  # Read present value from analog input 1
  edgeo-bacnet read -d 1234 -o analog-input:1 -p present-value

  # Read using short names
  edgeo-bacnet read -d 1234 -o ai:1 -p pv

  # Read object name
  edgeo-bacnet read -d 1234 -o device:1234 -p object-name

  # Read array element
  edgeo-bacnet read -d 1234 -o device:1234 -p object-list --index 1`,

	RunE: runRead,
}

func init() {
	readCmd.Flags().StringVarP(&readObjectType, "object", "O", "", "Object type and instance (e.g., analog-input:1 or ai:1)")
	readCmd.Flags().Uint32Var(&readObjectInst, "instance", 0, "Object instance (alternative to -O)")
	readCmd.Flags().StringVarP(&readProperty, "property", "P", "present-value", "Property identifier")
	readCmd.Flags().IntVar(&readArrayIndex, "index", -1, "Array index (-1 for no index)")

	readCmd.MarkFlagRequired("object")
}

func runRead(cmd *cobra.Command, args []string) error {
	if deviceID == 0 {
		return fmt.Errorf("device ID is required (-d or --device)")
	}

	// Parse object identifier
	objectID, err := parseObjectIdentifier(readObjectType)
	if err != nil {
		return fmt.Errorf("invalid object: %w", err)
	}

	// Parse property identifier
	propID, err := parsePropertyIdentifier(readProperty)
	if err != nil {
		return fmt.Errorf("invalid property: %w", err)
	}

	client, err := createClient()
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout*2)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	// Build read options
	var readOpts []bacnet.ReadOption
	if readArrayIndex >= 0 {
		readOpts = append(readOpts, bacnet.WithArrayIndex(uint32(readArrayIndex)))
	}

	// Read property
	value, err := client.ReadProperty(ctx, deviceID, objectID, propID, readOpts...)
	if err != nil {
		return fmt.Errorf("read property: %w", err)
	}

	// Output result
	switch outputFmt {
	case "json":
		return outputValueJSON(objectID, propID, value)
	case "csv":
		return outputValueCSV(objectID, propID, value)
	case "raw":
		fmt.Println(formatValue(value))
	default:
		return outputValueTable(objectID, propID, value)
	}

	return nil
}

func parseObjectIdentifier(s string) (bacnet.ObjectIdentifier, error) {
	// Format: type:instance (e.g., analog-input:1 or ai:1 or 0:1)
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return bacnet.ObjectIdentifier{}, fmt.Errorf("expected format type:instance (e.g., analog-input:1)")
	}

	// Parse instance
	instance, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return bacnet.ObjectIdentifier{}, fmt.Errorf("invalid instance number: %s", parts[1])
	}

	// Parse type
	if typeNum, err := strconv.ParseUint(parts[0], 10, 16); err == nil {
		return bacnet.NewObjectIdentifier(bacnet.ObjectType(typeNum), uint32(instance)), nil
	}

	objType, ok := bacnet.ParseObjectType(strings.ToLower(parts[0]))
	if !ok {
		return bacnet.ObjectIdentifier{}, fmt.Errorf("unknown object type: %s", parts[0])
	}

	return bacnet.NewObjectIdentifier(objType, uint32(instance)), nil
}

func parsePropertyIdentifier(s string) (bacnet.PropertyIdentifier, error) {
	// Try parsing as number
	if propNum, err := strconv.ParseUint(s, 10, 32); err == nil {
		return bacnet.PropertyIdentifier(propNum), nil
	}

	// Parse as name
	prop, ok := bacnet.ParsePropertyIdentifier(strings.ToLower(s))
	if !ok {
		return 0, fmt.Errorf("unknown property: %s", s)
	}

	return prop, nil
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case uint32:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case float32:
		return fmt.Sprintf("%.4f", v)
	case float64:
		return fmt.Sprintf("%.6f", v)
	case string:
		return v
	case bacnet.ObjectIdentifier:
		return v.String()
	case []byte:
		return fmt.Sprintf("%x", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func outputValueTable(objectID bacnet.ObjectIdentifier, propID bacnet.PropertyIdentifier, value interface{}) error {
	fmt.Printf("Object:   %s\n", objectID.String())
	fmt.Printf("Property: %s\n", propID.String())
	fmt.Printf("Value:    %s\n", formatValue(value))
	return nil
}

func outputValueJSON(objectID bacnet.ObjectIdentifier, propID bacnet.PropertyIdentifier, value interface{}) error {
	valStr := formatValue(value)

	// Quote strings
	switch value.(type) {
	case string:
		valStr = fmt.Sprintf("%q", valStr)
	case nil:
		valStr = "null"
	case bool:
		// Already formatted
	case bacnet.ObjectIdentifier:
		valStr = fmt.Sprintf("%q", valStr)
	default:
		// Numbers don't need quotes
	}

	fmt.Printf(`{"object": "%s", "property": "%s", "value": %s}`+"\n",
		objectID.String(), propID.String(), valStr)
	return nil
}

func outputValueCSV(objectID bacnet.ObjectIdentifier, propID bacnet.PropertyIdentifier, value interface{}) error {
	fmt.Printf("%s,%s,%s\n", objectID.String(), propID.String(), formatValue(value))
	return nil
}
