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

# Lights
mcs lights on           # Hazard lights on (find the car)
mcs lights off          # Hazard lights off

# Debug
mcs raw status          # Raw vehicle status JSON
mcs raw ev              # Raw EV status JSON
mcs raw health          # Raw health report (oil life, maintenance, warnings)
mcs raw hvac            # Raw remote climate setting

# Shell completions
mcs completion bash     # Also: zsh, fish, powershell
```

## MQTT and Home Assistant

`mcs mqtt` runs until stopped: it polls the car every 5 minutes, publishes an
allowlisted snapshot (no VIN, coordinates or ids) to `<prefix>/state`, announces
every reading to Home Assistant through MQTT discovery and, only for the groups
given in `--controls`, carries out commands. It never signs in with a password
nor attaches a new session (as with `MCS_REFRESH_ONLY=1`), so the phone stays
signed in.

```bash
mcs mqtt --prefix mazda/cx90 --once --dry-run      # one poll, printed, nothing sent
mcs mqtt --broker tcp://127.0.0.1:1883 --username mazda \
         --password-file /etc/mcs/mqtt.pw --prefix mazda/cx90   # read-only
mcs mqtt --prefix mazda/cx90 --broker tcp://127.0.0.1:1883 \
         --controls all --home 45.50,-73.57        # everything, with presence
```

The prefix and the broker have no default. Settings can also live in the config
file (or `MCS_MQTT_<KEY>` variables; the password only in `MCS_MQTT_PASSWORD` or
`password_file`):

```toml
[mqtt]
broker = "tcp://127.0.0.1:1883"
username = "mazda"
password_file = "/etc/mcs/mqtt.pw"
prefix = "mazda/cx90"
controls = ["lock", "unlock", "climate", "refresh"]   # also engine, charge, lights; "all"
home = "45.50,-73.57"      # home / not_home; coordinates are never published
# discovery = "homeassistant" (or "off"), device_name, client_id, home_radius_m = 200,
# poll = "5m", health = "30m", stale = "30m", command_interval = "1m",
# engine_interval = "10m", refresh_interval = "30m", state_dir, wait_on_auth_loss
```

- Topics: `<prefix>/state` (JSON), `status` (`online`/`offline`, also the will),
  `error`, `command_status`, `refresh_status`; commands on `<prefix>/cmd/<name>/set`
  (`PRESS` for buttons; `climate_temperature` 16..28 °C; `front_defroster` /
  `rear_defroster` `ON`/`OFF`) and `<prefix>/refresh/set` (a fresh EV report).
  Retained command messages never run.
- Locks, doors, windows and the climate state turn unavailable in Home Assistant
  once their report is older than `stale`; battery, fuel, tires and the health
  report keep their last value with its time as `reported_at`.
- Each command goes out at most once per `command_interval` (engine start: per
  `engine_interval`; fresh report: per `refresh_interval`), across restarts.
- Exit status: 0 stopped, 1 error, 2 bad configuration, 3 the Mazda session is
  lost (renew it with `mcs status` on a computer and copy the cache; do not
  restart in a loop). Under Docker, which restarts whatever the status, set
  `wait_on_auth_loss` to stay up offline instead. Network trouble never exits.
- About 7 MB of memory (a 31 MB distroless image); 2 API calls per poll, 4 every 30 minutes.

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

## Agent Integration

mcs ships an [Agent Skills](https://agentskills.io) skill (`SKILL.md`) for natural language vehicle control from any coding agent that supports skills. Install it once:

```bash
mcs skill install
```

Then just tell your agent what you want:

```
You: "warm up the car"
Agent: runs `mcs climate on`

You: "what's my battery level?"
Agent: runs `mcs status` and shows battery info

You: "lock the doors"
Agent: runs `mcs lock`
```

**Supported phrases:**
- Climate: "warm up the car", "cool it down", "set to 22 degrees"
- Locks: "lock the car", "unlock the doors"
- Engine: "start the car", "stop the engine"
- Charging: "start charging", "stop the charge"
- Status: "check the battery", "where is my car", "tire pressure"

**Skill management:**
```bash
mcs skill install    # Install to ~/.agents/skills/mcs-control/
mcs skill uninstall  # Remove the skill
mcs skill path       # Show installation path
```

The skill installs to `~/.agents/skills/` by default. If your agent reads skills from a different directory, set `MCS_SKILLS_DIR`:

```bash
MCS_SKILLS_DIR=/path/to/agent/skills mcs skill install
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
- For unattended monitoring set `MCS_REFRESH_ONLY=1`: cached sessions can refresh
  access tokens, but no password login or session attachment is allowed. Provision
  a cache interactively first; the service does not need the account password.
- Plain `status` reads Mazda's cached reports; it does not request a wake-up/refresh.
  `status --json` includes `reported_at.ev`, `reported_at.doors` and
  `reported_at.vehicle` (Mazda's UTC `yyyyMMddHHmmss`). These reports can have very
  different ages. A successful cloud poll does **not** make an old door report current.
- `MCS_DEBUG=1` logs request metadata but redacts tokens, sessions, sensor-data,
  query values and bodies. Raw status/JSON still contains VIN and location: keep it private.
- Remote start limited to 2 consecutive starts without driving

For developer documentation, see [AGENTS.md](AGENTS.md)
