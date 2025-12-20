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
    crypto.go                AES-128-CBC, RSA encryption
    errors.go                Custom error types
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
    crypto.go                AES and PKCS7 padding utilities
  sensordata/
    sensor_data.go           Anti-bot fingerprinting (Feistel cipher)
```

## How the API Works

### Authentication Flow

1. `checkVersion` → Get `encKey` and `signKey` (AES keys for payload encryption)
2. `usher/system/encryptionKey` → Get RSA public key
3. `usher/user/login` → Encrypt password with RSA, get `accessToken`
4. All subsequent requests use `accessToken` + encrypted payloads

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
| `GetTiresInfo()` | `TireInfo` | FrontLeft, FrontRight, RearLeft, RearRight |
| `GetLocationInfo()` | `LocationInfo` | Latitude, Longitude, Timestamp |
| `GetDoorsInfo()` | `DoorStatus` | Driver, Passenger, RearLeft, RearRight, Hood, Trunk + lock status |
| `GetWindowsInfo()` | `WindowStatus` | FrontLeft, FrontRight, RearLeft, RearRight |
| `GetHvacInfo()` | `HVACInfo` | On, FrontDefroster, RearDefroster, Temperature, TemperatureUnit |
| `GetOdometerInfo()` | `OdometerInfo` | Kilometers |

All getters return `(T, error)` for proper error handling.

### Status Constants

Named constants for API status values (in `types.go`):

```go
// Charger status
ChargerConnected, ChargerDisconnected = 1, 0

// Charging status
ChargeStatusCharging = 6

// Battery heater
BatteryHeaterOn, BatteryHeaterOff = 1, 0
BatteryHeaterAutoEnabled, BatteryHeaterAutoDisabled = 1, 0

// HVAC
HVACStatusOn, HVACStatusOff = 1, 0

// Doors
DoorOpen, DoorClosed = 1, 0
DoorLocked, DoorUnlocked = 0, 1  // Note: inverted!

// Windows
WindowClosed, WindowFullyOpen = 0, 100
```

### Type-Safe Map Helpers

For working with `map[string]interface{}` responses safely (in `maphelpers.go`):

```go
getString(m, "key")      // (string, bool)
getInt(m, "key")         // (int, bool) - handles float64 from JSON
getFloat64(m, "key")     // (float64, bool)
getBool(m, "key")        // (bool, bool)
getMap(m, "key")         // (map[string]interface{}, bool)
getSlice(m, "key")       // ([]interface{}, bool)
```

These prevent runtime panics from unsafe type assertions.

## Common Gotchas

### Longitude Sign Bug
The API returns location in two places:
- `alertInfos[].PositionInfo` - correct longitude sign
- `remoteInfos[].PositionInfo` - sometimes wrong sign

Always use `alertInfos` for location data.

### InternalVIN Type
The `internalVin` field comes as either `string` or `float64` from JSON. This is handled automatically by the custom `InternalVIN` type which implements `UnmarshalJSON`.

### Token Caching
Credentials are cached in `~/.cache/mcs/token.json`. The cache stores:
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
