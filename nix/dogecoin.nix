{ lib
, stdenv
, fetchurl
, autoPatchelfHook
, writeShellScriptBin
, gettext
, pkgs
}:

let
  dogecoin = stdenv.mkDerivation rec {
    pname = "dogecoin";
    version = "1.14.9";

    src = fetchurl {
      url = "https://github.com/dogecoin/dogecoin/releases/download/v${version}/dogecoin-${version}-x86_64-linux-gnu.tar.gz";
      sha256 = "sha256-TyJxF7QRp8mGIslwmG4nvPw/VHpyvvZefZ6CmJF11Pg=";
    };

    # We're installing prebuilt binaries → NO autoreconf/pkg-config
    dontConfigure = true;
    dontBuild = true;

    # Patchelf the downloaded binaries so they find Nix-provided libs
    nativeBuildInputs = [ autoPatchelfHook ];

    # Runtime deps typically needed by the prebuilt dogecoin binaries
    buildInputs = [
      pkgs.stdenv.cc.cc.lib   # libstdc++
      pkgs.openssl
      pkgs.libevent
      pkgs.zeromq             # enables ZMQ support at runtime
      pkgs.zlib
    ];

    # The tarball layout is dogecoin-1.14.9/bin/...
    # Let Nix know where to find it after unpack
    installPhase = ''
      runHook preInstall
      mkdir -p $out/bin
      # find the bin dir robustly (versioned folder)
      BIN_DIR="$(find . -maxdepth 2 -type d -name bin | head -n1)"
      # copy only headless binaries to avoid Qt/X11 deps
      for f in dogecoind dogecoin-cli dogecoin-tx; do
        if [ -f "$BIN_DIR/$f" ]; then
          cp "$BIN_DIR/$f" $out/bin/
        fi
      done
      runHook postInstall
    '';

    meta = with lib; {
      description = "Dogecoin Core (prebuilt binary)";
      homepage = "https://dogecoin.com/";
      license = licenses.mit;
      platforms = [ "x86_64-linux" ];
    };
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

    exec ${dogecoin}/bin/dogecoind \
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
