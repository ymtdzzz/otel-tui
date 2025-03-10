{
  config,
  pkgs,
  fetchFromGitHub,
  ...
}:

let
  otel-tui-version = "v0.4.3-pre5";
in
pkgs.buildGoModule {
  pname = "otel-tui";
  version = "${otel-tui-version}";
  src = pkgs.fetchFromGitHub {
    owner = "ymtdzzz";
    repo = "otel-tui";
    rev = "${otel-tui-version}";
    hash = "sha256-YO6FLppOsrF40u/uqkkE8eq2q9N8WspUiw35REBijzM=";
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
  vendorHash = "sha256-Kfru+SmcjlBB5ylViQceaTXATJHyVD6Kv/Uyy68D2cE=";
  subPackages = [ "." ];
  buildInputs = pkgs.lib.optionals pkgs.stdenv.isLinux [
    pkgs.xorg.libX11
  ];
}
