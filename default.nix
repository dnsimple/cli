{
  pkgs ? import <nixpkgs> { },
  lib ? pkgs.lib,
  buildGoModule ? pkgs.buildGoModule,
  version ? "dev",
}:

buildGoModule {
  pname = "dnsimple";
  inherit version;

  src = lib.cleanSource ./.;

  vendorHash = "sha256-9lVLlPokN+tIaKHgVhaCGzsnlpmjgLunBEoPhRZkrVU=";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];

  subPackages = [ "cmd/dnsimple" ];

  meta = {
    description = "Command-line interface for the DNSimple API";
    homepage = "https://github.com/dnsimple/cli";
    license = lib.licenses.mit;
    maintainers = [ "DNSimple" ];
    mainProgram = "dnsimple";
  };
}
