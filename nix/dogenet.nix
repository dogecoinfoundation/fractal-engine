{ lib, buildGoModule, fetchFromGitHub, pkg-config, systemd }:

buildGoModule rec {
  pname = "dogenet";
  version = "main";

  src = fetchFromGitHub {
    owner = "Dogebox-WG";
    repo = "dogenet";
    rev = "main"; # Can be overridden
    sha256 = "sha256-3hAZMqB4YKHonDnZEZLaF3Hb2ajjVkyMrTQjn5eXpKI="; # TODO: Add correct hash
  };

  vendorHash = "sha256-4XDgSVH+QAlIAv5/h30oqeVzMTEoAfEAySkVmMH6kFs="; # TODO: Add correct hash

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ systemd ];

  subPackages = [ "cmd/dogenet" ];

  env.CGO_ENABLED = "1";

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
    cat > $out/bin/dogenet-start << EOF
    #!/usr/bin/env bash

    export DOGE_NET_HANDLER="''${DOGE_NET_HANDLER:-127.0.0.1:42000}"
    export DOGENET_WEB_PORT="''${DOGENET_WEB_PORT:-8085}"
    export DOGENET_BIND_HOST="''${DOGENET_BIND_HOST:-0.0.0.0}"
    export DOGENET_BIND_PORT="''${DOGENET_BIND_PORT:-41000}"
    export INSTANCE_ID="''${INSTANCE_ID:-1}"

    # Set up storage directory
    # Choose a writable DOGENET_HOME

    if [ -z "''${DOGENET_HOME:-}" ]; then
      if [ -n "''${HOME:-}" ] && [ -w "''${HOME:-/}" ]; then
        DOGENET_HOME="''${HOME}/.dogenet''${INSTANCE_ID}"
      else

        DOGENET_HOME="/tmp/.dogenet''${INSTANCE_ID}"
      fi
    fi


    mkdir -p "\$DOGENET_HOME/storage"



    # Copy keys to working directory if they don't exist

    if [ ! -f "\$DOGENET_HOME/dev-key" ]; then
      cp $out/share/dogenet/dev-key "\$DOGENET_HOME/"

    fi
    if [ ! -f "\$DOGENET_HOME/ident-pub" ]; then
      cp $out/share/dogenet/ident-pub "\$DOGENET_HOME/"
    fi


    cd "\$DOGENET_HOME"

    export KEY="\$(cat dev-key)"

    export IDENT="\$(cat ident-pub)"

    echo \$DOGENET_WEB_PORT

    exec $out/bin/dogenet \\
    --local \\
    --public 0.0.0.0 \\
    --handler \$DOGE_NET_HANDLER \\
    --web 0.0.0.0:\$DOGENET_WEB_PORT \\
    --bind \$DOGENET_BIND_HOST:\$DOGENET_BIND_PORT
    EOF

    chmod +x $out/bin/dogenet-start
  '';

  meta = with lib; {
    description = "Dogenet networking service";
    homepage = "https://github.com/Dogebox-WG/dogenet";
    license = licenses.mit;
  };
}
