{
  writeShellScriptBin,
  pkgs,
}:

let
  dogecoin =
    pkgs.callPackage
      (pkgs.fetchurl {
        url = "https://raw.githubusercontent.com/Dogebox-WG/dogebox-nur-packages/92d4675dcb1f0412dee0e53f9c433422abca12da/pkgs/dogecoin-core/default.nix";
        sha256 = "sha256-UQfTBL2XoXqP2ZkYfE+Bocsqr+LuHiQuEKaT3u6evFY=";
      })
      {
        disableWallet = false;
        disableGUI = true;
        disableTests = true;
        enableZMQ = true;
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
    export CHAIN="''${CHAIN:-mainnet}"
    export DATADIR="''${DATADIR:-/tmp/dogecoin$INSTANCE_ID}"

    CONF="/tmp/dogecoin$INSTANCE_ID.conf"

    mkdir -p "$DATADIR"

    touch "$CONF"

    echo "server=1" >> $CONF
    echo "rpcuser=$RPC_USER" >> $CONF
    echo "rpcpassword=$RPC_PASSWORD" >> $CONF
    echo "rpcport=$RPC_PORT" >> $CONF
    echo "listen=1" >> $CONF
    echo "port=$P2P_PORT" >> $CONF
    echo "txindex=1" >> $CONF

    if [ "$CHAIN" = "regtest" ]; then
      echo "regtest=1" >> $CONF
      echo "acceptnonstdtxn=1" >> $CONF
    fi

    if [ "$CHAIN" = "testnet" ]; then
      echo "testnet=1" >> $CONF
    fi

    exec ${dogecoin}/bin/dogecoind \
      -printtoconsole \
      -reindex-chainstate \
      -daemon=0 \
      -splash=0 \
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
