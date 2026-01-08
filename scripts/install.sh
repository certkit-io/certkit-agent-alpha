#!/usr/bin/env bash
set -euo pipefail

if [[ $EUID -ne 0 ]]; then
  echo "Please run as root (sudo ...)" >&2
  exit 1
fi

: "${ACCESS_KEY:?ACCESS_KEY is required}"
: "${SECRET_KEY:?SECRET_KEY is required}"

OWNER="certkit-io"
REPO="certkit-agent-alpha"

BIN_NAME="certkit-agent"
INSTALL_DIR="/usr/local/bin"
ETC_DIR="/etc/certkit-agent"
CONFIG_FILE="$ETC_DIR/config.json"

TAG="${VERSION:-v0.0.1}"
API_BASE="${CERTKIT_API_BASE:-https://app.certkit.io}"

# Detect architecture
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64)  arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *)
    echo "Unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

ASSET_BIN="${BIN_NAME}_${TAG}_linux_${arch}"
ASSET_SHA="${BIN_NAME}_${TAG}_SHA256SUMS.txt"

BASE_URL="https://github.com/${OWNER}/${REPO}/releases/download/${TAG}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading ${ASSET_BIN}"
curl -fsSL "${BASE_URL}/${ASSET_BIN}" -o "$tmp/${ASSET_BIN}"

echo "Downloading ${ASSET_SHA}"
curl -fsSL "${BASE_URL}/${ASSET_SHA}" -o "$tmp/${ASSET_SHA}"

echo "Verifying checksum"
(
  cd "$tmp"
  grep -E "^[a-f0-9]{64}[[:space:]]+${ASSET_BIN}\$" "${ASSET_SHA}" | sha256sum -c -
)

echo "Installing binary to ${INSTALL_DIR}/${BIN_NAME}"
install -m 0755 "$tmp/${ASSET_BIN}" "${INSTALL_DIR}/${BIN_NAME}"

echo "Writing config to ${CONFIG_FILE}"
mkdir -p "$ETC_DIR"
chmod 0755 "$ETC_DIR"

umask 0077
cat > "$CONFIG_FILE" <<EOF
{
  "api_base": "${API_BASE}",
  "bootstrap": {
    "access_key": "${ACCESS_KEY}",
    "secret_key": "${SECRET_KEY}"
  },
  "agent": null,
  "desired_state": null
}
EOF
chmod 0600 "$CONFIG_FILE"

echo "Installing systemd service via certkit-agent install"
/usr/local/bin/${BIN_NAME} install --config "${CONFIG_FILE}"

echo "✅ CertKit Agent installed and started"
echo "   systemctl status certkit-agent.service"
