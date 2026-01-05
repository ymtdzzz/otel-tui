{
  config,
  pkgs,
  fetchFromGitHub,
  ...
}:

let
  otel-tui-version = "v0.6.3";
in
pkgs.buildGoModule {
  pname = "otel-tui";
  version = "${otel-tui-version}";
  src = pkgs.fetchFromGitHub {
    owner = "ymtdzzz";
    repo = "otel-tui";
    rev = "${otel-tui-version}";
    hash = "sha256-uBsHOIMDh3pgENbCPlg35CY1XAU1EkkLHhMenjk0EA8=";
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
  vendorHash = "sha256-Hy6+7r026t48xWFqkb4wcYR01yNmK6JS3nVuTilwHhM=";
  subPackages = [ "." ];
}
