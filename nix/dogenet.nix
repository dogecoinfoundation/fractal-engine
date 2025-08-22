{ pkgs, ... }:

let
  dogenet_upstream = pkgs.callPackage (pkgs.fetchurl {
    url = "https://raw.githubusercontent.com/dogebox-wg/dogebox-nur-packages/0ea16a94f4b7bbf3eedc2ab1c351e6f366bca23b/pkgs/dogenet/default.nix";
    sha256 = "sha256-JUWuZxDO30H7tNeAu4roBmUHhZHsGs072Ex8QD9Lzz4=";
  }) {};
  dogenet = pkgs.writeShellScriptBin "dogenet" ''
set -euo pipefail

# Runtime-configurable env with defaults
export DOGE_NET_HANDLER="''${DOGE_NET_HANDLER:-127.0.0.1:42000}"
export DOGENET_WEB_PORT="''${DOGENET_WEB_PORT:-8085}"
export DOGENET_BIND_HOST="''${DOGENET_BIND_HOST:-0.0.0.0}"
export DOGENET_BIND_PORT="''${DOGENET_BIND_PORT:-41000}"
export INSTANCE_ID="''${INSTANCE_ID:-1}"

# Derive public from bind unless explicitly overridden
export DOGENET_PUBLIC_HOST="''${DOGENET_PUBLIC_HOST:-''${DOGENET_BIND_HOST}}"
export DOGENET_PUBLIC_PORT="''${DOGENET_PUBLIC_PORT:-''${DOGENET_BIND_PORT}}"

DOGENET_HOME="$HOME/.fractal-stack-$INSTANCE_ID/dogenet"

mkdir -p "''${DOGENET_HOME}/storage"

# Generate keys if missing
if [ ! -f "''${DOGENET_HOME}/dev-key" ]; then
  ${dogenet_upstream}/bin/dogenet genkey "''${DOGENET_HOME}/dev-key"
fi
if [ ! -f "''${DOGENET_HOME}/ident-pub" ]; then
  ${dogenet_upstream}/bin/dogenet genkey "''${DOGENET_HOME}/ident-key" "''${DOGENET_HOME}/ident-pub"
fi

cd "''${DOGENET_HOME}"

# Load keys into env (dogenet reads KEY)
export KEY="$(cat dev-key)"
export IDENT="$(cat ident-pub)"

exec ${dogenet_upstream}/bin/dogenet \
  --local \
  --public "''${DOGENET_PUBLIC_HOST}:''${DOGENET_PUBLIC_PORT}" \
  --handler "''${DOGE_NET_HANDLER}" \
  --web "0.0.0.0:''${DOGENET_WEB_PORT}" \
  --bind "''${DOGENET_BIND_HOST}:''${DOGENET_BIND_PORT}"
'';
in dogenet