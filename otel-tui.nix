{
  config,
  pkgs,
  fetchFromGitHub,
  ...
}:

let
  otel-tui-version = "v0.5.3";
in
pkgs.buildGoModule {
  pname = "otel-tui";
  version = "${otel-tui-version}";
  src = pkgs.fetchFromGitHub {
    owner = "ymtdzzz";
    repo = "otel-tui";
    rev = "${otel-tui-version}";
    hash = "sha256-tlUv8nI2KKhlp8jofAstwhYm0KwpSz2QGuZg52hjGyA=";
  };
  overrideModAttrs = (
    _: {
      buildPhase = ''
        go work vendor
      '';
    }
  );
  ldflags = [
    "-X main.version=${otel-tui-version}"
  ];
  vendorHash = "sha256-nM2Sa9YchKHoPmoPUy59cbGl4ejyLfcV9twv2d654PA=";
  subPackages = [ "." ];
  buildInputs = pkgs.lib.optionals pkgs.stdenv.isLinux [
    pkgs.xorg.libX11
  ];
}
