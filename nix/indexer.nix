{ lib, buildGoModule, fetchFromGitHub, pkg-config, systemd, zeromq }:

buildGoModule rec {
  pname = "indexer";
  version = "main";

  src = fetchFromGitHub {
    owner = "dogeorg";
    repo = "indexer";
    rev = "main"; # Can be overridden
    sha256 = "sha256-Rps6v1dsBzlFrtWRt7TbtmdpyfNLVVLCK12kBP4j3ao="; # TODO: Add correct hash
  };

  vendorHash = "sha256-1NcilX79v5EZvtH1KQv04sqQ347dp0DBk9FtiKUQ1Uw="; # TODO: Add correct hash

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ systemd zeromq ];

  # Build main.go in root
  subPackages = [ "." ];

  env.CGO_ENABLED = "1";

  ldflags = [
    "-s" "-w"
  ];

  postInstall = ''
    # Create wrapper script (note: Dockerfile had wrong command, fixing it)
    cat > $out/bin/indexer-start << 'EOF'
    #!/usr/bin/env bash

    # Default environment variables
    export RPC_SERVER_PORT=''${RPC_SERVER_PORT:-8893}
    export DOGE_PORT=''${DOGE_PORT:-22556}
    export DOGE_HOST=''${DOGE_HOST:-localhost}

    # Create storage directory
    mkdir -p $HOME/.indexer/storage

    exec $out/bin/indexer \
      --doge-host $DOGE_HOST \
      --doge-port $DOGE_PORT \
      --rpc-server-host 0.0.0.0 \
      --rpc-server-port $RPC_SERVER_PORT
    EOF

    chmod +x $out/bin/indexer-start
  '';

  meta = with lib; {
    description = "Dogecoin indexer service";
    homepage = "https://github.com/dogeorg/indexer";
    license = licenses.mit;
  };
}
