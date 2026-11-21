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
    hash = "sha256-wQItMR7SOyCEt1u3iCbYvb03BhNA0cwSPrdZ4Zg46ZI=";
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
