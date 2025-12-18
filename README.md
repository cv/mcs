# cx90

CLI for Mazda CX-90 PHEV via MyMazda Connected Services.

## Install

```bash
go build -o cx90 ./cmd/cx90
mv cx90 /usr/local/bin/  # or anywhere in PATH
```

## Configuration

Create `~/.config/cx90/config.toml`:

```toml
email = "your@email.com"
password = "yourpassword"
region = "MNAO"  # MNAO=North America, MME=Europe, MJO=Japan
```

Or use environment variables: `MYMAZDA_EMAIL`, `MYMAZDA_PASSWORD`, `MYMAZDA_REGION`

## Usage

```bash
# Status
cx90 status              # Full summary
cx90 status battery      # Battery/charging only
cx90 status fuel         # Fuel level
cx90 status location     # GPS + Google Maps link
cx90 status tires        # Tire pressures
cx90 status doors        # Lock status
cx90 status --json       # JSON output

# Control
cx90 lock                # Lock doors
cx90 unlock              # Unlock doors
cx90 start               # Remote start engine
cx90 stop                # Stop engine

# Charging
cx90 charge start        # Start charging
cx90 charge stop         # Stop charging

# Climate
cx90 climate on          # Turn on HVAC
cx90 climate off         # Turn off HVAC
cx90 climate set 21      # Set temperature (Celsius)

# Debug
cx90 raw status          # Raw vehicle status JSON
cx90 raw ev              # Raw EV status JSON

# Shell completions
cx90 completion bash     # Also: zsh, fish, powershell
```

## Example

```
$ cx90 status

CX-90 GT PHEV (2025)
Last Updated: 2025-12-18 05:32:39

BATTERY: 66% (630.0 km range) [plugged in, charging]
FUEL: 92% (630.0 km range)
DOORS: All locked
TIRES: FL:32.0 FR:32.0 RL:32.0 RR:32.0 PSI
```

## Notes

- Uses unofficial MyMazda API (reverse-engineered from Android app)
- Tokens cached in `~/.cache/cx90/token.json`
- Remote start limited to 2 consecutive starts without driving
- If Mazda updates the app, constants in `internal/api/auth.go` may need updating
