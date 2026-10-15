{ pkgs, system }:
let
  sources = {
    version = "1.2.0";

    platforms = {
      x86_64-linux = {
        target = "x86_64-unknown-linux-gnu";
        sha256 = "489e2493d10499d1e0f19f0c7938b1baa5cd488be1f2e3abd3fbf820131452bf";
      };
      aarch64-linux = {
        target = "aarch64-unknown-linux-gnu";
        sha256 = "42d25a6676c28d0749da7edc97003a7a34c5e31ea1ea479f3d8239ad16d09d2d";
      };
      x86_64-darwin = {
        target = "x86_64-apple-darwin";
        sha256 = "b19096e684a6d6924a7e3792e2690b331f7c8486fcc3ead62335eb7a09ac0e1c";
      };
      aarch64-darwin = {
        target = "aarch64-apple-darwin";
        sha256 = "a4dd88b5d6b9e7dfa93694fa11e5d27213ae163052ae4fd5f4301723064bef8e";
      };
    };
  };
  platform = sources.platforms.${system};
in
pkgs.stdenv.mkDerivation rec {
  pname = "vite-plus";
  version = sources.version;
  src = pkgs.fetchurl {
    url = "https://github.com/voidzero-dev/vite-plus/releases/download/v${version}/vp-${platform.target}.tar.gz";
    sha256 = platform.sha256;
  };

  sourceRoot = ".";

  nativeBuildInputs = [ pkgs.autoPatchelfHook ];
  buildInputs = [ pkgs.stdenv.cc.cc.lib ];

  installPhase = ''
    runHook preInstall
    install -Dm755 "$(find . -type f -name vp | head -n1)" $out/bin/vp
    runHook postInstall
  '';

  meta = with pkgs.lib; {
    description = "The unified toolchain and entry point for web development";
    homepage = "https://github.com/voidzero-dev/vite-plus";
    license = licenses.mit;
    mainProgram = "vp";
  };
}
