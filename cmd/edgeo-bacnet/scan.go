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
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/edgeo-scada/bacnet/bacnet"
)

var (
	scanTimeout  time.Duration
	scanLowLimit uint32
	scanHighLimit uint32
	scanNetwork  uint16
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan for BACnet devices on the network",
	Long: `Scan discovers BACnet devices by sending Who-Is broadcast requests.

Examples:
  # Discover all devices
  edgeo-bacnet scan

  # Discover devices with instance IDs 1-100
  edgeo-bacnet scan --low 1 --high 100

  # Discover with extended timeout
  edgeo-bacnet scan --scan-timeout 10s`,

	RunE: runScan,
}

func init() {
	scanCmd.Flags().DurationVar(&scanTimeout, "scan-timeout", 5*time.Second, "Discovery timeout")
	scanCmd.Flags().Uint32Var(&scanLowLimit, "low", 0, "Low limit for device instance range (0 = no limit)")
	scanCmd.Flags().Uint32Var(&scanHighLimit, "high", 0, "High limit for device instance range (0 = no limit)")
	scanCmd.Flags().Uint16Var(&scanNetwork, "network", 0, "Target network number (0 = local)")
}

func runScan(cmd *cobra.Command, args []string) error {
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout+scanTimeout)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Close()

	fmt.Fprintln(os.Stderr, "Scanning for BACnet devices...")

	// Build discovery options
	discoverOpts := []bacnet.DiscoverOption{
		bacnet.WithDiscoveryTimeout(scanTimeout),
	}

	if scanLowLimit > 0 || scanHighLimit > 0 {
		low := scanLowLimit
		high := scanHighLimit
		if high == 0 {
			high = 0x3FFFFF // Max device instance
		}
		discoverOpts = append(discoverOpts, bacnet.WithDeviceRange(low, high))
	}

	if scanNetwork > 0 {
		discoverOpts = append(discoverOpts, bacnet.WithTargetNetwork(scanNetwork))
	}

	devices, err := client.WhoIs(ctx, discoverOpts...)
	if err != nil {
		return fmt.Errorf("discovery: %w", err)
	}

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return nil
	}

	// Output results
	switch outputFmt {
	case "json":
		return outputDevicesJSON(devices)
	case "csv":
		return outputDevicesCSV(devices)
	default:
		return outputDevicesTable(devices)
	}
}

func outputDevicesTable(devices []*bacnet.DeviceInfo) error {
	fmt.Printf("\n%-12s %-20s %-8s %-20s %-10s\n", "DEVICE ID", "ADDRESS", "VENDOR", "SEGMENTATION", "MAX APDU")
	fmt.Println("------------ -------------------- -------- -------------------- ----------")

	for _, dev := range devices {
		addr := formatAddress(dev.Address)
		fmt.Printf("%-12d %-20s %-8d %-20s %-10d\n",
			dev.ObjectID.Instance,
			addr,
			dev.VendorID,
			dev.Segmentation.String(),
			dev.MaxAPDULength,
		)
	}

	fmt.Printf("\nFound %d device(s)\n", len(devices))
	return nil
}

func outputDevicesJSON(devices []*bacnet.DeviceInfo) error {
	fmt.Println("[")
	for i, dev := range devices {
		comma := ","
		if i == len(devices)-1 {
			comma = ""
		}
		fmt.Printf(`  {"device_id": %d, "address": "%s", "vendor_id": %d, "segmentation": "%s", "max_apdu": %d}%s`+"\n",
			dev.ObjectID.Instance,
			formatAddress(dev.Address),
			dev.VendorID,
			dev.Segmentation.String(),
			dev.MaxAPDULength,
			comma,
		)
	}
	fmt.Println("]")
	return nil
}

func outputDevicesCSV(devices []*bacnet.DeviceInfo) error {
	fmt.Println("device_id,address,vendor_id,segmentation,max_apdu")
	for _, dev := range devices {
		fmt.Printf("%d,%s,%d,%s,%d\n",
			dev.ObjectID.Instance,
			formatAddress(dev.Address),
			dev.VendorID,
			dev.Segmentation.String(),
			dev.MaxAPDULength,
		)
	}
	return nil
}

func formatAddress(addr bacnet.Address) string {
	if len(addr.Addr) == 4 {
		return fmt.Sprintf("%d.%d.%d.%d", addr.Addr[0], addr.Addr[1], addr.Addr[2], addr.Addr[3])
	} else if len(addr.Addr) == 6 {
		port := int(addr.Addr[4])<<8 | int(addr.Addr[5])
		return fmt.Sprintf("%d.%d.%d.%d:%d", addr.Addr[0], addr.Addr[1], addr.Addr[2], addr.Addr[3], port)
	}
	return fmt.Sprintf("%x", addr.Addr)
}
