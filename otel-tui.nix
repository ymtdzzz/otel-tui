{
  config,
  pkgs,
  fetchFromGitHub,
  ...
}:

let
  otel-tui-version = "v0.7.5";
in
pkgs.buildGoModule {
  pname = "otel-tui";
  version = "${otel-tui-version}";
  src = pkgs.fetchFromGitHub {
    owner = "ymtdzzz";
    repo = "otel-tui";
    rev = "${otel-tui-version}";
    hash = "sha256-m+3yVFlB4cSJMiuBdbPZ7oEFLw32980bbZ0KUpb5FSY=";
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
  vendorHash = "sha256-kx1NyvQcAaWufJ5srQ5exiceWQaYUVViTjlHMTFpKQo=";
  subPackages = [ "." ];
}
