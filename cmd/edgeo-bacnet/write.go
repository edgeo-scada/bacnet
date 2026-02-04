package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/edgeo-scada/bacnet/bacnet"
)

var (
	writeObjectType  string
	writeProperty    string
	writeValue       string
	writePriority    int
	writeArrayIndex  int
)

var writeCmd = &cobra.Command{
	Use:   "write",
	Short: "Write a property to a BACnet object",
	Long: `Write sets property values on BACnet objects.

Value types are automatically detected:
  - Numbers: 123, 45.67, -10
  - Booleans: true, false, active, inactive
  - Strings: "text value"
  - Null: null (to release priority)

Examples:
  # Write present value to analog output
  edgeo-bacnet write -d 1234 -o analog-output:1 -p present-value -V 75.5

  # Write with priority
  edgeo-bacnet write -d 1234 -o binary-output:1 -p present-value -V true --priority 8

  # Release a priority (write null)
  edgeo-bacnet write -d 1234 -o analog-output:1 -p present-value -V null --priority 8

  # Write object name
  edgeo-bacnet write -d 1234 -o analog-value:1 -p object-name -V "Temperature Setpoint"`,

	RunE: runWrite,
}

func init() {
	writeCmd.Flags().StringVarP(&writeObjectType, "object", "O", "", "Object type and instance (e.g., analog-output:1)")
	writeCmd.Flags().StringVarP(&writeProperty, "property", "P", "present-value", "Property identifier")
	writeCmd.Flags().StringVarP(&writeValue, "value", "V", "", "Value to write")
	writeCmd.Flags().IntVar(&writePriority, "priority", 0, "Write priority (1-16, 0 for no priority)")
	writeCmd.Flags().IntVar(&writeArrayIndex, "index", -1, "Array index (-1 for no index)")

	writeCmd.MarkFlagRequired("object")
	writeCmd.MarkFlagRequired("value")
}

func runWrite(cmd *cobra.Command, args []string) error {
	if deviceID == 0 {
		return fmt.Errorf("device ID is required (-d or --device)")
	}

	// Parse object identifier
	objectID, err := parseObjectIdentifier(writeObjectType)
	if err != nil {
		return fmt.Errorf("invalid object: %w", err)
	}

	// Parse property identifier
	propID, err := parsePropertyIdentifier(writeProperty)
	if err != nil {
		return fmt.Errorf("invalid property: %w", err)
	}

	// Parse value
	value, err := parseValue(writeValue)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
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

	// Build write options
	var writeOpts []bacnet.WriteOption
	if writePriority > 0 && writePriority <= 16 {
		writeOpts = append(writeOpts, bacnet.WithPriority(uint8(writePriority)))
	}
	if writeArrayIndex >= 0 {
		writeOpts = append(writeOpts, bacnet.WithWriteArrayIndex(uint32(writeArrayIndex)))
	}

	// Write property
	if err := client.WriteProperty(ctx, deviceID, objectID, propID, value, writeOpts...); err != nil {
		return fmt.Errorf("write property: %w", err)
	}

	fmt.Printf("Successfully wrote %s to %s.%s\n", formatValue(value), objectID.String(), propID.String())
	return nil
}

func parseValue(s string) (interface{}, error) {
	s = strings.TrimSpace(s)

	// Null
	if strings.ToLower(s) == "null" {
		return nil, nil
	}

	// Boolean
	switch strings.ToLower(s) {
	case "true", "active", "on", "1":
		return true, nil
	case "false", "inactive", "off", "0":
		// Check if it's actually a number
		if s == "0" {
			return uint32(0), nil
		}
		return false, nil
	}

	// Quoted string
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return s[1 : len(s)-1], nil
	}

	// Try float
	if strings.Contains(s, ".") {
		if f, err := strconv.ParseFloat(s, 32); err == nil {
			return float32(f), nil
		}
	}

	// Try integer
	if i, err := strconv.ParseInt(s, 10, 32); err == nil {
		if i < 0 {
			return int32(i), nil
		}
		return uint32(i), nil
	}

	// Default to string
	return s, nil
}
