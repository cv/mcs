# mcs Developer Guide

## Project Structure

```
cmd/mcs/main.go           Entry point
internal/
  api/
    auth.go                Authentication, encryption keys, login
    client.go              API request handling with retry logic
    crypto.go              AES-128-CBC, RSA, MD5/SHA256
    endpoints.go           Vehicle control endpoints (lock, start, etc.)
    errors.go              Custom error types
  cache/
    cache.go               Token caching (~/.cache/mcs/token.json)
  cli/
    root.go                Cobra root command
    status.go              Status command + subcommands
    lock.go, engine.go     Control commands
    charge.go, climate.go  EV/HVAC commands
    raw.go                 Debug raw JSON output
    client.go              Helper to create API client with caching
  config/
    config.go              Config loading (TOML + env vars)
  sensordata/
    sensor_data.go         Anti-bot fingerprinting (X-acf-sensor-data header)
```

## How the API Works

### Authentication Flow

1. `checkVersion` → Get `encKey` and `signKey` (AES keys for payload encryption)
2. `usher/system/encryptionKey` → Get RSA public key
3. `usher/user/login` → Encrypt password with RSA, get `accessToken`
4. All subsequent requests use `accessToken` + encrypted payloads

### Request Signing

Every request needs:
- `X-acf-sensor-data` header (anti-bot fingerprint, see `sensordata/`)
- `sign` header (SHA256 of encrypted payload + timestamp + signKey)
- Encrypted query params and body (AES-128-CBC with encKey)

### Key Constants (internal/api/auth.go)

```go
AppVersion = "9.0.5"           // Must match mobile app version
IV = "0102030405060708"        // AES initialization vector
SignatureMD5 = "C383D8C4..."   // For key derivation
```

If the manufacturer updates the app, `AppVersion` and related user-agent strings may need updating.

## Common Gotchas

### Longitude Sign Bug
The API returns location in two places:
- `alertInfos[].PositionInfo` - correct longitude sign
- `remoteInfos[].PositionInfo` - sometimes wrong sign

Always use `alertInfos` for location data.

### internalVin Type
The `internalVin` field comes back as `float64`, not `string`. Use type switch:
```go
switch v := cvInfo["internalVin"].(type) {
case string:
    internalVIN = v
case float64:
    internalVIN = fmt.Sprintf("%.0f", v)
}
```

### Token Caching
Credentials are cached in `~/.cache/mcs/token.json`. The cache stores:
- `accessToken` + expiration timestamp
- `encKey` and `signKey`

Without caching, each command takes ~4.5s (full auth). With caching: ~2.7s.

## Testing

```bash
go test ./...
golangci-lint run
```

## Ticket Tracking

Uses `bd` (bead) for issue tracking. Database in `.beads/`.

```bash
bd list              # List issues
bd create "title"    # Create issue
bd close mcs-XX      # Close issue
```

## Reference Implementation

This was ported from a reference Python implementation. The core API behavior follows the manufacturer's mobile app protocol.
