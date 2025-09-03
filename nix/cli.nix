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
  pname = "fractal-cli";
  version = releaseVersion;

  src = ../.;

  vendorHash = "sha256-Ll0T8pgW4Fj/hWukEuVynfiUwhFNcPZqmiJ3l5QkZ4Q=";

  nativeBuildInputs = [ pkgs.pkg-config ];
  buildInputs = [ ];

  # Build the CLI binary
  subPackages = [ "cmd/cli" ];

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
    description = "Fractal Engine CLI";
    homepage = "https://github.com/dogecoinfoundation/fractal-engine";
    license = licenses.mit;
    maintainers = [ ];
  };
}
