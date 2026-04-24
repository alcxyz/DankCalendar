# Contributing to DankCalendar

## Development setup

Prerequisites: [DankMaterialShell](https://github.com/AvengeMedia/DankMaterialShell) >= 1.4.0, Go 1.22+, `secret-tool` (libsecret), `notify-send` (libnotify)

```bash
git clone https://github.com/alcxyz/DankCalendar.git
cd DankCalendar
```

Build the binary:

```bash
go build -o dankcalendar ./cmd/dankcalendar
```

For development, symlink the plugin into the DMS plugins directory:

```bash
ln -s "$(pwd)" ~/.config/DankMaterialShell/plugins/DankCalendar
cp dankcalendar ~/.local/bin/
```

Reload after changes:

```bash
dms ipc call plugins reload dankCalendar
```

## Project structure

- `plugin.json` -- plugin manifest (id, type, permissions)
- `CalendarWidget.qml` -- main widget component
- `CalendarSettings.qml` -- settings UI
- `cmd/dankcalendar/` -- CLI entry points (one file per subcommand)
- `internal/` -- Go packages (caldav, ical, keyring, config, output)

## Making changes

1. Fork the repo and create a branch from `dev`
2. Make your changes
3. Run tests: `go test ./...`
4. Test by reloading the plugin in DMS
5. Open a pull request against `dev`

## Commit messages

Use conventional-ish prefixes to keep history scannable:

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation only
- `chore:` maintenance, CI, dependencies
- `refactor:` code changes that don't add features or fix bugs

## Releasing

Releases are automated via GitHub Actions. The `VERSION` file is the single source of truth.

To cut a release:

1. Bump the `VERSION` file on `dev`
2. Merge `dev` into `main`
3. CI automatically creates the git tag and a GitHub release

### Version numbering

Follow [semver](https://semver.org/):

- **Patch** (`v0.1.x`): bug fixes, minor tweaks
- **Minor** (`v0.x.0`): new features, non-breaking changes
- **Major** (`vx.0.0`): breaking changes

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
