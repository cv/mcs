# mcs Developer Guide

This document contains developer-focused documentation for the mcs CLI tool. For user-facing documentation (installation, usage, configuration), see [README.md](README.md).

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
  config/
    config.go                Config loading (TOML + env vars)
  crypto/
    crypto.go                Low-level AES-128-CBC and PKCS7 primitives
  sensordata/
    sensor_data.go           Anti-bot fingerprinting (16-round Feistel cipher, see line 255)
```

## How the API Works

### Authentication Flow

1. `service/checkVersion` (baseURL) → Get `encKey` and `signKey` (AES keys)
2. `system/encryptionKey` (usherURL) → Get RSA public key
3. `user/login` (usherURL) → Encrypt password with RSA, get `accessToken`
4. All subsequent requests use `accessToken` + encrypted payloads

Note: The API uses two base URLs - `baseURL` for vehicle operations and `usherURL` for authentication.

### Request Methods

Two methods for API requests:
- `APIRequest()` → Returns `map[string]interface{}` for dynamic access
- `APIRequestJSON()` → Returns raw bytes for direct unmarshaling to typed structs (preferred)

### Request Signing

Every request needs:
- `X-acf-sensor-data` header (anti-bot fingerprint using Feistel cipher)
- `sign` header (SHA256 of encrypted payload + timestamp + signKey)
- Encrypted query params and body (AES-128-CBC with encKey)

### Key Constants (internal/api/auth.go)

```go
AppVersion = "9.0.5"           // Must match mobile app version
IV = "0102030405060708"        // AES initialization vector
SignatureMD5 = "C383D8C4..."   // For key derivation
```

If the manufacturer updates the app, `AppVersion` and related user-agent strings may need updating.

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

For working with `map[string]interface{}` responses safely (in `maphelpers.go`):

```go
getString(m, "key")       // (string, bool)
getInt(m, "key")          // (int, bool) - handles float64 from JSON
getFloat64(m, "key")      // (float64, bool)
getBool(m, "key")         // (bool, bool)
getMap(m, "key")          // (map[string]interface{}, bool)
getSlice(m, "key")        // ([]interface{}, bool)
getMapSlice(m, "key")     // ([]map[string]interface{}, bool)
getMapFromSlice(s, idx)   // (map[string]interface{}, bool)
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
- `accessToken` + expiration timestamp
- `encKey` and `signKey`

Without caching, each command takes ~4.5s (full auth). With caching: ~2.7s.

## Testing

```bash
go test ./...              # Run all tests
golangci-lint run          # Lint check
go test -cover ./...       # With coverage
```

Integration tests cover config→client→cache flows. See `*_integration_test.go` files.

## Ticket Tracking

Uses `bd` (bead) for issue tracking. Database in `.beads/`.

```bash
bd list              # List issues
bd create "title"    # Create issue
bd close mcs-XX      # Close issue
bd ready             # Show unblocked work
```

## Reference Implementation

Ported from a reference Python implementation. The core API behavior follows the manufacturer's mobile app protocol. The anti-bot fingerprinting uses a 16-round Feistel cipher (documented in `sensordata/sensor_data.go`).
