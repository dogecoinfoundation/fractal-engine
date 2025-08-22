{
  lib,
  buildGoModule,
  pkg-config,
  systemd,
}:

buildGoModule rec {
  pname = "fractalengine";
  version = "latest";

  src = lib.cleanSource ../.;

  vendorHash = "sha256-VSxnayMOT+ctF27IOH4okWD5pUisANJn77Ksb2yXR6I=";

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ systemd ];

  # Build the main binary
  subPackages = [ "cmd/fractal-engine" ];

  # Set build flags
  ldflags = [
    "-s"
    "-w"
  ];

  # Environment variables for build
  env.CGO_ENABLED = "1";

  # Copy migrations after build
  postInstall = ''
    mkdir -p $out/share/fractalengine
    if [ -d $src/db/migrations ]; then
      cp -r $src/db/migrations $out/share/fractalengine/
    fi
  '';

  meta = with lib; {
    description = "Fractal Engine - Core Dogecoin service";
    homepage = "https://github.com/dogecoinfoundation/fractal-engine";
    license = licenses.mit;
    maintainers = [ ];
  };
}
