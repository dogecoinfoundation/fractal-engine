{ lib, buildGoModule, fetchFromGitHub, pkg-config, systemd }:

buildGoModule rec {
  pname = "dogenet";
  version = "main";

  src = fetchFromGitHub {
    owner = "Dogebox-WG";
    repo = "dogenet";
    rev = "main"; # Can be overridden
    sha256 = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # TODO: Add correct hash
  };

  vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # TODO: Add correct hash

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ systemd ];

  subPackages = [ "cmd/dogenet" ];

  CGO_ENABLED = "1";

  ldflags = [
    "-s" "-w"
  ];

  # Post-build setup to generate keys
  postInstall = ''
    mkdir -p $out/share/dogenet
    
    # Generate development keys
    cd $out/share/dogenet
    $out/bin/dogenet genkey dev-key
    $out/bin/dogenet genkey ident-key ident-pub
    
    # Create wrapper script
    cat > $out/bin/dogenet-start << 'EOF'
    #!/usr/bin/env bash
    
    # Default environment variables
    export DOGE_NET_HANDLER=''${DOGE_NET_HANDLER:-unix:///tmp/dogenet.sock}
    export DOGENET_WEB_PORT=''${DOGENET_WEB_PORT:-8085}
    export DOGENET_BIND_HOST=''${DOGENET_BIND_HOST:-0.0.0.0}
    export DOGENET_BIND_PORT=''${DOGENET_BIND_PORT:-42000}
    
    # Set up storage directory
    mkdir -p $HOME/.dogenet/storage
    
    cd $out/share/dogenet
    export KEY=$(cat dev-key)
    export IDENT=$(cat ident-pub)
    
    exec $out/bin/dogenet \
      --local \
      --public 0.0.0.0 \
      --handler $DOGE_NET_HANDLER \
      --web 0.0.0.0:$DOGENET_WEB_PORT \
      --bind $DOGENET_BIND_HOST:$DOGENET_BIND_PORT
    EOF
    
    chmod +x $out/bin/dogenet-start
  '';

  meta = with lib; {
    description = "Dogenet networking service";
    homepage = "https://github.com/Dogebox-WG/dogenet";
    license = licenses.mit;
  };
}
