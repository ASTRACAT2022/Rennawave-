# Migration from a vanilla Remnawave panel

This procedure migrates an existing standard Remnawave installation to the
`ghcr.io/astracat2022/rennawave-aesingflow` image without intentionally
creating a new database, users, nodes, subscriptions, hosts, or API tokens.

It applies to the single-container compose layout with containers named
`remnawave`, `remnawave-db`, and `remnawave-redis`, and volumes named
`remnawave-db-data` and `valkey-socket`.

> Do not run `docker compose down -v`, `docker volume prune`, or recreate the
> PostgreSQL volume during this migration. Those commands remove data.

## What is retained

- PostgreSQL volume: users, subscriptions, nodes, hosts, configuration
  profiles, settings, API tokens, and audit data.
- Existing `.env`: application secrets, database password, domains, and node
  communication settings.
- `/var/lib/remnawave/configs`: Xray configuration assets and TLS certificate
  files. The fork compose persists this directory in `remnawave-configs`.

## 1. Pre-flight checks

Run this on the panel server in the directory containing the current vanilla
compose file and `.env`:

```bash
docker compose ps
docker volume inspect remnawave-db-data >/dev/null
docker volume inspect valkey-socket >/dev/null
```

Confirm that the database container is healthy and that there is enough free
space for a database dump and a copy of the configuration directory.

## 2. Create an offline backup

Create a backup directory with permissions limited to the current user:

```bash
BACKUP_DIR="$HOME/remnawave-backup-$(date +%F-%H%M%S)"
mkdir -p "$BACKUP_DIR"
chmod 700 "$BACKUP_DIR"

docker compose exec -T remnawave-db sh -lc \
  'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc' \
  > "$BACKUP_DIR/postgres.dump"

pg_restore -l "$BACKUP_DIR/postgres.dump" >/dev/null
cp .env "$BACKUP_DIR/panel.env"
cp docker-compose.yml "$BACKUP_DIR/vanilla-compose.yml"
docker cp remnawave:/var/lib/remnawave/configs "$BACKUP_DIR/configs"

# Optional but recommended: an exact archive of the PostgreSQL volume.
docker run --rm \
  -v remnawave-db-data:/from:ro \
  -v "$BACKUP_DIR":/backup \
  alpine:3.21 \
  tar -czf /backup/postgres-volume.tar.gz -C /from .

sha256sum "$BACKUP_DIR/postgres.dump" "$BACKUP_DIR/postgres-volume.tar.gz"
```

Copy this directory off the server before continuing. Keep it private: it
contains database data, session secrets, and possibly private TLS keys.

If your old container has another name, use `docker compose ps` and replace
`remnawave` in the `docker cp` command.

## 3. Stop the old panel without deleting volumes

```bash
docker compose down
```

This stops and removes containers only. Do **not** append `-v`.

## 4. Install the fork compose file

Use a separate deployment directory:

```bash
sudo mkdir -p /opt/rennawave
sudo curl -fsSL \
  https://raw.githubusercontent.com/ASTRACAT2022/Rennawave-/main/docker-compose.fork.yml \
  -o /opt/rennawave/docker-compose.fork.yml
sudo cp "$BACKUP_DIR/panel.env" /opt/rennawave/.env
sudo chown "$USER":"$USER" /opt/rennawave/.env /opt/rennawave/docker-compose.fork.yml
cd /opt/rennawave
```

Do not generate new JWT, API-token, or application secrets. Keeping the old
`.env` preserves active sessions and communication with existing Nodes.

Pin the first migration to the image built by GitHub Actions:

```bash
echo 'REMNAWAVE_IMAGE=ghcr.io/astracat2022/rennawave-aesingflow:sha-5d21332afd87009c9ebb7a941bcdf008af603b70' >> .env
```

## 5. Restore persistent configuration files

Create the new configuration volume and populate it **before** starting the
fork container:

```bash
docker volume create remnawave-configs
docker run --rm \
  -v remnawave-configs:/to \
  -v "$BACKUP_DIR/configs":/from:ro \
  alpine:3.21 \
  sh -c 'cp -a /from/. /to/'
```

This step preserves certificate files referenced by
`tlsSettings.certificates[].keyFile` and `certificateFile`. Do not create a
separate AesingFlow certificate store.

## 6. Start the fork and verify it

```bash
docker compose -f docker-compose.fork.yml pull
docker compose -f docker-compose.fork.yml up -d
docker compose -f docker-compose.fork.yml ps
docker compose -f docker-compose.fork.yml logs --tail=150 remnawave
curl -fsS http://127.0.0.1:3001/health
```

The fork compose intentionally pins PostgreSQL 17 with
`PGDATA=/var/lib/postgresql` for the existing vanilla volume layout. Do not
change it to PostgreSQL 18 as part of this panel migration; upgrading the
database major version is a separate `pg_upgrade` maintenance task.

Valkey keeps its Unix socket and also listens on its internal Docker TCP port
6379. This preserves compatibility with vanilla `.env` files that use either
`REDIS_SOCKET` or `REDIS_HOST`/`REDIS_PORT`; the port is not published on the
host.

The first start applies the normal Remnawave migrations. Then verify in the UI:

1. Existing users, hosts, Nodes, and subscriptions are present.
2. One existing standard Xray config still validates.
3. TLS certificate paths under `/var/lib/remnawave/configs/xray/ssl/` exist in
   the running container.
4. Existing Nodes reconnect; do not delete and recreate them.

### Common first-start errors

If Docker reports that `remnawave`, `remnawave-db`, or `remnawave-redis` is
already in use, the vanilla containers are still present. Return to the old
compose directory and run `docker compose down` **without** `-v`, then retry.
`docker stop` alone is not enough: it leaves the container names reserved. If
the old compose directory is unavailable, remove only the stopped containers:

```bash
docker rm remnawave remnawave-db remnawave-redis
```

This does not remove named volumes. The fork compose has `name: remnawave` so
it deliberately reuses volumes and the network created by the vanilla project.

If Compose reports an unset `POSTGRES_*` variable, verify that `/opt/rennawave/.env`
contains literal values, not shell-variable references:

```dotenv
POSTGRES_USER=postgres
POSTGRES_PASSWORD=the-existing-password
POSTGRES_DB=postgres
DATABASE_URL="postgresql://postgres:the-existing-password@remnawave-db:5432/postgres"
```

Do not use `$POSTGRES_PASSWORD` inside `DATABASE_URL`; retain the existing
password value from the vanilla `.env`.

## 7. Rollback plan

If the fork does not start, stop it without volumes:

```bash
cd /opt/rennawave
docker compose -f docker-compose.fork.yml down
```

If no database migration was applied, return to the saved vanilla compose:

```bash
BACKUP_DIR=/path/to/remnawave-backup-YYYY-MM-DD-HHMMSS
cd "$(dirname "$BACKUP_DIR")"
mkdir -p vanilla-rollback
cp "$BACKUP_DIR/vanilla-compose.yml" vanilla-rollback/docker-compose.yml
cp "$BACKUP_DIR/panel.env" vanilla-rollback/.env
cd vanilla-rollback
docker compose up -d
```

If a migration completed and a database rollback is required, restore from the
offline backup during a maintenance window. Do not delete or overwrite
`remnawave-db-data` until you have verified both `postgres.dump` and
`postgres-volume.tar.gz` and have a second copy off the server.

## Important AesingFlow limitation

The panel validates AesingFlow configurations. It does not turn a standard
Remnawave Node into an AesingFlow-capable core. Add an AesingFlow inbound only
after each Node runs `ghcr.io/astracat2022/rennawave-node-aesingflow`, which
bakes in the matching custom Xray core.
