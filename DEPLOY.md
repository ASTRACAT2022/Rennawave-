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

The workflow is **Build AesingFlow node image**. It builds the bundled
`node-aesingflow/Xray-core` source and bakes that custom Xray binary into the
image. No `AESINGFLOW_CORE_URL` or `CUSTOM_CORE_URL` is required.

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
```

If the AesingFlow inbound uses certificate paths such as
`/var/lib/remnawave/configs/xray/ssl/fullchain.pem`, mount those files or their
directory into the Node container. For example, if the Node host has
`/opt/remnawave/nginx/fullchain.pem` and `/opt/remnawave/nginx/privkey.key`,
uncomment the `volumes` block in `docker-compose.node-aesingflow.yml` or add:

```yaml
    volumes:
      - /opt/remnawave/nginx:/var/lib/remnawave/configs/xray/ssl:ro
```

The private key must not be group/world-readable:

```bash
sudo chmod 600 /opt/remnawave/nginx/privkey.key
sudo chmod 644 /opt/remnawave/nginx/fullchain.pem
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

The bundled custom core registers AesingFlow both as an inbound protocol and
as an Xray stream transport. Therefore set `streamSettings.network` to
`aesingflow`, keep `streamSettings.security` as `tls`, and set
`tlsSettings.alpn` so it includes `aesingflow`. TLS still uses the standard
Xray `streamSettings.tlsSettings` object; do not create a second certificate
store in AesingFlow settings.

Minimal inbound template:

```json
{
  "tag": "AESINGFLOW-4433",
  "listen": "0.0.0.0",
  "port": 4433,
  "protocol": "aesingflow",
  "settings": {
    "clients": [],
    "maxStreams": 256,
    "brutalBps": 250000000
  },
  "streamSettings": {
    "network": "aesingflow",
    "security": "tls",
    "tlsSettings": {
      "serverName": "{{NODE_ADDRESS}}",
      "minVersion": "1.3",
      "alpn": ["aesingflow"],
      "certificates": [
        {
          "keyFile": "/var/lib/remnawave/configs/xray/ssl/privkey.key",
          "certificateFile": "/var/lib/remnawave/configs/xray/ssl/fullchain.pem"
        }
      ]
    }
  }
}
```

For an existing vanilla panel, follow [MIGRATION_FROM_VANILLA.md](MIGRATION_FROM_VANILLA.md)
instead of starting with an empty database.
