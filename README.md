# Rennawave-

Remnawave backend and frontend with AesingFlow configuration validation.

## Quick Docker test

Runs the panel, PostgreSQL, and Valkey on port 80. This is only for testing;
it does not deploy a Node or provide an AesingFlow-compatible Xray core.

```bash
curl -fsSL https://raw.githubusercontent.com/ASTRACAT2022/Rennawave-/main/install-test.sh | sudo bash
```

The first build downloads images and compiles the frontend, so it may take a
few minutes. The script refuses to overwrite `/opt/rennawave-test`.

## Production-style Docker deployment

GitHub Actions builds the unified frontend/backend image and publishes it to
GHCR. The deployment compose file and complete server instructions are in
[DEPLOY.md](DEPLOY.md).
