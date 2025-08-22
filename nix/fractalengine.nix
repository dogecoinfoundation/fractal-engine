{
  lib,
  buildGoModule,
  pkg-config,
  rev,
  date,
}:

let
  releaseVersion = "0.0.1";
in
buildGoModule rec {
  pname = "fractal-engine";
  version = releaseVersion;

  src = ../.;

  vendorHash = "sha256-VSxnayMOT+ctF27IOH4okWD5pUisANJn77Ksb2yXR6I=";

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ ];

  # Build the main binary
  subPackages = [ "cmd/fractal-engine" ];

  # Set build flags
  ldflags = [
    "-s"
    "-w"
    "-X"
    "dogecoin.org/fractal-engine/pkg/version.Version=${releaseVersion}"
    "-X"
    "dogecoin.org/fractal-engine/pkg/version.Commit=${rev}"
    "-X"
    "dogecoin.org/fractal-engine/pkg/version.Date=${date}"
  ];

  # Environment variables for build
  env.CGO_ENABLED = "1";

  meta = with lib; {
    description = "Fractal Engine - Core Dogecoin service";
    homepage = "https://github.com/dogecoinfoundation/fractal-engine";
    license = licenses.mit;
    maintainers = [ ];
  };
}
