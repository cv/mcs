# MCS CLI Command Reference

Complete reference for all `mcs` commands and their options.

## Global Flags

These flags work with all commands:

| Flag | Description |
|------|-------------|
| `-c, --config <path>` | Config file path (default: ~/.config/mcs/config.toml) |
| `--no-color` | Disable colored output |
| `-h, --help` | Show help for any command |

## Status Commands

### `mcs status`
Show comprehensive vehicle status.

```bash
mcs status              # Full status display
mcs status --json       # JSON output
mcs status --refresh    # Request fresh data from vehicle (PHEV/EV)
mcs status -r           # Short form of --refresh
```

**Flags:**
- `--json` - Output in JSON format
- `-r, --refresh` - Request fresh status from vehicle (PHEV/EV only)
- `--refresh-wait <seconds>` - Max wait for vehicle response (default: 90)

## Climate Commands

### `mcs climate on`
Turn HVAC system on.

```bash
mcs climate on                    # Turn on with defaults
mcs climate on --confirm=false    # Don't wait for confirmation
mcs climate on --confirm-wait 60  # Wait up to 60 seconds
```

### `mcs climate off`
Turn HVAC system off.

```bash
mcs climate off
```

### `mcs climate set`
Set temperature and defroster settings.

```bash
mcs climate set --temp 22                # Set to 22°C
mcs climate set --temp 72 --unit f       # Set to 72°F
mcs climate set --temp 20 --front-defrost    # With front defroster
mcs climate set --temp 21 --rear-defrost     # With rear defroster
```

**Flags:**
- `--temp <value>` - Temperature to set (required)
- `--unit <c|f>` - Temperature unit (default: c)
- `--front-defrost` - Enable front defroster
- `--rear-defrost` - Enable rear defroster

## Lock Commands

### `mcs lock`
Lock all vehicle doors.

```bash
mcs lock                      # Lock and wait for confirmation
mcs lock --confirm=false      # Lock without waiting
```

### `mcs unlock`
Unlock all vehicle doors.

```bash
mcs unlock                    # Unlock and wait for confirmation
mcs unlock --confirm=false    # Unlock without waiting
```

## Engine Commands

### `mcs start`
Start the vehicle engine remotely.

```bash
mcs start                     # Start and wait for confirmation
mcs start --confirm=false     # Start without waiting
```

### `mcs stop`
Stop the vehicle engine.

```bash
mcs stop                      # Stop and wait for confirmation
mcs stop --confirm=false      # Stop without waiting
```

## Charging Commands

### `mcs charge start`
Start charging (EV/PHEV).

```bash
mcs charge start
```

### `mcs charge stop`
Stop charging.

```bash
mcs charge stop
```

## Confirmation Polling

All control commands support confirmation polling:

| Flag | Description |
|------|-------------|
| `--confirm` | Wait for vehicle to confirm action (default: true) |
| `--confirm=false` | Return immediately without waiting |
| `--confirm-wait <seconds>` | Custom timeout (default: 90) |

**Behavior:**
- 20 second initial delay before first poll
- 5 second intervals between polls
- Command shows success when vehicle reports new state

## Debug Commands

### `mcs raw status`
Output raw JSON response from API (for debugging).

```bash
mcs raw status
```

## Configuration

Create `~/.config/mcs/config.toml`:

```toml
email = "your.email@example.com"
password = "your-password"
region = "MNAO"  # MNAO, MME, or MJO
```

Or use environment variables:
```bash
export MCS_EMAIL="your.email@example.com"
export MCS_PASSWORD="your-password"
export MCS_REGION="MNAO"
```

## Output Examples

### Text Status Output
```
CX-90 PHEV (2024)
VIN: JM3XXXXXXXXXX1234
Status as of 2024-03-15 14:30:45 (2 min ago)

BATTERY: 85% [plugged in, not charging]
FUEL: 75% (45 km EV + 450 km fuel = 495 km total)
CLIMATE: Off, 18°C
DOORS: All locked
WINDOWS: All closed
TIRES: FL:35.0 FR:35.0 RL:33.0 RR:33.0 PSI
ODOMETER: 12,345.6 km
```

### JSON Status Output
```json
{
  "vehicle": {
    "model": "CX-90 PHEV",
    "year": 2024,
    "vin": "JM3XXXXXXXXXX1234"
  },
  "battery": {
    "level": 85,
    "range_km": 45,
    "plugged_in": true,
    "charging": false
  },
  "fuel": {
    "level": 75,
    "range_km": 450
  }
}
```
