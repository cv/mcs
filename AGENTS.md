# mcs Developer Guide

Developer and agent documentation for the mcs CLI. For user-facing documentation (installation, usage, configuration), see [README.md](README.md).

Issues are tracked in GitHub Issues (`gh issue list`, `gh issue create`, `gh issue close <n>`).

## Modern Go

Follow [JetBrains Modern Go Guidelines](https://github.com/JetBrains/go-modern-guidelines) via the `use-modern-go` skill (`.agents/skills/use-modern-go`). Before editing Go code, run:

```bash
sh .agents/skills/use-modern-go/scripts/run-tool.sh list --file-path <file.go>
sh .agents/skills/use-modern-go/scripts/run-tool.sh explain <guideline_id>
```

Read the full `list` output; it is scoped to the `go` version in `go.mod`. Update with `npx skills update use-modern-go -p -y` (tracked in `skills-lock.json`).

In practice: `t.Context()` instead of `context.Background()` in tests, `any` instead of `interface{}`, `errors.AsType[T]` instead of `errors.As`, and stdlib `slices`/`maps`/`cmp` helpers. The `modernize` linter enforces most of this.

## Project Structure

```
cmd/mcs/main.go              Entry point
internal/
  api/
    auth.go                  Authentication, encryption keys, login
    client.go                API request handling with retry logic
    control.go               Vehicle control endpoints (lock, start, etc.)
    crypto.go                API wrappers (base64, RSA, uses fixed IV)
    errors.go                Custom error types
    keys.go                  Encryption key storage struct
    maphelpers.go            Type-safe map accessor functions
    types.go                 Response types and data structures
    vehicle.go               Vehicle data retrieval endpoints
  cache/
    cache.go                 Token caching (~/.cache/mcs/token.json)
  cli/
    root.go                  Cobra root command
    client.go                API client creation with caching
    command_factory.go       Command builder helpers
    status_cmd.go            Status command orchestration
    status_display.go        Status display formatting
    status_extract.go        Data extraction for JSON output
    status_format.go         Formatting helpers
    lock.go, engine.go       Control commands
    charge.go, climate.go    EV/HVAC commands
    raw.go                   Debug raw JSON output
    mqtt_cmd.go              mcs mqtt: flags, config, the Vehicle over the API client
  config/
    config.go                Config loading (TOML + env vars)
    mqtt.go                  The [mqtt] section (MCS_MQTT_* env)
  crypto/
    crypto.go                Low-level AES-128-CBC and PKCS7 primitives
  mqttd/                     mcs mqtt: snapshot, discovery, commands, poll loop, brokers
  sensordata/
    sensor_data.go           Anti-bot fingerprinting (16-round Feistel cipher, see line 255)
```

## mcs mqtt (internal/mqttd)

The daemon is pure apart from two interfaces, `Vehicle` (implemented in
`internal/cli/mqtt_cmd.go` with the API client) and `Broker` (paho, or the
printing one for `--dry-run`), so `daemon_test.go` drives it with fakes. Keep
the published snapshot an allowlist (`snapshot.go`): never VIN, coordinates,
nickname or ids; log result codes, never raw errors. Discovery unique ids are
`mcs_<last prefix level>_<key>`: changing them makes new HA entities. Disabled
controls get empty retained configs, so their entities disappear.

## How the API Works

### Authentication Flow

1. Azure AD B2C authorization code + PKCE (account country required) → access and refresh tokens.
2. Regional MC host `service/checkVersion` → AES/signing keys. Derivation uses the Android
   signing certificate's SHA-256 digest, rather than the legacy MD5 digest.
3. `remoteServices/attach/v4` → `data.userinfo.sessionId`; persist it.
4. Vehicle requests carry access-token, Authorization Bearer and X-device-session-id.
   Refresh access tokens without reattaching; do not fight a phone takeover by signing in again.

North America now uses hgs2ivna.mazda.com, without /prod/. The legacy 0cxo7m58 host
returns generic failures. Legacy Usher helpers remain for compatibility tests, not login.
See b2c.go/session.go; all automated tests use local mock servers.

### Request Methods

Two methods for API requests:
- `APIRequest()` → Returns `map[string]any` for dynamic access
- `APIRequestJSON()` → Returns raw bytes for direct unmarshaling to typed structs (preferred)

### Request Signing

Every request needs:
- `X-acf-sensor-data` header (anti-bot fingerprint using Feistel cipher)
- `sign` header (SHA256 of encrypted payload + timestamp + signKey)
- Encrypted query params and body (AES-128-CBC with encKey)

### Key Constants (internal/api/auth.go)

```go
AppVersion = "9.5.0"           // Must match mobile app version
IV = "0102030405060708"        // AES initialization vector
SignatureMD5 = "C383D8C4..."   // For key derivation
```

If the manufacturer updates the app, `AppVersion` and related user-agent strings may need updating.

### Confirmation Polling

Control commands (lock, start, climate, etc.) support confirmation polling to wait for the vehicle to report the new state:
- 20s initial delay before first poll (allows vehicle time to process)
- 5s poll interval thereafter
- Configurable timeout via `--confirm-wait` flag (default: 2 minutes)
- Disable with `--confirm=false`

Constants in `command_factory.go`:
```go
ConfirmationInitialDelay = 20 * time.Second
DefaultPollInterval      = 5 * time.Second
```

## Data Structures

### Response Structs

Getter methods return strongly-typed structs instead of tuples:

| Method | Returns | Fields |
|--------|---------|--------|
| `GetBatteryInfo()` | `BatteryInfo` | BatteryLevel, RangeKm, ChargeTimeACMin, ChargeTimeQBCMin, PluggedIn, Charging, HeaterOn, HeaterAuto |
| `GetFuelInfo()` | `FuelInfo` | FuelLevel, RangeKm |
| `GetTiresInfo()` | `TireInfo` | FrontLeftPsi, FrontRightPsi, RearLeftPsi, RearRightPsi |
| `GetLocationInfo()` | `LocationInfo` | Latitude, Longitude, Timestamp |
| `GetDoorsInfo()` | `DoorStatus` | DriverOpen, PassengerOpen, RearLeftOpen, RearRightOpen, TrunkOpen, HoodOpen, FuelLidOpen, DriverLocked, PassengerLocked, RearLeftLocked, RearRightLocked, AllLocked |
| `GetWindowsInfo()` | `WindowStatus` | DriverPosition, PassengerPosition, RearLeftPosition, RearRightPosition |
| `GetHvacInfo()` | `HVACInfo` | HVACOn, FrontDefroster, RearDefroster, InteriorTempC, TargetTempC |
| `GetOdometerInfo()` | `OdometerInfo` | OdometerKm |

All getters return `(T, error)` for proper error handling.

### Status Constants

Named constants for API status values (in `types.go`):

```go
// Temperature units
Celsius, Fahrenheit = 1, 2

// Result codes
ResultCodeSuccess = "200S00"

// Charger status
ChargerConnected, ChargerDisconnected = 1, 0

// Charging status
ChargeStatusCharging, ChargeStatusNotCharging = 6, 0

// Battery heater
BatteryHeaterOn, BatteryHeaterOff = 1, 0
BatteryHeaterAutoEnabled, BatteryHeaterAutoDisabled = 1, 0

// HVAC
HVACStatusOn, HVACStatusOff = 1, 0

// Defrosters
DefrosterOn, DefrosterOff = 1, 0

// Doors
DoorOpen, DoorClosed = 1, 0
DoorLocked, DoorUnlocked = 0, 1  // Note: inverted!

// Hazard lights
HazardLightsOn, HazardLightsOff = 1, 0

// Windows
WindowClosed, WindowFullyOpen = 0, 100
```

### Type-Safe Map Helpers

For working with `map[string]any` responses safely (in `maphelpers.go`):

```go
getString(m, "key")       // (string, bool)
getInt(m, "key")          // (int, bool) - handles float64 from JSON
getFloat64(m, "key")      // (float64, bool)
getBool(m, "key")         // (bool, bool)
getMap(m, "key")          // (map[string]any, bool)
getSlice(m, "key")        // ([]any, bool)
getMapSlice(m, "key")     // ([]map[string]any, bool)
getMapFromSlice(s, idx)   // (map[string]any, bool)
```

These prevent runtime panics from unsafe type assertions.

### Error Types

Custom error types in `errors.go` for specific error handling:

```go
// Error codes from API
ErrorCodeEncryption   = 600001  // Server rejected encrypted request
ErrorCodeTokenExpired = 600002  // Access token expired
ErrorCodeRequestIssue = 920000  // Check ExtraCode for details

// Extra codes (used with ErrorCodeRequestIssue)
ExtraCodeRequestInProgress = "400S01"  // Request already in progress
ExtraCodeEngineStartLimit  = "400S11"  // Engine start limit reached

// Error types (use errors.Is/errors.As for checking)
*APIError              // General API error
*EncryptionError       // Triggers key refresh and retry
*TokenExpiredError     // Triggers re-login and retry
*RequestInProgressError // Vehicle is processing another request
*EngineStartLimitError  // Remote start limit (2x) reached
*ResultCodeError        // Unexpected result code from API
```

## Common Gotchas

### Longitude Sign Bug
The API returns location in two places:
- `alertInfos[].PositionInfo` - correct longitude sign
- `remoteInfos[].PositionInfo` - sometimes wrong sign

Always use `alertInfos` for location data.

### InternalVIN Type
The `internalVin` field comes as either `string` or `float64` from JSON. This is handled automatically by the custom `InternalVIN` type which implements `UnmarshalJSON`.

### Token Caching
Credentials are cached in `~/.cache/mcs/token.json` (see `cache/cache.go`). The cache stores:
- `accessToken` + expiration, refresh token and device session ID
- `encKey` and `signKey`
- protocol version and account/country scope (legacy and other-account caches rejected)

Without caching, each command takes ~4.5s (full auth). With caching: ~2.7s.

## Testing

```bash
go test ./...              # Run all tests
golangci-lint run          # Lint check
go test -cover ./...       # With coverage
```

Integration tests cover config→client→cache flows. See `*_integration_test.go` files.

### Assertion Guidelines

Use specific testify assertions for better failure messages:

| Instead of | Use |
|-----------|-----|
| `assert.True(t, strings.Contains(s, sub))` | `assert.Contains(t, s, sub)` |
| `assert.True(t, bytes.Equal(a, b))` | `assert.Equal(t, a, b)` |
| `assert.True(t, regexp.MatchString(p, s))` | `assert.Regexp(t, p, s)` |
| `assert.True(t, err != nil)` | `assert.Error(t, err)` |
| `assert.True(t, x == y)` | `assert.Equal(t, x, y)` |
| `assert.False(t, x != y)` | `assert.Equal(t, x, y)` |
| `assert.False(t, s != "")` | `assert.Empty(t, s)` |
| `require.False(t, len(x) == 0)` | `require.NotEmpty(t, x)` |
| `assert.True(t, x == 0)` | `assert.Zero(t, x)` |

Omit assertion messages when the assertion type and variable names are self-explanatory:
```go
// Good - variable name is descriptive
assert.True(t, callbackExecuted)
assert.NotNil(t, client)

// Good - adds context not obvious from the assertion
assert.Equal(t, 0600, perm, "cache file permissions")
```

## Reference Implementation

Ported from a reference Python implementation. The core API behavior follows the manufacturer's mobile app protocol. The anti-bot fingerprinting uses a 16-round Feistel cipher (documented in `sensordata/sensor_data.go`).

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create GitHub issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished issues, comment on in-progress ones
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds

