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

// Package main demonstrates basic BACnet client usage
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/edgeo-scada/bacnet"
)

func main() {
	// Create client with options
	client, err := bacnet.NewClient(
		bacnet.WithTimeout(3*time.Second),
		bacnet.WithRetries(3),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Connect to the network
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	fmt.Println("Connected to BACnet network")

	// Discover devices
	fmt.Println("Discovering devices...")
	devices, err := client.WhoIs(ctx, bacnet.WithDiscoveryTimeout(5*time.Second))
	if err != nil {
		log.Fatalf("Discovery failed: %v", err)
	}

	fmt.Printf("Found %d device(s)\n", len(devices))
	for _, dev := range devices {
		fmt.Printf("  Device %d (Vendor: %d)\n", dev.ObjectID.Instance, dev.VendorID)
	}

	if len(devices) == 0 {
		fmt.Println("No devices found")
		return
	}

	// Use first device
	deviceID := devices[0].ObjectID.Instance

	// Read device name
	name, err := client.ReadProperty(ctx, deviceID,
		bacnet.NewObjectIdentifier(bacnet.ObjectTypeDevice, deviceID),
		bacnet.PropertyObjectName,
	)
	if err != nil {
		log.Printf("Failed to read device name: %v", err)
	} else {
		fmt.Printf("Device name: %v\n", name)
	}

	// Read analog input present value (if exists)
	value, err := client.ReadProperty(ctx, deviceID,
		bacnet.NewObjectIdentifier(bacnet.ObjectTypeAnalogInput, 1),
		bacnet.PropertyPresentValue,
	)
	if err != nil {
		log.Printf("Failed to read AI:1 present value: %v", err)
	} else {
		fmt.Printf("Analog Input 1 present value: %v\n", value)
	}

	// Print metrics
	metrics := client.Metrics().Snapshot()
	fmt.Printf("\nMetrics:\n")
	fmt.Printf("  Requests sent: %d\n", metrics.RequestsSent)
	fmt.Printf("  Requests succeeded: %d\n", metrics.RequestsSucceeded)
	fmt.Printf("  Avg latency: %v\n", metrics.LatencyStats.Avg)
}
