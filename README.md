# DankCalendar

CalDAV calendar plugin for [DankMaterialShell](https://github.com/AvengeMedia/DankMaterialShell). Single Go binary, stdlib-only, keyring-only credentials.

## Commands

| Command | Description |
|---|---|
| `dankcalendar list` | List upcoming events |
| `dankcalendar calendars` | Discover available calendars |
| `dankcalendar add` | Create a new event |
| `dankcalendar edit` | Modify an existing event |
| `dankcalendar delete` | Delete an event |
| `dankcalendar notify` | Send desktop notifications for upcoming events |
| `dankcalendar setup` | Configure CalDAV credentials |

## Install as DMS Plugin

1. Build the binary and place it in PATH:
   ```sh
   go build -o dankcalendar ./cmd/dankcalendar
   install dankcalendar ~/.local/bin/
   ```

2. Copy the plugin directory to DMS:
   ```sh
   cp -r . ~/.config/DankMaterialShell/plugins/DankCalendar/
   ```

3. Configure your CalDAV account in DMS plugin settings, or run:
   ```sh
   dankcalendar setup
   ```

## Build

```sh
go build -o dankcalendar ./cmd/dankcalendar
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
