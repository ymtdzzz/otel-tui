{
  config,
  pkgs,
  fetchFromGitHub,
  ...
}:

let
  otel-tui-version = "v0.7.6";
in
pkgs.buildGoModule {
  pname = "otel-tui";
  version = "${otel-tui-version}";
  src = pkgs.fetchFromGitHub {
    owner = "ymtdzzz";
    repo = "otel-tui";
    rev = "${otel-tui-version}";
    hash = "sha256-0H/9BxrxYLnFhJ03+GcW+fsQyf1+3jxHG1pLUZun/7g=";
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
  vendorHash = "sha256-WMwQlFKxBPPMzD2ay6ikQBO1B3mxYcXOOICwRKepcsQ=";
  subPackages = [ "." ];
}
