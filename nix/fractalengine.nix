{
  lib,
  pkgs,
  rev,
  date,
}:

let
  releaseVersion = "0.0.1";
  musl = pkgs.stdenv.hostPlatform.isMusl;
in
pkgs.buildGo124Module rec {
  pname = "fractal-engine";
  version = releaseVersion;

  src = ../.;

  vendorHash = "sha256-cEazt2Cq7D7JBFDf3oUW5dQhmMzqFNuccDde7RYJDZs=";

  nativeBuildInputs = [ pkgs.pkg-config ];
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
  ]
  ++ lib.optionals musl [
    "-linkmode=external"
    "-extldflags=-static"
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
