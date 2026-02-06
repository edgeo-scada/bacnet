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
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/edgeo-scada/bacnet"
)

var (
	cfgFile      string
	host         string
	port         int
	deviceID     uint32
	timeout      time.Duration
	retries      int
	outputFmt    string
	verbose      bool
	localAddress string
	bbmdAddress  string
	bbmdPort     int
	bbmdTTL      time.Duration

	client *bacnet.Client
	logger *slog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "edgeo-bacnet",
	Short: "A comprehensive BACnet/IP client CLI",
	Long: `edgeo-bacnet is a command-line tool for communicating with BACnet/IP devices.

It supports device discovery, property read/write operations, COV subscriptions,
and various diagnostic functions for building automation systems.

Examples:
  # Discover devices on the network
  edgeo-bacnet scan

  # Read a property from a device
  edgeo-bacnet read -d 1234 -o analog-input:1 -p present-value

  # Write a value to a device
  edgeo-bacnet write -d 1234 -o analog-output:1 -p present-value -v 75.5

  # Watch for value changes
  edgeo-bacnet watch -d 1234 -o analog-input:1`,

	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Setup logger
		logLevel := slog.LevelInfo
		if verbose {
			logLevel = slog.LevelDebug
		}
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))

		return nil
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.edgeo-bacnet.yaml)")
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "Target device IP address")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", bacnet.DefaultPort, "BACnet/IP port")
	rootCmd.PersistentFlags().Uint32VarP(&deviceID, "device", "d", 0, "Target device instance ID")
	rootCmd.PersistentFlags().DurationVarP(&timeout, "timeout", "t", 3*time.Second, "Request timeout")
	rootCmd.PersistentFlags().IntVar(&retries, "retries", 3, "Number of retries")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "table", "Output format (table, json, csv, raw)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&localAddress, "local", "", "Local address to bind to (e.g., 0.0.0.0:47808)")
	rootCmd.PersistentFlags().StringVar(&bbmdAddress, "bbmd", "", "BBMD address for foreign device registration")
	rootCmd.PersistentFlags().IntVar(&bbmdPort, "bbmd-port", bacnet.DefaultPort, "BBMD port")
	rootCmd.PersistentFlags().DurationVar(&bbmdTTL, "bbmd-ttl", 60*time.Second, "BBMD registration TTL")

	// Bind flags to viper
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("device", rootCmd.PersistentFlags().Lookup("device"))
	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("retries", rootCmd.PersistentFlags().Lookup("retries"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("local", rootCmd.PersistentFlags().Lookup("local"))
	viper.BindPFlag("bbmd", rootCmd.PersistentFlags().Lookup("bbmd"))
	viper.BindPFlag("bbmd-port", rootCmd.PersistentFlags().Lookup("bbmd-port"))
	viper.BindPFlag("bbmd-ttl", rootCmd.PersistentFlags().Lookup("bbmd-ttl"))

	// Add subcommands
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(dumpCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(interactiveCmd)
	rootCmd.AddCommand(versionCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".edgeo-bacnet")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("BACNET")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}

// createClient creates a BACnet client with current configuration
func createClient() (*bacnet.Client, error) {
	opts := []bacnet.Option{
		bacnet.WithTimeout(timeout),
		bacnet.WithRetries(retries),
		bacnet.WithLogger(logger),
	}

	if localAddress != "" {
		opts = append(opts, bacnet.WithLocalAddress(localAddress))
	}

	if bbmdAddress != "" {
		opts = append(opts, bacnet.WithBBMD(bbmdAddress, bbmdPort, bbmdTTL))
	}

	return bacnet.NewClient(opts...)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("edgeo-bacnet version 1.0.0")
	},
}
