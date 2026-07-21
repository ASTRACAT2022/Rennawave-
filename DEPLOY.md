# Deployment with GitHub Actions and Docker

The workflow builds one image containing this repository's frontend and backend:

```text
ghcr.io/astracat2022/rennawave-aesingflow:latest
```

It runs automatically for every push to `main` that changes `backend`, `frontend`,
or the Dockerfile. A second immutable image is published as `sha-<commit>`.

## 1. Enable and inspect the build

Open **Actions → Build unified panel image** in GitHub and wait for the run to
finish. In **Packages**, change the `rennawave-aesingflow` container package visibility to
Public; otherwise the server must authenticate to GHCR before pulling it.

## 2. Prepare the server

Install Docker Engine with the Docker Compose v2 plugin. Then create a directory
for deployment and download the compose file and example environment:

```bash
sudo mkdir -p /opt/rennawave
cd /opt/rennawave

sudo curl -fsSLO https://raw.githubusercontent.com/ASTRACAT2022/Rennawave-/main/docker-compose.fork.yml
sudo curl -fsSL https://raw.githubusercontent.com/ASTRACAT2022/Rennawave-/main/backend/.env.sample -o .env
sudo chown "$USER":"$USER" .env docker-compose.fork.yml
```

Edit `.env`. At minimum set a unique database password and two 64-character
secrets, then set the public domain values:

```dotenv
POSTGRES_PASSWORD=replace-with-a-strong-password
DATABASE_URL="postgresql://postgres:replace-with-a-strong-password@remnawave-db:5432/postgres"
JWT_AUTH_SECRET=64-random-characters
JWT_API_TOKENS_SECRET=another-64-random-characters
PANEL_DOMAIN=panel.example.com
FRONT_END_DOMAIN=https://panel.example.com
SUB_PUBLIC_DOMAIN=panel.example.com/api/sub
```

Generate a 64-character secret with `openssl rand -hex 32`.

## 3. Start and update

```bash
cd /opt/rennawave
docker compose -f docker-compose.fork.yml pull
docker compose -f docker-compose.fork.yml up -d
docker compose -f docker-compose.fork.yml ps
```

The API listens only on `127.0.0.1:3000`. Put Nginx or Caddy with HTTPS in
front of it. PostgreSQL is intentionally available only on `127.0.0.1:6767`.

To roll out a newer GitHub Actions image, repeat `pull` and `up -d`. To pin an
exact build, add the following line to `.env` before the commands above:

```dotenv
REMNAWAVE_IMAGE=ghcr.io/astracat2022/rennawave-aesingflow:sha-<commit-sha>
```

## AesingFlow Node

This panel image validates AesingFlow configuration. A Node is separate and
requires an AesingFlow-compatible Xray core via `CUSTOM_CORE_URL`; the standard
Remnawave Node image does not provide that protocol.

### Per-Node TLS name

For one AesingFlow profile shared by multiple Nodes, set this exact value in
the inbound:

```json
"serverName": "{{NODE_ADDRESS}}"
```

Immediately before sending the Xray config, the panel replaces it with the
`Address` configured for that specific Node. Every Node address must therefore
be a public DNS name covered by that Node's certificate. Do not use this token
when the Node address is an IP address or a private management endpoint.

For an existing vanilla panel, follow [MIGRATION_FROM_VANILLA.md](MIGRATION_FROM_VANILLA.md)
instead of starting with an empty database.
