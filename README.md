# gsjson

[![Go Reference](https://pkg.go.dev/badge/github.com/sonnt85/gsjson.svg)](https://pkg.go.dev/github.com/sonnt85/gsjson)

JSON utility library for Go — thread-safe in-memory JSON store, file-backed JSON with encryption, and encrypted environment variable JSON store.

## Installation

```bash
go get github.com/sonnt85/gsjson
```

## Features

- `Ejson` — thread-safe in-memory JSON buffer with get/set/delete via dotted key paths (backed by `gjson`/`sjson`)
- `Fjson` — file-backed JSON with file locking, optional AES encryption/decryption, fsnotify watch, and auto-persist
- `EnvJson` — store encrypted JSON inside an environment variable; get/set individual keys transparently
- Change hooks: register a callback triggered on any mutation
- Load hooks: transform raw bytes on load (e.g., decrypt)
- Atomic file writes using `lockedfile` (lock with timeout)
- Decode helper: read and decrypt a JSON file to string without creating a full `Fjson`

## Usage

```go
import "github.com/sonnt85/gsjson"

// In-memory JSON store
ej, _ := gsjson.NewGsjson([]byte(`{"name":"alice","age":30}`), nil)
result := ej.Get("name")          // gjson.Result
ej.Set("age", 31)
ej.Delete("name")
fmt.Println(ej.String())

// File-backed JSON (plaintext)
fj, _ := gsjson.NewFjson("/var/data/config.json", nil, true, map[string]interface{}{
    "version": "1.0",
})
fj.Set("version", "2.0")
fj.WriteToFile()

// File-backed JSON (AES-encrypted)
fj, _ := gsjson.NewFjson("/var/data/config.enc", []byte("mypassphrase"), true)
fj.Set("token", "secret")
fj.WriteToFile()

// Environment variable JSON store
ev := gsjson.NewEnvJson("APP_CONFIG", "mypassword")
ev.Setenv("db.host", "localhost")
ev.Setenv("db.port", "5432")
host := ev.Getenv("db.host")

// Global singleton env-json
gsjson.Init()
gsjson.Setenv("key", "value")
val := gsjson.Getenv("key")
```

## API

### Ejson (in-memory)
- `NewGsjson(data []byte, hookChange func(), forceInit ...bool) (*Ejson, error)` — create from bytes
- `(*Ejson).Get(path string) gjson.Result` — get value at dotted path
- `(*Ejson).Set(path string, value interface{}) error` — set value
- `(*Ejson).SetRaw(path string, value []byte) error` — set raw JSON bytes
- `(*Ejson).Delete(path string) error` — remove key
- `(*Ejson).LoadAndDelete(path string) gjson.Result` — atomic get-and-delete
- `(*Ejson).LoadOrStore(path string, value any) gjson.Result` — get existing or store default
- `(*Ejson).Loads() gjson.Result` — parse entire buffer
- `(*Ejson).String() string` — return raw JSON string
- `(*Ejson).WriteTo(path string, password []byte) error` — write to file (optionally encrypted)
- `(*Ejson).LoadNewData(buffer []byte) error` — replace buffer with new valid JSON
- `(*Ejson).SetHookLoad(f func(*[]byte))` — register load transform hook

### Fjson (file-backed)
- `NewFjson(path string, passphrase []byte, removeIfInvalid bool, datas ...map[string]interface{}) (*Fjson, error)` — open or create JSON file
- `(*Fjson).ReadFile(paths ...string) ([]byte, error)` — read and (optionally) decrypt file
- `(*Fjson).WriteToFile()` — async persist to file
- `(*Fjson).Remove() error` — delete the backing file
- `(*Fjson).GetJsonFilePath() string` — return file path
- `(*Fjson).GetPassPhrase() []byte` — return passphrase
- Inherits all `Ejson` methods

### EnvJson (environment variable)
- `NewEnvJson(envkey, pwd string) *EnvJson` — create an env-JSON handle
- `GetGenvjson() *EnvJson` — return the global singleton EnvJson instance
- `(*EnvJson).Setenv(key, value string) error` / `Getenv(key string) string` / `Unsetenv(key string) error` / `Hasenv(key string) bool` — manage individual keys
- `(*EnvJson).GetOrCreateEnv(key, defaultValue string) string` — get or initialize key
- `(*EnvJson).GetDecryptedEnvValue() string` — return decrypted JSON string from the environment variable
- `(*EnvJson).GetDecryptedEnvValueThenAddEnv(m map[string]string) string` — decrypt JSON and merge extra env entries
- `(*EnvJson).GetEncryptedEnvValueThenAddEnv(m map[string]string) string` — encrypt JSON with extra entries merged
- `(*EnvJson).SetEnvFromDecryptedValue(value string)` — store a plaintext JSON string back as an encrypted env variable
- `(*EnvJson).GetEncryptedEnvValue() string` — return the raw (encrypted) env variable value
- `(*EnvJson).GetEnvName() string` — return the environment variable key name
- `(*EnvJson).Decode(encenv string) string` — decrypt an arbitrary encrypted string
- `(*EnvJson).String() string` — return the decrypted JSON as a plain string

### Package-level helpers
- `DecodeJsonFile(path string, passphrase []byte) (string, error)` — read and decrypt JSON file
- `DecodeJsonFileNoErr(path string, passphrase []byte) string` — same, ignoring errors
- `Init()` — initialize global singleton EnvJson (env key `ENVJSON`, password `ENVJSON`)
- `GetGenvjson() *EnvJson` — return the global singleton instance
- `Decode(encenv string) string` — decrypt using global singleton
- `String() string` — JSON string via global singleton
- `Getenv(key string) string` / `Setenv(key, value string) error` / `Unsetenv(key string) error` / `Hasenv(key string) bool` — delegate to global singleton
- `GetDecryptedEnvValueThenAddEnv(m map[string]string) string` — package-level variant
- `GetEncryptedEnvValueThenAddEnv(m map[string]string) string` — package-level variant
- `SetEnvFromDecryptedValue(value string)` — package-level variant
- `GetEncryptedEnvValue() string` — package-level variant
- `GetEnvName() string` — package-level variant

## Author

**sonnt85** — [thanhson.rf@gmail.com](mailto:thanhson.rf@gmail.com)

## License

MIT License - see [LICENSE](LICENSE) for details.
