#!/usr/bin/env bash
# Dogecoin EC2 bootstrap script for UserData
# - Formats and mounts EBS volume at /var/lib/dogecoin
# - Installs Nix
# - Builds a Dogecoin wrapper via Nix
# - Installs a systemd unit to run dogecoind on boot
set -euxo pipefail

# --- Configuration / Defaults ---
DEVICE_NAME="${DEVICE_NAME:-/dev/xvdb}"
MOUNT_POINT="${MOUNT_POINT:-/var/lib/dogecoin}"

RPC_USER="${RPC_USER:-test}"
RPC_PASSWORD="${RPC_PASSWORD:-test}"
RPC_PORT="${RPC_PORT:-22555}"     # Dogecoin mainnet RPC default
P2P_PORT="${P2P_PORT:-22556}"     # Dogecoin mainnet P2P default
ZMQ_PORT="${ZMQ_PORT:-28000}"
CHAIN="${CHAIN:-mainnet}"

# --- Filesystem setup for EBS volume ---
# Create filesystem if not present
if ! file -s "${DEVICE_NAME}" | grep -q filesystem; then
  mkfs -t xfs "${DEVICE_NAME}"
fi

mkdir -p "${MOUNT_POINT}"

# Add to fstab if not already present
if ! grep -q -E "^${DEVICE_NAME} ${MOUNT_POINT} " /etc/fstab; then
  echo "${DEVICE_NAME} ${MOUNT_POINT} xfs defaults,nofail 0 2" >> /etc/fstab
fi

# Mount all
mount -a

# --- Basic tooling required for Nix install ---
dnf install -y curl xz tar git ca-certificates

# --- Install Nix (single-user) ---
if [ ! -d "/root/.nix-profile" ]; then
  curl -L https://nixos.org/nix/install | sh -s -- --no-daemon
fi

# Load nix environment for this shell
# shellcheck disable=SC1091
source /root/.nix-profile/etc/profile.d/nix.sh || true
export PATH="/root/.nix-profile/bin:$PATH"

# --- Nix expressions are provided via UserData S3 assets ---
install -d -m 0755 /opt/dogecoin-nix
if [ ! -f /opt/dogecoin-nix/dogecoin.nix ]; then
  echo "Expected dogecoin.nix to exist; ensure CDK assets are downloaded in UserData before running bootstrap.sh"
  exit 1
fi

# --- Build and install the wrapper into root profile ---
nix-env -if /opt/dogecoin-nix

# --- Create systemd unit for dogecoind ---
cat > /etc/systemd/system/dogecoind.service <<EOF
[Unit]
Description=Dogecoin node via Nix wrapper
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=root
Environment="CHAIN=${CHAIN}"
Environment="RPC_USER=${RPC_USER}"
Environment="RPC_PASSWORD=${RPC_PASSWORD}"
Environment="RPC_PORT=${RPC_PORT}"
Environment="P2P_PORT=${P2P_PORT}"
Environment="ZMQ_PORT=${ZMQ_PORT}"
Environment="DATADIR=${MOUNT_POINT}"
ExecStart=/root/.nix-profile/bin/dogecoind
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# --- Enable and start ---
systemctl daemon-reload
systemctl enable --now dogecoind

echo "Dogecoin bootstrap complete."
