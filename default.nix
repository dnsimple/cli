{
  lib,
  buildGoModule,
  version ? "dev",
}:

buildGoModule {
  pname = "dnsimple";
  inherit version;

  src = lib.cleanSource ./.;

  vendorHash = "sha256-4dOKR9BYsmL083eCBpZzXRPr3CHBYF4Kc+HMo6v9MU0=";

  ldflags = [
    "-s"
    "-w"
    "-X main.version=${version}"
  ];

  subPackages = [ "cmd/dnsimple" ];

  meta = with lib; {
    description = "Command-line interface for the DNSimple API";
    homepage = "https://github.com/dnsimple/dnsimple-cli";
    license = licenses.mit;
    maintainers = [ "DNSimple" ];
    mainProgram = "dnsimple";
  };
}
