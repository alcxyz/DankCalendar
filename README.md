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

Use the published release flake so the helper and DMS manifest come from the
same package:

```nix
inputs.dankcalendar.url = "github:alcxyz/DankCalendar/v0.7.0";
```

```nix
{ inputs, pkgs, ... }:
let
  package = inputs.dankcalendar.packages.${pkgs.stdenv.hostPlatform.system}.release;
in {
  home.packages = [ package ];
  programs.dank-material-shell.plugins.dankCalendar = {
    enable = true;
    src = "${package}/share/dms-plugins/DankCalendar";
  };
}
```

### Manual

```sh
git clone --branch v0.7.0 https://github.com/alcxyz/DankCalendar.git
cd DankCalendar
python3 scripts/package.py --release --output dist/release
install -Dm755 dist/release/bin/dankcalendar ~/.local/bin/dankcalendar
mkdir -p ~/.config/DankMaterialShell/plugins/DankCalendar
cp -R dist/release/share/dms-plugins/DankCalendar/. \
  ~/.config/DankMaterialShell/plugins/DankCalendar/
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

## Build identity

The tracked `plugin.json` remains a release version. Development packages stamp
`X.Y.Z-dev.<commit>` (plus `.dirty` for local changes) into both the helper and
the installed manifest. Build them with:

```sh
python3 scripts/package.py --output dist/dev
```

Install `dist/dev/bin/dankcalendar` and use
`dist/dev/share/dms-plugins/DankCalendar` as the DMS plugin directory. In Nix,
use `packages.<system>.default` and its `share/dms-plugins/DankCalendar`
subdirectory, passing the source revision when using `callPackage`.

Release packages use the stable version: manual `--release` requires a clean
checkout at the manifest's `vX.Y.Z` tag; use `#release` with a published tag for
Nix release builds. Source-only Nix imports use a public-source fingerprint;
manual archives without Git metadata are labelled `dev.unknown`. Direct
`go build` identifies the helper's commit but does not stage a DMS manifest.
