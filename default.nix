{ lib, buildGoModule, python3
, version ? (builtins.fromJSON (builtins.readFile ./plugin.json)).version
, revision ? null, release ? false }:
let
  metadata = import ./build-metadata.nix { inherit lib version revision release; };
in
buildGoModule {
  pname = "dankcalendar";
  version = metadata.version;
  src = metadata.source;
  vendorHash = null;
  subPackages = [ "cmd/dankcalendar" ];
  ldflags = [ "-s" "-w" "-X main.version=${metadata.version}" ];
  nativeBuildInputs = [ python3 ];
  postInstall = ''
    python3 scripts/package.py --stage-only --output "$out" \
      --revision ${lib.escapeShellArg metadata.revision} ${lib.optionalString release "--release"}
  '';
  meta = {
    description = "CalDAV CLI client for DankMaterialShell";
    homepage = "https://github.com/alcxyz/DankCalendar";
    license = lib.licenses.mit;
    mainProgram = "dankcalendar";
  };
}
