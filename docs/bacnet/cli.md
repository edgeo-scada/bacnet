# CLI Reference

The `edgeo-bacnet` command-line tool provides comprehensive BACnet/IP device interaction.

## Installation

```bash
go install github.com/edgeo-scada/bacnet/cmd/edgeo-bacnet@latest
```

## Global Flags

These flags apply to all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | | `~/.edgeo-bacnet.yaml` | Config file path |
| `--host` | `-H` | | Target device IP address |
| `--port` | `-p` | `47808` | BACnet/IP port |
| `--device` | `-d` | | Target device instance ID |
| `--timeout` | `-t` | `3s` | Request timeout |
| `--retries` | | `3` | Number of retries |
| `--output` | `-o` | `table` | Output format (table, json, csv, raw) |
| `--verbose` | `-v` | `false` | Enable verbose output |
| `--local` | | | Local address to bind to |
| `--bbmd` | | | BBMD address for foreign device |
| `--bbmd-port` | | `47808` | BBMD port |
| `--bbmd-ttl` | | `60s` | BBMD registration TTL |

## Commands

### scan

Discover devices on the BACnet network.

```bash
edgeo-bacnet scan [flags]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--low` | `-l` | | Low device ID limit |
| `--high` | `-h` | | High device ID limit |
| `--timeout` | `-t` | `5s` | Discovery timeout |

**Examples:**

```bash
# Discover all devices
edgeo-bacnet scan

# Discover devices in range
edgeo-bacnet scan --low 1000 --high 2000

# Longer timeout for slow networks
edgeo-bacnet scan -t 10s

# JSON output
edgeo-bacnet scan -o json
```

**Output:**

```
Discovering BACnet devices...

DEVICE ID    ADDRESS            VENDOR ID    MAX APDU
1234         192.168.1.100      123          1476
5678         192.168.1.101      456          480

Found 2 device(s)
```

### read

Read a property from a BACnet object.

```bash
edgeo-bacnet read [flags]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--device` | `-d` | | Device instance ID (required) |
| `--object` | `-o` | | Object identifier (required) |
| `--property` | `-p` | `present-value` | Property identifier |
| `--index` | `-i` | | Array index |

**Examples:**

```bash
# Read present value from analog input
edgeo-bacnet read -d 1234 -o analog-input:1 -p present-value

# Using aliases
edgeo-bacnet read -d 1234 -o ai:1 -p pv

# Read object name
edgeo-bacnet read -d 1234 -o ai:1 -p name

# Read array element
edgeo-bacnet read -d 1234 -o device:1234 -p object-list -i 1

# Read array length
edgeo-bacnet read -d 1234 -o device:1234 -p object-list -i 0

# JSON output
edgeo-bacnet read -d 1234 -o ai:1 -p pv -o json
```

**Output:**

```
Object:   analog-input:1
Property: present-value
Value:    72.5
```

### write

Write a property value to a BACnet object.

```bash
edgeo-bacnet write [flags]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--device` | `-d` | | Device instance ID (required) |
| `--object` | `-o` | | Object identifier (required) |
| `--property` | `-p` | `present-value` | Property identifier |
| `--value` | `-v` | | Value to write (required) |
| `--priority` | `-P` | | Write priority (1-16) |
| `--index` | `-i` | | Array index |
| `--type` | | | Value type (auto, real, int, bool, string) |

**Examples:**

```bash
# Write a setpoint
edgeo-bacnet write -d 1234 -o ao:1 -p pv -v 72.5

# Write with priority
edgeo-bacnet write -d 1234 -o bo:1 -p pv -v true -P 8

# Write boolean
edgeo-bacnet write -d 1234 -o bo:1 -p pv -v true

# Write integer
edgeo-bacnet write -d 1234 -o msv:1 -p pv -v 3 --type int

# Write string
edgeo-bacnet write -d 1234 -o dev:1234 -p description -v "Main HVAC"
```

**Output:**

```
Object:   analog-output:1
Property: present-value
Value:    72.5
Priority: 8
Status:   Success
```

### watch

Monitor an object for value changes using COV subscriptions.

```bash
edgeo-bacnet watch [flags]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--device` | `-d` | | Device instance ID (required) |
| `--object` | `-o` | | Object identifier(s), comma-separated |
| `--duration` | | | Watch duration (e.g., 5m, 1h) |
| `--confirmed` | | `false` | Use confirmed notifications |
| `--increment` | | | COV increment for analog values |

**Examples:**

```bash
# Watch single object
edgeo-bacnet watch -d 1234 -o ai:1

# Watch multiple objects
edgeo-bacnet watch -d 1234 -o ai:1,ai:2,bi:1

# Watch for 5 minutes
edgeo-bacnet watch -d 1234 -o ai:1 --duration 5m

# With COV increment
edgeo-bacnet watch -d 1234 -o ai:1 --increment 0.5
```

**Output:**

```
Watching analog-input:1 on device 1234...

[2024-02-01 10:15:30] analog-input:1
  present-value: 72.5
  status-flags: {in-alarm:false, fault:false, overridden:false, out-of-service:false}

[2024-02-01 10:15:45] analog-input:1
  present-value: 73.2
  status-flags: {in-alarm:false, fault:false, overridden:false, out-of-service:false}

^C
Received 2 notifications
```

### dump

Dump all objects and their values from a device.

```bash
edgeo-bacnet dump [flags]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--device` | `-d` | | Device instance ID (required) |
| `--type` | | | Filter by object type |
| `--properties` | | `pv,name` | Properties to read |

**Examples:**

```bash
# Dump all objects
edgeo-bacnet dump -d 1234

# Dump only analog inputs
edgeo-bacnet dump -d 1234 --type analog-input

# Dump with specific properties
edgeo-bacnet dump -d 1234 --properties pv,name,units,desc

# JSON output
edgeo-bacnet dump -d 1234 -o json
```

**Output:**

```
Device 1234 Object Dump
========================

analog-input:1
  object-name:    Zone Temperature
  present-value:  72.5
  units:          °F

analog-input:2
  object-name:    Supply Air Temp
  present-value:  55.3
  units:          °F

binary-output:1
  object-name:    Fan Enable
  present-value:  active

Total: 15 objects
```

### info

Display detailed device information.

```bash
edgeo-bacnet info [flags]
```

**Flags:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--device` | `-d` | | Device instance ID (required) |

**Examples:**

```bash
# Get device info
edgeo-bacnet info -d 1234

# JSON output
edgeo-bacnet info -d 1234 -o json
```

**Output:**

```
Device Information
==================

Device ID:       1234
Vendor ID:       123
Vendor Name:     ACME Controls
Model Name:      VAV-2000
Firmware:        v2.3.1
Application:     BACnet Controller
Description:     Main Building VAV Controller
Location:        Floor 2, Mechanical Room

Protocol:        BACnet/IP
Max APDU:        1476
Segmentation:    no-segmentation

Objects:         45
  analog-input:     12
  analog-output:    6
  analog-value:     8
  binary-input:     5
  binary-output:    4
  schedule:         3
  trend-log:        7
```

### interactive

Start an interactive BACnet shell.

```bash
edgeo-bacnet interactive [flags]
```

**Commands in interactive mode:**

```
> help                          Show available commands
> scan                          Discover devices
> read <obj> <prop>             Read property
> write <obj> <prop> <value>    Write property
> watch <obj>                   Start watching object
> info                          Show device info
> device <id>                   Switch target device
> quit                          Exit
```

**Example session:**

```bash
$ edgeo-bacnet interactive -d 1234
edgeo-bacnet> read ai:1 pv
72.5
edgeo-bacnet> read ai:1 name
Zone Temperature
edgeo-bacnet> write ao:1 pv 75
OK
edgeo-bacnet> quit
```

### version

Display version information.

```bash
edgeo-bacnet version
```

**Output:**

```
edgeo-bacnet version 1.0.0
```

## Object Type Aliases

Use shorthand aliases for object types:

| Alias | Full Name |
|-------|-----------|
| ai | analog-input |
| ao | analog-output |
| av | analog-value |
| bi | binary-input |
| bo | binary-output |
| bv | binary-value |
| msi | multi-state-input |
| mso | multi-state-output |
| msv | multi-state-value |
| dev | device |
| sch | schedule |
| cal | calendar |
| tl | trend-log |
| nc | notification-class |
| prg | program |

## Property Aliases

Use shorthand aliases for properties:

| Alias | Full Name |
|-------|-----------|
| pv | present-value |
| name | object-name |
| oid | object-identifier |
| type | object-type |
| desc | description |
| sf | status-flags |
| oos | out-of-service |
| pa | priority-array |
| rd | relinquish-default |

## Output Formats

### Table (default)

Human-readable tabular format.

```bash
edgeo-bacnet scan -o table
```

### JSON

Machine-readable JSON format.

```bash
edgeo-bacnet scan -o json
```

```json
{
  "devices": [
    {
      "device_id": 1234,
      "address": "192.168.1.100:47808",
      "vendor_id": 123,
      "max_apdu": 1476,
      "segmentation": "no-segmentation"
    }
  ]
}
```

### CSV

Comma-separated values for spreadsheet import.

```bash
edgeo-bacnet scan -o csv
```

```
device_id,address,vendor_id,max_apdu,segmentation
1234,192.168.1.100:47808,123,1476,no-segmentation
```

### Raw

Minimal output, one value per line.

```bash
edgeo-bacnet read -d 1234 -o ai:1 -p pv -o raw
```

```
72.5
```

## Configuration File

Create `~/.edgeo-bacnet.yaml` for default settings:

```yaml
# Default device
device: 1234

# Network settings
timeout: 5s
retries: 3

# BBMD configuration (for cross-subnet)
bbmd: 192.168.1.1
bbmd-port: 47808
bbmd-ttl: 60s

# Output preferences
output: table
verbose: false
```

## Environment Variables

Settings can also be configured via environment variables with `BACNET_` prefix:

```bash
export BACNET_DEVICE=1234
export BACNET_TIMEOUT=5s
export BACNET_VERBOSE=true

edgeo-bacnet read -o ai:1 -p pv
```

## Examples

### Monitor Building HVAC

```bash
# Discover controllers
edgeo-bacnet scan

# Get AHU info
edgeo-bacnet info -d 1234

# Read zone temperatures
for i in 1 2 3 4 5; do
  echo "Zone $i: $(edgeo-bacnet read -d 1234 -o ai:$i -p pv -o raw)"
done

# Watch for alarms
edgeo-bacnet watch -d 1234 -o ai:1,ai:2,ai:3 --duration 1h
```

### Commissioning Script

```bash
#!/bin/bash

DEVICE=$1

# Verify device
edgeo-bacnet info -d $DEVICE || exit 1

# Test all analog outputs
for i in $(seq 1 10); do
  echo "Testing AO:$i"
  edgeo-bacnet write -d $DEVICE -o ao:$i -p pv -v 50 -P 8
  sleep 2
  edgeo-bacnet write -d $DEVICE -o ao:$i -p pv -v null -P 8
done

echo "Commissioning complete"
```

### Export Device Points

```bash
# Export all points to CSV
edgeo-bacnet dump -d 1234 --properties name,pv,units -o csv > device_1234_points.csv
```
