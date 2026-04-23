{ lib, buildGoModule, version ? "dev" }:

buildGoModule {
  pname = "dankcalendar";
  inherit version;

  src = ./.;

  vendorHash = null;

  subPackages = [ "cmd/dankcalendar" ];

  ldflags = [ "-s" "-w" "-X main.version=${version}" ];

  meta = with lib; {
    description = "CalDAV CLI client for DankMaterialShell";
    homepage = "https://github.com/alcxyz/DankCalendar";
    license = licenses.mit;
    mainProgram = "dankcalendar";
  };
}
