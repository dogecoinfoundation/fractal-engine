{ lib, stdenv, fetchurl, autoPatchelfHook, writeShellScriptBin, gettext, gcc-unwrapped, xorg, libxkbcommon, fontconfig, freetype }:

let
  dogecoin = stdenv.mkDerivation rec {
    pname = "dogecoin";
    version = "1.14.9";

    src = fetchurl {
      url = "https://github.com/dogecoin/dogecoin/releases/download/v${version}/dogecoin-${version}-x86_64-linux-gnu.tar.gz";
      sha256 = "sha256-TyJxF7QRp8mGIslwmG4nvPw/VHpyvvZefZ6CmJF11Pg=";
    };

    nativeBuildInputs = [ autoPatchelfHook ];
    buildInputs = [
      gcc-unwrapped.lib
      xorg.libxcb
      libxkbcommon
      fontconfig
      freetype
    ];

    installPhase = ''
      runHook preInstall

      mkdir -p $out/bin
      cp bin/* $out/bin/

      runHook postInstall
    '';

    meta = with lib; {
      description = "Dogecoin Core";
      homepage = "https://dogecoin.com/";
      license = licenses.mit;
      platforms = [ "x86_64-linux" ];
    };
  };

  # Wrapper script that mimics the Docker behavior
  dogecoin-wrapper = writeShellScriptBin "dogecoind" ''
    #!/usr/bin/env bash

    # Default environment variables
    export RPC_USER=''${RPC_USER:-test}
    export RPC_PASSWORD=''${RPC_PASSWORD:-test}
    export RPC_PORT=''${RPC_PORT:-22556}
    export INSTANCE_ID=''${INSTANCE_ID:-1}
    export P2P_PORT=''${P2P_PORT:-18000}

    # Generate config from template
    ${gettext}/bin/envsubst < ${../regtest.conf} > /tmp/dogecoin$INSTANCE_ID.conf

    echo "Generated dogecoin.conf:"
    cat /tmp/dogecoin$INSTANCE_ID.conf

    mkdir -p /tmp/dogecoin$INSTANCE_ID

    # Start dogecoind
    exec ${dogecoin}/bin/dogecoind \
      -printtoconsole \
      -regtest \
      -reindex-chainstate \
      -min \
      -splash=0 \
      -port=$P2P_PORT \
      -datadir=/tmp/dogecoin$INSTANCE_ID/ \
      -conf=/tmp/dogecoin$INSTANCE_ID.conf
  '';

in dogecoin-wrapper
