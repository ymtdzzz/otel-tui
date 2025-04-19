{
  config,
  pkgs,
  fetchFromGitHub,
  ...
}:

let
  otel-tui-version = "v0.4.10";
in
pkgs.buildGoModule {
  pname = "otel-tui";
  version = "${otel-tui-version}";
  src = pkgs.fetchFromGitHub {
    owner = "ymtdzzz";
    repo = "otel-tui";
    rev = "${otel-tui-version}";
    hash = "sha256-7Lxl9DEBnMDvk+ZAAaYQ6MuYmqEyl7TG3nifFVqSuEo=";
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
  vendorHash = "sha256-131lg7SdvpI03aVnDX7cFZSlMDr7XFcItw3mQFtmnXw=";
  subPackages = [ "." ];
  buildInputs = pkgs.lib.optionals pkgs.stdenv.isLinux [
    pkgs.xorg.libX11
  ];
}
