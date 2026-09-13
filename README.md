# DankCalendar

CalDAV calendar plugin for
[DankMaterialShell](https://github.com/AvengeMedia/DankMaterialShell). One
stdlib-only Go binary talks to your CalDAV server; credentials live only in
the system keyring.

![Screenshot](docs/screenshot.png)

## What you get

- **Upcoming events in the bar**, with the popout to create, edit, and delete
  them.
- **Reminders** as desktop notifications.
- **Any CalDAV server**, plus Google Workspace through your own OAuth desktop
  client.
- **Multiple calendars**, normalised to one timezone so cross-calendar
  ordering is correct, with recurring events expanded server-side and a
  client-side fallback for subscribed calendars.
- **Secure by default**: HTTPS only, keyring-only secrets, `0600` config.

If you already manage calendars with khal or Evolution, the
[khal-calendar](https://github.com/fishman/dms-khal-calendar),
[qCal Calendar](https://github.com/szabolcsf/dms-qcal-calendar), or
[Calendar for DMS](https://github.com/arqueon/dms-calendar) plugins may fit
your existing setup better. DankCalendar is for when you want account setup
and CalDAV handled by its own backend.

## Requirements

Go 1.22+ to build. At runtime: `secret-tool` (libsecret) and `notify-send`
(libnotify).

## Install

### Nix (flake)

Add as a `flake = false` input and include it in your DMS plugin
configuration:

```nix
inputs.dms-plugin-calendar = {
  url = "github:alcxyz/DankCalendar";
  flake = false;
};
```

```nix
programs.dank-material-shell.plugins.dankCalendar = {
  enable = true;
  src = inputs.dms-plugin-calendar;
};
```

### Manual

```sh
go build -o dankcalendar ./cmd/dankcalendar
cp dankcalendar ~/.local/bin/
cp -r . ~/.config/DankMaterialShell/plugins/DankCalendar/
```

## Set up an account

Configure your CalDAV account in DMS plugin settings, or run:

```sh
dankcalendar setup
```

Google accounts need OAuth; basic auth and app passwords are rejected by
Google's CalDAV endpoint. Follow
[Google Workspace and OAuth](docs/google-oauth.md).

## Learn more

| Topic | Read |
|---|---|
| Google Workspace and OAuth setup | [docs/google-oauth.md](docs/google-oauth.md) |
| Command-line reference | [docs/cli.md](docs/cli.md) |
| Design decisions | [docs/adr/README.md](docs/adr/README.md) |
| Development, design boundaries, and releasing | [CONTRIBUTING.md](CONTRIBUTING.md) |

## License

[MIT](LICENSE)

<details>
<summary>Support</summary>

- **BTC:** `bc1pzdt3rjhnme90ev577n0cnxvlwvclf4ys84t2kfeu9rd3rqpaaafsgmxrfa`
- **ETH / ERC-20:** `0x2122c7817381B74762318b506c19600fF8B8372c`
</details>
