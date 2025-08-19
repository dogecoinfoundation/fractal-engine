{ lib, buildGoModule, fetchFromGitHub, pkg-config, systemd }:

buildGoModule rec {
  pname = "dogenet";
  version = "main";

  src = fetchFromGitHub {
    owner = "Dogebox-WG";
    repo = "dogenet";
    rev = "main"; # Can be overridden
    sha256 = "sha256-3hAZMqB4YKHonDnZEZLaF3Hb2ajjVkyMrTQjn5eXpKI="; # TODO: Update with the correct hash
  };

  vendorHash = "sha256-4XDgSVH+QAlIAv5/h30oqeVzMTEoAfEAySkVmMH6kFs="; # TODO: Update with the correct hash

  nativeBuildInputs = [ pkg-config ];
  buildInputs = [ systemd ];

  subPackages = [ "cmd/dogenet" ];

  env.CGO_ENABLED = "1";

  ldflags = [
    "-s" "-w"
  ];

  # Post-build setup to generate keys and wrapper
  postInstall = ''
    mkdir -p $out/share/dogenet

    # Generate development keys
    cd $out/share/dogenet
    $out/bin/dogenet genkey dev-key
    $out/bin/dogenet genkey ident-key ident-pub

    # Create wrapper script with runtime env defaults and public host:port derivation
    cat > $out/bin/dogenet-start <<'EOF'
#!/usr/bin/env bash
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

# Determine DOGENET_HOME directory
if [ -z "''${DOGENET_HOME:-}" ]; then
  if [ -n "''${HOME:-}" ] && [ -w "''${HOME:-/}" ] ; then
    DOGENET_HOME="''${HOME}/.dogenet''${INSTANCE_ID}"
  else
    DOGENET_HOME="/tmp/.dogenet''${INSTANCE_ID}"
  fi
fi
mkdir -p "''${DOGENET_HOME}/storage"

# Seed keys if missing
if [ ! -f "''${DOGENET_HOME}/dev-key" ]; then
  cp "__OUT_PATH__/share/dogenet/dev-key" "''${DOGENET_HOME}/"
fi
if [ ! -f "''${DOGENET_HOME}/ident-pub" ]; then
  cp "__OUT_PATH__/share/dogenet/ident-pub" "''${DOGENET_HOME}/"
fi

cd "''${DOGENET_HOME}"

# Load keys into env (dogenet reads KEY)
export KEY="$(cat dev-key)"
export IDENT="$(cat ident-pub)"

exec "__OUT_PATH__/bin/dogenet" \
  --local \
  --public "''${DOGENET_PUBLIC_HOST}:''${DOGENET_PUBLIC_PORT}" \
  --handler "''${DOGE_NET_HANDLER}" \
  --web "0.0.0.0:''${DOGENET_WEB_PORT}" \
  --bind "''${DOGENET_BIND_HOST}:''${DOGENET_BIND_PORT}"
EOF

    # Inject the actual store path into the wrapper (while keeping runtime env expansion intact)
    sed -i "s#__OUT_PATH__#$out#g" "$out/bin/dogenet-start"

    chmod +x $out/bin/dogenet-start
  '';

  meta = with lib; {
    description = "Dogenet networking service";
    homepage = "https://github.com/Dogebox-WG/dogenet";
    license = licenses.mit;
    mainProgram = "dogenet-start";
  };
}
