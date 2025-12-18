# cx90 - CLI for Mazda CX-90 PHEV

A command-line interface for interacting with Mazda Connected Services, specifically designed for the CX-90 PHEV.

## Usage

```bash
# Status commands
cx90 status              # Full vehicle status summary
cx90 status battery      # Battery/charging info only
cx90 status fuel         # Fuel level only
cx90 status location     # GPS coordinates
cx90 status tires        # Tire pressures
cx90 status doors        # Door/window/lock status

# Control commands
cx90 lock                # Lock all doors
cx90 unlock              # Unlock all doors
cx90 start               # Remote start
cx90 stop                # Stop engine
cx90 charge start        # Start charging
cx90 charge stop         # Stop charging
cx90 hazards on          # Turn on hazard lights
cx90 hazards off         # Turn off hazard lights

# Climate commands
cx90 climate on          # Turn on HVAC
cx90 climate off         # Turn off HVAC
cx90 climate set 21      # Set temperature (uses config unit)
cx90 climate defrost     # Front + rear defrost on

# Utility commands
cx90 refresh             # Force status refresh from vehicle
cx90 vehicles            # List all registered vehicles
cx90 raw status          # Raw JSON vehicle status
cx90 raw ev              # Raw JSON EV status
```

## Configuration

Config file: `~/.config/cx90/config.toml`

```toml
[auth]
email = "your@email.com"
password = "yourpassword"
region = "MNAO"  # MNAO=North America, MME=Europe, MJO=Japan

[preferences]
vehicle_id = 202379672      # Default vehicle (optional, uses first if omitted)
temperature_unit = "C"      # C or F
distance_unit = "km"        # km or mi
```

Environment variables override config:
- `MYMAZDA_EMAIL`
- `MYMAZDA_PASSWORD`
- `MYMAZDA_REGION`

## Output Formats

```bash
cx90 status                    # Human-readable (default)
cx90 status --json             # JSON output
cx90 status --json | jq .battery.percent  # Pipe to jq
```

## Implementation Plan

### Phase 1: Project Setup

1. Initialize Go module
   ```
   cx90/
   ├── cmd/
   │   └── cx90/
   │       └── main.go          # Entry point
   ├── internal/
   │   ├── api/
   │   │   ├── client.go        # API client
   │   │   ├── auth.go          # Authentication (encryption, tokens)
   │   │   ├── crypto.go        # AES/RSA encryption utils
   │   │   └── endpoints.go     # API endpoint definitions
   │   ├── config/
   │   │   └── config.go        # Config file + env loading
   │   └── cli/
   │       ├── root.go          # Root command
   │       ├── status.go        # Status subcommands
   │       ├── control.go       # Lock/unlock/start/stop
   │       ├── charge.go        # Charging commands
   │       ├── climate.go       # HVAC commands
   │       └── output.go        # Human/JSON formatters
   ├── go.mod
   ├── go.sum
   └── README.md
   ```

2. Dependencies
   - `github.com/spf13/cobra` - CLI framework
   - `github.com/spf13/viper` - Config management
   - `github.com/fatih/color` - Colored output (optional)

### Phase 2: API Client (Port from pymazda)

Key implementation details from pymazda that must be ported:

1. **Region Configuration**
   ```go
   var regions = map[string]RegionConfig{
       "MNAO": {AppCode: "202007270941270111799", BaseURL: "https://0cxo7m58.mazda.com/prod/", UsherURL: "https://ptznwbh8.mazda.com/appapi/v1/"},
       "MME":  {AppCode: "202008100250281064816", BaseURL: "https://e9stj7g7.mazda.com/prod/", UsherURL: "https://rz97suam.mazda.com/appapi/v1/"},
       "MJO":  {AppCode: "202009170613074283422", BaseURL: "https://wcs9p6wj.mazda.com/prod/", UsherURL: "https://c5ulfwxr.mazda.com/appapi/v1/"},
   }
   ```

2. **App Version** (must match current MyMazda app)
   ```go
   const (
       AppVersion     = "9.0.5"
       UserAgentBase  = "MyMazda-Android/9.0.5"
       UserAgentUsher = "MyMazda/9.0.5 (Google Pixel 3a; Android 11)"
       UsherSDKVer    = "11.3.0700.001"
       AppPackageID   = "com.interrait.mymazda"
       SignatureMD5   = "C383D8C4D279B78130AD52DC71D95CAA"
   )
   ```

3. **Encryption Flow**
   - AES-128-CBC for payload encryption (IV: "0102030405060708")
   - RSA-ECB-PKCS1 for password encryption
   - MD5/SHA256 for signing requests
   - Device ID generated from email hash

4. **Authentication Flow**
   ```
   1. POST /service/checkVersion → get encKey, signKey
   2. GET  usher/system/encryptionKey → get RSA public key
   3. POST usher/user/login → get accessToken
   4. Use accessToken for all subsequent requests
   ```

5. **Key API Endpoints**
   | Endpoint | Method | Description |
   |----------|--------|-------------|
   | `service/checkVersion` | POST | Get encryption keys |
   | `remoteServices/vehicle/status/get` | POST | Vehicle status |
   | `remoteServices/ev/status/get` | POST | EV/PHEV status |
   | `remoteServices/door/lock` | POST | Lock doors |
   | `remoteServices/door/unlock` | POST | Unlock doors |
   | `remoteServices/engine/start` | POST | Remote start |
   | `remoteServices/engine/stop` | POST | Stop engine |
   | `remoteServices/ev/charge/start` | POST | Start charging |
   | `remoteServices/ev/charge/stop` | POST | Stop charging |
   | `remoteServices/light/on` | POST | Hazards on |
   | `remoteServices/light/off` | POST | Hazards off |
   | `remoteServices/hvac/on` | POST | Climate on |
   | `remoteServices/hvac/off` | POST | Climate off |

### Phase 3: CLI Commands

1. **Root command** - version, help, config path
2. **status** - aggregate view with subcommands
3. **lock/unlock** - door control
4. **start/stop** - engine control
5. **charge** - EV charging control
6. **climate** - HVAC control
7. **raw** - debug/raw JSON output

### Phase 4: Polish

1. Token caching (avoid re-auth on every command)
   - Store in `~/.cache/cx90/token.json`
   - Refresh when expired

2. Error handling
   - Friendly messages for common errors
   - "Account locked" warning
   - "Engine already running" handling

3. Interactive confirmations for destructive actions
   - `cx90 unlock` prompts unless `--yes` flag

4. Shell completions
   - `cx90 completion bash/zsh/fish`

## Build & Install

```bash
cd cx90
go build -o cx90 ./cmd/cx90
sudo mv cx90 /usr/local/bin/
```

## Example Output

```
$ cx90 status

CX-90 GT PHEV 2.5L I4 AT (2025)
VIN: JM3KKDHAXS1233203
Last Updated: 2025-12-18 05:36:11

BATTERY & CHARGING
  Battery:     66% ████████████████░░░░░░░░
  Plugged In:  Yes
  Charging:    Yes (220 min to full)

FUEL
  Tank:        92% ██████████████████████░░
  Range:       630 km

LOCATION
  50.84941, -118.986701

DOORS & LOCKS
  All doors locked
  All windows closed

TIRES (PSI)
  FL: 32  FR: 32
  RL: 32  RR: 32

CLIMATE
  HVAC: Off
  Interior: 22.9°C
```

## Notes

- The MyMazda API is unofficial and undocumented
- Mazda may block requests via TLS fingerprinting
- App version must be updated when Mazda releases new app versions
- Remote start limited to 2 consecutive starts without driving
