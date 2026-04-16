# gsjson

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
- `(*EnvJson).Setenv/Getenv/Unsetenv/Hasenv(key string)` — manage individual keys
- `(*EnvJson).GetOrCreateEnv(key, defaultValue string) string` — get or initialize key
- `(*EnvJson).GetDecryptedEnvValue() string` — return decrypted JSON string
- `(*EnvJson).Decode(encenv string) string` — decrypt an arbitrary encrypted string

### Package-level helpers
- `DecodeJsonFile(path string, passphrase []byte) (string, error)` — read and decrypt JSON file
- `DecodeJsonFileNoErr(path string, passphrase []byte) string` — same, ignoring errors
- `Init()` — initialize global singleton EnvJson (env key `ENVJSON`, password `ENVJSON`)
- `Setenv/Getenv/Unsetenv/Hasenv/GetOrCreateEnv(...)` — delegate to global singleton

## License

MIT License - see [LICENSE](LICENSE) for details.
