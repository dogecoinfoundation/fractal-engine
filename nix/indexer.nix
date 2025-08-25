{
  lib,
  buildGoModule,
  fetchFromGitHub,
  pkg-config,
  zeromq,
}:

buildGoModule rec {
  pname = "indexer";
  version = "main";

  src = fetchFromGitHub {
    owner = "dogeorg";
    repo = "indexer";
    rev = "v0.0.4";
    sha256 = "sha256-CVwZLwiE83h8SbkW+EMzymuTyziNOAzA82q59Qhsx20=";
  };

  vendorHash = "sha256-EpogYqHjdxiXK9WgpR/3P86BvlvmDuuGFvMrRpkubH0=";

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ zeromq ];

  subPackages = [ "." ];

  env.CGO_ENABLED = "1";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = with lib; {
    description = "Dogecoin indexer service";
    homepage = "https://github.com/dogeorg/indexer";
    license = licenses.mit;
  };
}
