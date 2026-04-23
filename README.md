# dankcal

CalDAV CLI client for [DankMaterialShell](https://github.com/AvengeMedia/DankMaterialShell). Single Go binary, stdlib-only, keyring-only credentials.

Replaces the `qcal` submodule + `qcal-wrapper.py` Python bridge in [dms-qcal-calendar](https://github.com/alcxyz/dms-qcal-calendar) with a single binary that outputs JSON directly.

## Commands

| Command | Description |
|---|---|
| `dankcal list` | List upcoming events |
| `dankcal calendars` | Discover available calendars |
| `dankcal add` | Create a new event |
| `dankcal edit` | Modify an existing event |
| `dankcal delete` | Delete an event |
| `dankcal notify` | Send desktop notifications for upcoming events |
| `dankcal setup` | Configure CalDAV credentials |

## Build

```sh
go build -o dankcal ./cmd/dankcal
```

## Design

- **Single binary** — no Python, no submodules
- **Stdlib-only** — no external Go dependencies
- **Keyring-only** — passwords stored via `secret-tool`, never in config files
- **Security by default** — HTTPS-only, ICS escaping, path traversal protection, `0600` config
- **JSON output** — one JSON object per command on stdout, errors on stderr

See [docs/adr/](docs/adr/) for architectural decision records.

## Dependencies

- **Build**: Go 1.22+
- **Runtime**: `secret-tool` (libsecret), `notify-send` (libnotify)

## License

MIT
