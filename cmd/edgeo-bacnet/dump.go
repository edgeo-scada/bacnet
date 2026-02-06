// Copyright 2025 Edgeo SCADA
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/edgeo-scada/bacnet/bacnet"
)

var (
	dumpFile       string
	dumpProperties []string
	dumpObjects    []string
	dumpAll        bool
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all objects and properties from a device",
	Long: `Dump reads all objects and their properties from a BACnet device.

This is useful for device configuration backup, documentation, or debugging.

Examples:
  # Dump all objects to stdout
  edgeo-bacnet dump -d 1234

  # Dump to a JSON file
  edgeo-bacnet dump -d 1234 -f device_backup.json -o json

  # Dump specific object types
  edgeo-bacnet dump -d 1234 --objects analog-input,analog-output

  # Dump specific properties
  edgeo-bacnet dump -d 1234 --props present-value,object-name,description`,

	RunE: runDump,
}

func init() {
	dumpCmd.Flags().StringVarP(&dumpFile, "file", "f", "", "Output file (default: stdout)")
	dumpCmd.Flags().StringSliceVar(&dumpProperties, "props", []string{"present-value", "object-name", "description", "units", "status-flags"}, "Properties to read")
	dumpCmd.Flags().StringSliceVar(&dumpObjects, "objects", nil, "Object types to include (default: all)")
	dumpCmd.Flags().BoolVar(&dumpAll, "all", false, "Dump all properties (may be slow)")
}

type DumpObject struct {
	ObjectID   string                 `json:"object_id"`
	ObjectType string                 `json:"object_type"`
	Instance   uint32                 `json:"instance"`
	Properties map[string]interface{} `json:"properties"`
}

type DumpResult struct {
	DeviceID   uint32       `json:"device_id"`
	Timestamp  time.Time    `json:"timestamp"`
	Objects    []DumpObject `json:"objects"`
}

func runDump(cmd *cobra.Command, args []string) error {
	if deviceID == 0 {
		return fmt.Errorf("device ID is required (-d or --device)")
	}

	client, err := createClient()
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	fmt.Fprintln(os.Stderr, "Retrieving object list...")

	// Get object list
	objects, err := client.GetObjectList(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("get object list: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Found %d objects\n", len(objects))

	// Filter objects if specified
	if len(dumpObjects) > 0 {
		filtered := make([]bacnet.ObjectIdentifier, 0)
		for _, obj := range objects {
			for _, typeStr := range dumpObjects {
				objType, ok := bacnet.ParseObjectType(typeStr)
				if ok && obj.Type == objType {
					filtered = append(filtered, obj)
					break
				}
			}
		}
		objects = filtered
		fmt.Fprintf(os.Stderr, "Filtered to %d objects\n", len(objects))
	}

	// Parse properties to read
	props := make([]bacnet.PropertyIdentifier, 0, len(dumpProperties))
	if dumpAll {
		// Read common properties
		props = []bacnet.PropertyIdentifier{
			bacnet.PropertyObjectIdentifier,
			bacnet.PropertyObjectName,
			bacnet.PropertyObjectType,
			bacnet.PropertyPresentValue,
			bacnet.PropertyDescription,
			bacnet.PropertyStatusFlags,
			bacnet.PropertyEventState,
			bacnet.PropertyReliability,
			bacnet.PropertyOutOfService,
			bacnet.PropertyUnits,
			bacnet.PropertyPriorityArray,
			bacnet.PropertyRelinquishDefault,
			bacnet.PropertyCOVIncrement,
			bacnet.PropertyHighLimit,
			bacnet.PropertyLowLimit,
		}
	} else {
		for _, propStr := range dumpProperties {
			prop, ok := bacnet.ParsePropertyIdentifier(propStr)
			if ok {
				props = append(props, prop)
			}
		}
	}

	// Read all objects
	result := DumpResult{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Objects:   make([]DumpObject, 0, len(objects)),
	}

	for i, obj := range objects {
		fmt.Fprintf(os.Stderr, "\rReading object %d/%d: %s", i+1, len(objects), obj.String())

		dumpObj := DumpObject{
			ObjectID:   obj.String(),
			ObjectType: obj.Type.String(),
			Instance:   obj.Instance,
			Properties: make(map[string]interface{}),
		}

		for _, prop := range props {
			readCtx, readCancel := context.WithTimeout(ctx, timeout)
			value, err := client.ReadProperty(readCtx, deviceID, obj, prop)
			readCancel()

			if err != nil {
				continue // Skip properties that fail
			}

			dumpObj.Properties[prop.String()] = formatValueForDump(value)
		}

		result.Objects = append(result.Objects, dumpObj)
	}

	fmt.Fprintln(os.Stderr, "\nDump complete")

	// Output results
	var out *os.File
	if dumpFile != "" {
		out, err = os.Create(dumpFile)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer out.Close()
	} else {
		out = os.Stdout
	}

	switch outputFmt {
	case "json":
		return outputDumpJSON(out, result)
	case "csv":
		return outputDumpCSV(out, result)
	default:
		return outputDumpTable(out, result)
	}
}

func formatValueForDump(value interface{}) interface{} {
	switch v := value.(type) {
	case bacnet.ObjectIdentifier:
		return v.String()
	case []byte:
		return fmt.Sprintf("%x", v)
	default:
		return v
	}
}

func outputDumpJSON(out *os.File, result DumpResult) error {
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputDumpCSV(out *os.File, result DumpResult) error {
	writer := csv.NewWriter(out)
	defer writer.Flush()

	// Write header
	header := []string{"object_id", "object_type", "instance"}
	propNames := make([]string, 0)
	if len(result.Objects) > 0 {
		for prop := range result.Objects[0].Properties {
			propNames = append(propNames, prop)
			header = append(header, prop)
		}
	}
	writer.Write(header)

	// Write data
	for _, obj := range result.Objects {
		row := []string{obj.ObjectID, obj.ObjectType, fmt.Sprintf("%d", obj.Instance)}
		for _, prop := range propNames {
			val := obj.Properties[prop]
			row = append(row, fmt.Sprintf("%v", val))
		}
		writer.Write(row)
	}

	return nil
}

func outputDumpTable(out *os.File, result DumpResult) error {
	fmt.Fprintf(out, "Device %d - %d objects\n", result.DeviceID, len(result.Objects))
	fmt.Fprintf(out, "Timestamp: %s\n\n", result.Timestamp.Format(time.RFC3339))

	for _, obj := range result.Objects {
		fmt.Fprintf(out, "=== %s ===\n", obj.ObjectID)
		for prop, val := range obj.Properties {
			fmt.Fprintf(out, "  %-25s: %v\n", prop, val)
		}
		fmt.Fprintln(out)
	}

	return nil
}
