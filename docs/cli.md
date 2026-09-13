# Command-line reference

The `dankcalendar` binary is what the widget calls. Every command prints one
JSON object on stdout and errors on stderr, so it is also usable from scripts.

| Command | Description |
|---|---|
| `dankcalendar list` | List upcoming events |
| `dankcalendar calendars` | Discover available calendars |
| `dankcalendar add` | Create a new event |
| `dankcalendar edit` | Modify an existing event |
| `dankcalendar delete` | Delete an event |
| `dankcalendar notify` | Send desktop notifications for upcoming events |
| `dankcalendar setup` | Configure CalDAV credentials |
| `dankcalendar google-discover` | Authorize and discover Google calendars |

Passwords and OAuth refresh tokens are stored in the system keyring via
`secret-tool`, never in the config file. Google setup is described in
[Google Workspace and OAuth](google-oauth.md).
