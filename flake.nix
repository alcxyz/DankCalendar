{
  description = "CalDAV CLI client for DankMaterialShell";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = (builtins.fromJSON (builtins.readFile ./plugin.json)).version;
      in {
        packages = rec {
          dankcalendar = pkgs.callPackage ./default.nix { inherit version; };
          default = dankcalendar;
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [ go gopls gotools ];
        };
      }
    );
}
