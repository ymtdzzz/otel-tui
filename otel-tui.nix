{
  config,
  pkgs,
  fetchFromGitHub,
  ...
}:

let
  otel-tui-version = "v0.4.11";
in
pkgs.buildGoModule {
  pname = "otel-tui";
  version = "${otel-tui-version}";
  src = pkgs.fetchFromGitHub {
    owner = "ymtdzzz";
    repo = "otel-tui";
    rev = "${otel-tui-version}";
    hash = "sha256-6OyKO37OqvJDjrr7CzTwMXWWe/tqpap6EZJBM7qf4J0=";
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
  vendorHash = "sha256-Mbhgc5cpoKTamPAuY8CRGoplM6N2/GzBumMC2/PMn2Y=";
  subPackages = [ "." ];
  buildInputs = pkgs.lib.optionals pkgs.stdenv.isLinux [
    pkgs.xorg.libX11
  ];
}
