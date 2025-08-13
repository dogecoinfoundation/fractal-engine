{ lib, buildGoModule, fetchFromGitHub, pkg-config, systemd, zeromq }:

buildGoModule rec {
  pname = "indexer";
  version = "main";

  src = fetchFromGitHub {
    owner = "dogeorg";
    repo = "indexer";
    rev = "main"; # Can be overridden
    sha256 = "sha256-piDP8o4qFJkKOgU0S4uR9unMSpW02toi5IyKkJ1VlUY=";
  };

  vendorHash = "sha256-2JU4LayU5mLZXyGH1rt9tduQgqUycyfMNSendcCNLhw=";

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ systemd zeromq ];

  # Build main.go in root
  subPackages = [ "." ];

  env.CGO_ENABLED = "1";

  ldflags = [
    "-s" "-w"
  ];

  postInstall = ''
    # Create wrapper script with new CLI flags
    cat > $out/bin/indexer-start << 'EOF'
    #!/usr/bin/env bash

    # Default environment variables
    export INDEXER_BINDAPI=''${INDEXER_BINDAPI:-localhost:8888}
    export INDEXER_CHAIN=''${INDEXER_CHAIN:-regtest}
    export INDEXER_DBURL=''${INDEXER_DBURL:-index.db}
    export INDEXER_LISTENPORT=''${INDEXER_LISTENPORT:-8001}
    export INDEXER_RPCHOST=''${INDEXER_RPCHOST:-127.0.0.1}
    export INDEXER_RPCPASS=''${INDEXER_RPCPASS:-dogecoin}
    export INDEXER_RPCPORT=''${INDEXER_RPCPORT:-22555}
    export INDEXER_RPCUSER=''${INDEXER_RPCUSER:-dogecoin}
    export INDEXER_STARTINGHEIGHT=''${INDEXER_STARTINGHEIGHT:-5830000}
    export INDEXER_WEBPORT=''${INDEXER_WEBPORT:-8000}
    export INDEXER_ZMQHOST=''${INDEXER_ZMQHOST:-127.0.0.1}
    export INDEXER_ZMQPORT=''${INDEXER_ZMQPORT:-28332}

    # Create storage directory
    mkdir -p $HOME/.indexer/storage

    exec $out/bin/indexer \
      -bindapi $INDEXER_BINDAPI \
      -chain $INDEXER_CHAIN \
      -dburl $INDEXER_DBURL \
      -listenport $INDEXER_LISTENPORT \
      -rpchost $INDEXER_RPCHOST \
      -rpcpass $INDEXER_RPCPASS \
      -rpcport $INDEXER_RPCPORT \
      -rpcuser $INDEXER_RPCUSER \
      -startingheight $INDEXER_STARTINGHEIGHT \
      -webport $INDEXER_WEBPORT \
      -zmqhost $INDEXER_ZMQHOST \
      -zmqport $INDEXER_ZMQPORT
    EOF

    chmod +x $out/bin/indexer-start
  '';

  meta = with lib; {
    description = "Dogecoin indexer service";
    homepage = "https://github.com/dogeorg/indexer";
    license = licenses.mit;
  };
}
