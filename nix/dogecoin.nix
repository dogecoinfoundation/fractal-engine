{ lib, fetchurl, autoPatchelfHook, writeShellScriptBin, gettext, gcc-unwrapped, xorg, libxkbcommon, fontconfig, freetype }:

let
  core = pkgs.callPackage (pkgs.fetchurl {
    url = "https://raw.githubusercontent.com/Dogebox-WG/dogebox-nur-packages/f485a380d516436c2310b289622f39bf21382f9c/pkgs/dogecoin-core/default.nix";
    sha256 = "sha256-n27e6ZpRJDKXpL8Fs5QNqELFec5htai33lLY566tHoo=";
  }) {
    disableWallet = true;
    disableGUI = true;
    disableTests = true;
  };

  # Wrapper that starts regtest with your env-driven ports + ZMQ on one socket
  dogecoin-wrapper = writeShellScriptBin "dogecoind" ''
    #!/usr/bin/env bash
    set -euo pipefail

    # Defaults (override via env)
    export RPC_USER=''${RPC_USER:-test}
    export RPC_PASSWORD=''${RPC_PASSWORD:-test}
    export RPC_PORT="''${RPC_PORT:-22556}"
    export INSTANCE_ID="''${INSTANCE_ID:-1}"
    export P2P_PORT="''${P2P_PORT:-18000}"
    export ZMQ_PORT="''${ZMQ_PORT:-28000}"

    CONF="/tmp/dogecoin$INSTANCE_ID.conf"
    DATADIR="/tmp/dogecoin$INSTANCE_ID"

    mkdir -p "$DATADIR"

    # Generate config if you have a template
    if [[ -f "${../regtest.conf}" ]]; then
      ${gettext}/bin/envsubst < ${../regtest.conf} > "$CONF"
      echo "Generated dogecoin.conf:"
      cat "$CONF"
    else
      # Minimal fallback config
      cat > "$CONF" <<EOF
regtest=1
server=1
rpcuser=$RPC_USER
rpcpassword=$RPC_PASSWORD
rpcport=$RPC_PORT
listen=1
port=$P2P_PORT
txindex=1
EOF
    fi

    exec ${core}/bin/dogecoind \
      -printtoconsole \
      -regtest \
      -reindex-chainstate \
      -daemon=0 \
      -splash=0 \
      -acceptnonstdtxn \
      -port="$P2P_PORT" \
      -rpcport="$RPC_PORT" \
      -zmqpubhashtx="tcp://0.0.0.0:$ZMQ_PORT" \
      -zmqpubhashblock="tcp://0.0.0.0:$ZMQ_PORT" \
      -zmqpubrawtx="tcp://0.0.0.0:$ZMQ_PORT" \
      -zmqpubrawblock="tcp://0.0.0.0:$ZMQ_PORT" \
      -datadir="$DATADIR" \
      -conf="$CONF"
  '';
in
  dogecoin-wrapper
