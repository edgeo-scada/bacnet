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
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/edgeo-scada/bacnet"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "Start an interactive BACnet session",
	Long: `Interactive mode provides a REPL for exploring BACnet devices.

Commands:
  scan                                  - Discover devices
  use <device-id>                       - Select a device
  list                                  - List objects on current device
  read <object> <property>              - Read a property
  write <object> <property> <value>     - Write a property
  info                                  - Show device info
  metrics                               - Show client metrics
  help                                  - Show help
  exit                                  - Exit interactive mode

Examples:
  bacnet> scan
  bacnet> use 1234
  bacnet[1234]> list
  bacnet[1234]> read ai:1 pv
  bacnet[1234]> write ao:1 pv 75.5`,

	RunE: runInteractive,
}

func runInteractive(cmd *cobra.Command, args []string) error {
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	fmt.Println("BACnet Interactive Shell")
	fmt.Println("Type 'help' for available commands, 'exit' to quit")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	currentDevice := uint32(0)

	for {
		// Print prompt
		if currentDevice > 0 {
			fmt.Printf("bacnet[%d]> ", currentDevice)
		} else {
			fmt.Print("bacnet> ")
		}

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		command := strings.ToLower(parts[0])

		switch command {
		case "exit", "quit", "q":
			fmt.Println("Goodbye!")
			return nil

		case "help", "?":
			printInteractiveHelp()

		case "scan":
			runInteractiveScan(ctx, client)

		case "use":
			if len(parts) < 2 {
				fmt.Println("Usage: use <device-id>")
				continue
			}
			var newDevice uint32
			fmt.Sscanf(parts[1], "%d", &newDevice)
			if newDevice > 0 {
				currentDevice = newDevice
				fmt.Printf("Selected device %d\n", currentDevice)
			} else {
				fmt.Println("Invalid device ID")
			}

		case "list":
			if currentDevice == 0 {
				fmt.Println("No device selected. Use 'use <device-id>' first.")
				continue
			}
			runInteractiveList(ctx, client, currentDevice)

		case "read":
			if currentDevice == 0 {
				fmt.Println("No device selected. Use 'use <device-id>' first.")
				continue
			}
			if len(parts) < 2 {
				fmt.Println("Usage: read <object> [property]")
				continue
			}
			prop := "present-value"
			if len(parts) >= 3 {
				prop = parts[2]
			}
			runInteractiveRead(ctx, client, currentDevice, parts[1], prop)

		case "write":
			if currentDevice == 0 {
				fmt.Println("No device selected. Use 'use <device-id>' first.")
				continue
			}
			if len(parts) < 4 {
				fmt.Println("Usage: write <object> <property> <value>")
				continue
			}
			runInteractiveWrite(ctx, client, currentDevice, parts[1], parts[2], strings.Join(parts[3:], " "))

		case "info":
			if currentDevice == 0 {
				fmt.Println("No device selected. Use 'use <device-id>' first.")
				continue
			}
			runInteractiveInfo(ctx, client, currentDevice)

		case "metrics":
			runInteractiveMetrics(client)

		default:
			fmt.Printf("Unknown command: %s (type 'help' for available commands)\n", command)
		}
	}

	return nil
}

func printInteractiveHelp() {
	fmt.Println(`
Available commands:
  scan                              Discover BACnet devices on the network
  use <device-id>                   Select a device to work with
  list                              List all objects on current device
  read <object> [property]          Read a property (default: present-value)
  write <object> <property> <value> Write a property value
  info                              Show current device information
  metrics                           Show client metrics
  help                              Show this help message
  exit                              Exit interactive mode

Object format: <type>:<instance>
  Examples: analog-input:1, ai:1, binary-output:5, device:1234

Property shortcuts:
  pv = present-value
  name = object-name
  desc = description
  sf = status-flags
  oos = out-of-service
`)
}

func runInteractiveScan(ctx context.Context, client *bacnet.Client) {
	fmt.Println("Scanning for devices...")

	scanCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	devices, err := client.WhoIs(scanCtx, bacnet.WithDiscoveryTimeout(3*time.Second))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return
	}

	fmt.Printf("\nFound %d device(s):\n", len(devices))
	for _, dev := range devices {
		fmt.Printf("  Device %d - %s (Vendor: %d)\n",
			dev.ObjectID.Instance,
			formatAddress(dev.Address),
			dev.VendorID,
		)
	}
	fmt.Println()
}

func runInteractiveList(ctx context.Context, client *bacnet.Client, devID uint32) {
	listCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	objects, err := client.GetObjectList(listCtx, devID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("\nDevice %d has %d objects:\n", devID, len(objects))

	// Group by type
	byType := make(map[bacnet.ObjectType][]bacnet.ObjectIdentifier)
	for _, obj := range objects {
		byType[obj.Type] = append(byType[obj.Type], obj)
	}

	for objType, objs := range byType {
		fmt.Printf("\n  %s (%d):\n", objType.String(), len(objs))
		for _, obj := range objs {
			fmt.Printf("    %d\n", obj.Instance)
		}
	}
	fmt.Println()
}

func runInteractiveRead(ctx context.Context, client *bacnet.Client, devID uint32, objStr, propStr string) {
	objectID, err := parseObjectIdentifier(objStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	propID, err := parsePropertyIdentifier(propStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	readCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	value, err := client.ReadProperty(readCtx, devID, objectID, propID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("%s.%s = %s\n", objectID.String(), propID.String(), formatValue(value))
}

func runInteractiveWrite(ctx context.Context, client *bacnet.Client, devID uint32, objStr, propStr, valStr string) {
	objectID, err := parseObjectIdentifier(objStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	propID, err := parsePropertyIdentifier(propStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	value, err := parseValue(valStr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	writeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := client.WriteProperty(writeCtx, devID, objectID, propID, value); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("OK: %s.%s = %s\n", objectID.String(), propID.String(), formatValue(value))
}

func runInteractiveInfo(ctx context.Context, client *bacnet.Client, devID uint32) {
	deviceOID := bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, devID)

	props := []struct {
		name string
		prop bacnet.PropertyIdentifier
	}{
		{"Name", bacnet.PropertyObjectName},
		{"Vendor", bacnet.PropertyVendorName},
		{"Model", bacnet.PropertyModelName},
		{"Firmware", bacnet.PropertyFirmwareRevision},
	}

	fmt.Printf("\nDevice %d:\n", devID)
	for _, p := range props {
		readCtx, cancel := context.WithTimeout(ctx, timeout)
		val, err := client.ReadProperty(readCtx, devID, deviceOID, p.prop)
		cancel()

		if err == nil {
			fmt.Printf("  %-10s: %s\n", p.name, formatValue(val))
		}
	}
	fmt.Println()
}

func runInteractiveMetrics(client *bacnet.Client) {
	m := client.Metrics().Snapshot()

	fmt.Println("\nClient Metrics:")
	fmt.Printf("  Uptime:              %s\n", m.Uptime.Round(time.Second))
	fmt.Printf("  Requests Sent:       %d\n", m.RequestsSent)
	fmt.Printf("  Requests Succeeded:  %d\n", m.RequestsSucceeded)
	fmt.Printf("  Requests Failed:     %d\n", m.RequestsFailed)
	fmt.Printf("  Requests Timed Out:  %d\n", m.RequestsTimedOut)
	fmt.Printf("  Devices Discovered:  %d\n", m.DevicesDiscovered)
	fmt.Printf("  Bytes Sent:          %d\n", m.BytesSent)
	fmt.Printf("  Bytes Received:      %d\n", m.BytesReceived)

	if m.LatencyStats.Count > 0 {
		fmt.Printf("  Avg Latency:         %s\n", m.LatencyStats.Avg.Round(time.Microsecond))
		fmt.Printf("  Min Latency:         %s\n", m.LatencyStats.Min.Round(time.Microsecond))
		fmt.Printf("  Max Latency:         %s\n", m.LatencyStats.Max.Round(time.Microsecond))
	}
	fmt.Println()
}
