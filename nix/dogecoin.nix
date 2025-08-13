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
  dogecoin-wrapper = writeShellScriptBin "dogecoin-regtest" ''
    #!/usr/bin/env bash

    # Default environment variables
    export RPC_USER=''${RPC_USER:-test}
    export RPC_PASSWORD=''${RPC_PASSWORD:-test}
    export RPC_PORT=''${RPC_PORT:-22556}

    # Generate config from template
    ${gettext}/bin/envsubst < ${../regtest.conf} > /tmp/dogecoin.conf

    echo "Generated dogecoin.conf:"
    cat /tmp/dogecoin.conf

    # Start dogecoind
    exec ${dogecoin}/bin/dogecoind \
      -printtoconsole \
      -regtest \
      -reindex-chainstate \
      -min \
      -splash=0 \
      -conf=/tmp/dogecoin.conf
  '';

in dogecoin-wrapper
