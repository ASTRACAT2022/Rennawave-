#!/usr/bin/env bash
set -Eeuo pipefail

REPOSITORY='ASTRACAT2022/Rennawave-'
REF="${REF:-main}"
INSTALL_DIR="${INSTALL_DIR:-/opt/rennawave-test}"
TEST_PORT="${TEST_PORT:-80}"

die() {
    echo "Error: $*" >&2
    exit 1
}

[[ "$(id -u)" -eq 0 ]] || die 'Run as root: curl ... | sudo bash'
command -v docker >/dev/null || die 'Docker is required.'
docker compose version >/dev/null 2>&1 || die 'Docker Compose v2 is required.'

if [[ -e "$INSTALL_DIR" ]]; then
    die "$INSTALL_DIR already exists. Choose another INSTALL_DIR or remove it deliberately."
fi

if ! [[ "$TEST_PORT" =~ ^[0-9]+$ ]] || (( TEST_PORT < 1 || TEST_PORT > 65535 )); then
    die 'TEST_PORT must be between 1 and 65535.'
fi

umask 077
mkdir -p "$INSTALL_DIR"
trap 'rm -rf "${ARCHIVE:-}" "${EXTRACT_DIR:-}"' EXIT

ARCHIVE="$(mktemp)"
EXTRACT_DIR="$(mktemp -d)"

echo 'Downloading Rennawave sources...'
curl --fail --location --silent --show-error \
    "https://github.com/${REPOSITORY}/archive/refs/heads/${REF}.tar.gz" \
    --output "$ARCHIVE"
tar -xzf "$ARCHIVE" -C "$EXTRACT_DIR" --strip-components=1
cp -a "$EXTRACT_DIR/." "$INSTALL_DIR/"

secret() {
    tr -dc 'A-Za-z0-9' </dev/urandom | head -c 64
}

db_password="$(secret)"
jwt_auth_secret="$(secret)"
jwt_api_secret="$(secret)"

cp "$INSTALL_DIR/backend/.env.sample" "$INSTALL_DIR/.env"
sed -i \
    -e "s|^DATABASE_URL=.*|DATABASE_URL=\"postgresql://postgres:${db_password}@db:5432/postgres\"|" \
    -e "s|^JWT_AUTH_SECRET=.*|JWT_AUTH_SECRET=${jwt_auth_secret}|" \
    -e "s|^JWT_API_TOKENS_SECRET=.*|JWT_API_TOKENS_SECRET=${jwt_api_secret}|" \
    -e 's|^PANEL_DOMAIN=.*|PANEL_DOMAIN=localhost|' \
    -e 's|^SUB_PUBLIC_DOMAIN=.*|SUB_PUBLIC_DOMAIN=localhost/api/sub|' \
    -e "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD=${db_password}|" \
    "$INSTALL_DIR/.env"

echo "Building and starting the test panel on port ${TEST_PORT}..."
cd "$INSTALL_DIR"
TEST_PORT="$TEST_PORT" docker compose -f docker-compose.test.yml up -d --build

host_ip="$(hostname -I 2>/dev/null | awk '{print $1}' || true)"

echo
echo "Ready: http://${host_ip:-localhost}:${TEST_PORT}"
echo "Logs:  cd ${INSTALL_DIR} && docker compose -f docker-compose.test.yml logs -f api"
echo "Stop:  cd ${INSTALL_DIR} && docker compose -f docker-compose.test.yml down"
