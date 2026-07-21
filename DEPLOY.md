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

This panel image validates AesingFlow configuration. A Node is separate, and the
standard Remnawave Node core does not provide that protocol. This repository
also builds an AesingFlow-ready Node image:

```text
ghcr.io/astracat2022/rennawave-node-aesingflow:latest
```

For a baked-in core, add a repository variable or secret named
`AESINGFLOW_CORE_URL`. It must be a direct URL to the Linux `xray`/`rw-core`
executable, not a `.zip`, `.tar.gz`, or release page. The workflow is
**Build AesingFlow node image**.

If `AESINGFLOW_CORE_URL` is not set, the image still builds successfully and
keeps the upstream Remnawave Node core. In that mode, set `CUSTOM_CORE_URL` in
the Node `.env`; the Node downloads the AesingFlow-compatible core on container
start.

Install or update a Node server with:

```bash
sudo mkdir -p /opt/remnanode
cd /opt/remnanode

sudo curl -fsSLO https://raw.githubusercontent.com/ASTRACAT2022/Rennawave-/main/docker-compose.node-aesingflow.yml
sudo nano .env
```

The Node `.env` needs the same secret that is shown in the panel when creating
or editing that Node:

```dotenv
SECRET_KEY=replace-with-node-secret-from-panel
NODE_PORT=2222
CUSTOM_CORE_URL=https://example.com/path/to/aesingflow-xray-linux-amd64
```

Then start it:

```bash
cd /opt/remnanode
docker compose --env-file .env -f docker-compose.node-aesingflow.yml pull
docker compose --env-file .env -f docker-compose.node-aesingflow.yml up -d
docker compose --env-file .env -f docker-compose.node-aesingflow.yml ps
docker logs --tail=100 remnanode
docker exec remnanode rw-core version
```

If you do not have a ready direct binary URL yet, you can temporarily use the
official image and pass `CUSTOM_CORE_URL` in the Node compose file. The final
image above is cleaner because the Node does not download its core on every
container recreation.

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

Do not set `streamSettings.network` to `aesingflow`. Current AesingFlow core
builds expose AesingFlow as an inbound protocol, while TLS still uses the
standard Xray `streamSettings.tlsSettings` object. Leave `network` omitted, or
use a standard Xray value such as `tcp` only if your core explicitly requires
it.

For an existing vanilla panel, follow [MIGRATION_FROM_VANILLA.md](MIGRATION_FROM_VANILLA.md)
instead of starting with an empty database.
