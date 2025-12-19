# mcs

A simple, unofficial CLI for [Mazda Connected Services](https://approval.mymazda.com/MyMazda/mazdaConnectedServices.action), ported from [pymazda](https://github.com/ungineer/pymazda).

## Install

```bash
go build -o mcs ./cmd/mcs
mv mcs /usr/local/bin/  # or anywhere in PATH
```

## Configuration

Create `~/.config/mcs/config.toml`:

```toml
email = "your@email.com"
password = "yourpassword"
region = "MNAO"  # MNAO=North America, MME=Europe, MJO=Japan
```

Or use environment variables: `MCS_EMAIL`, `MCS_PASSWORD`, `MCS_REGION`

## Usage

```bash
# Status
mcs status              # Full summary
mcs status battery      # Battery/charging only
mcs status fuel         # Fuel level
mcs status location     # GPS + Google Maps link
mcs status tires        # Tire pressures
mcs status doors        # Lock status
mcs status --json       # JSON output

# Control
mcs lock                # Lock doors
mcs unlock              # Unlock doors
mcs start               # Remote start engine
mcs stop                # Stop engine

# Charging
mcs charge start        # Start charging
mcs charge stop         # Stop charging

# Climate
mcs climate on          # Turn on HVAC
mcs climate off         # Turn off HVAC
mcs climate set --temp 21   # Set temperature (Celsius)

# Debug
mcs raw status          # Raw vehicle status JSON
mcs raw ev              # Raw EV status JSON

# Shell completions
mcs completion bash     # Also: zsh, fish, powershell
```

## Example

```
$ mcs status
CX-90 PHEV 2.5L (2025)
VIN: JM3KKDHA*********

Status as of 2025-12-19 09:17:02 (5 min ago)

BATTERY: 43%
FUEL: 92% (10 km EV + 610 km fuel = 620 km total)
CLIMATE: Off
DOORS: All locked
WINDOWS: All closed
TIRES: FL:32.0 FR:32.0 RL:32.0 RR:32.0 PSI
ODOMETER: 5,279.4 km
```

## Notes

- Uses vehicle manufacturer's API (reverse-engineered from mobile app)
- Tokens cached in `~/.cache/mcs/token.json`
- Remote start limited to 2 consecutive starts without driving
- If the manufacturer updates the app, constants in `internal/api/auth.go` may need updating
