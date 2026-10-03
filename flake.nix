{
  description = "A tiny self-hosted secret manager";

  inputs = {
    # Declared explicitly so all contributors resolve the same nixpkgs source.
    # We use nixos-unstable to get modern tool versions (Go 1.26).
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };

        name = "yily";
        version = "1.0.0";

        devPkgs = with pkgs; [
          air
          go
          golangci-lint
          hivemind
          nilaway
          just
          nodejs
        ];
      in
      {
        devShells = {
          default = pkgs.mkShell {
            # `packages` is the modern, recommended attribute for development shells.
            packages = devPkgs;
            shellHook = ''
              unset GOROOT
              (cd client && npm install --silent)
              (cd server && go mod tidy)
              echo "welcome to the ${name} v${version} dev shell!"
            '';
          };
        };

        # Note: Packages (binary and docker) will be added when distribution
        # packaging is ready. Empty sets break `nix flake check`.
      }
    );
}