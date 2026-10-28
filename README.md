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
country = "CA"   # Account country; defaults: US, GB, JP respectively
```

Or use environment variables: `MCS_EMAIL`, `MCS_PASSWORD`, `MCS_REGION`, `MCS_COUNTRY`

## Usage

```bash
# Status
mcs status              # Full vehicle status
mcs status --json       # JSON output
mcs status --refresh    # Request fresh status from vehicle

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

Note: Tire pressures are color-coded based on deviation from target (36 PSI):
- **Green**: within ±3 PSI (33-39)
- **Yellow**: 4-6 PSI deviation
- **Red**: >6 PSI deviation (potential safety issue)

## Claude Code Integration

mcs includes a Claude Code skill for natural language vehicle control. Install it once:

```bash
mcs skill install
```

Then in Claude Code, just tell Claude what you want:

```
You: "warm up the car"
Claude: runs `mcs climate on`

You: "what's my battery level?"
Claude: runs `mcs status` and shows battery info

You: "lock the doors"
Claude: runs `mcs lock`
```

**Supported phrases:**
- Climate: "warm up the car", "cool it down", "set to 22 degrees"
- Locks: "lock the car", "unlock the doors"
- Engine: "start the car", "stop the engine"
- Charging: "start charging", "stop the charge"
- Status: "check the battery", "where is my car", "tire pressure"

**Skill management:**
```bash
mcs skill install    # Install to ~/.claude/skills/mcs-control/
mcs skill uninstall  # Remove the skill
mcs skill path       # Show installation path
```

## Notes

- Uses vehicle manufacturer's API (reverse-engineered from mobile app)
- MyMazda 9.5 authentication: Azure AD B2C + the regional MC API. Tested with a
  Canadian CX-90 PHEV; Europe/Japan settings come from the app but are not live-tested.
- A new sign-in/session attachment can sign the phone app out. Access tokens,
  refresh tokens and the device session are reused from `~/.cache/mcs/token.json`
  (atomic writes, mode 0600). Expired access tokens use refresh, not password login.
  A refused refresh does **not** fall back to another login. If the phone takes over,
  do not run a login loop; choose which client should hold the session. To deliberately
  sign in again, remove the cache and rerun the command.
- Plain `status` reads Mazda's cached reports; it does not request a wake-up/refresh.
  `status --json` includes `reported_at.ev`, `reported_at.doors` and
  `reported_at.vehicle` (Mazda's UTC `yyyyMMddHHmmss`). These reports can have very
  different ages. A successful cloud poll does **not** make an old door report current.
- `MCS_DEBUG=1` logs request metadata but redacts tokens, sessions, sensor-data,
  query values and bodies. Raw status/JSON still contains VIN and location: keep it private.
- Remote start limited to 2 consecutive starts without driving

For developer documentation, see [CLAUDE.md](CLAUDE.md)
