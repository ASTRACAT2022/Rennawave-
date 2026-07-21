# AesingFlow

This Node image includes a native Xray outbound named `aesingflow`. It creates
one multiplexed AesingFlow QUIC connection per outbound profile and carries
TCP CONNECT streams through it.

```json
{
  "protocol": "aesingflow",
  "tag": "aesingflow-de1",
  "settings": {
    "server": "de1.node.example",
    "serverPort": 4433,
    "token": "replace-with-secret-token",
    "tls": {
      "serverName": "de1.node.example",
      "caFile": "/etc/ssl/certs/aesingflow-private-ca.pem"
    },
    "maxStreams": 256,
    "brutalBps": 250000000
  }
}
```

`tls.serverName` defaults to `server` for a hostname. It is required when
`server` is an IP address. TLS 1.3 certificate verification is always enabled;
there is intentionally no insecure mode. Omit `caFile` when the certificate is
issued by a public CA; otherwise provide a file containing only the public PEM
certificate for the private CA.

The outbound accepts TCP only. Xray requests for UDP return an explicit
unsupported-protocol error and are never silently routed outside AesingFlow.
`brutalBps` is the client-side send ceiling in bit/s; zero uses AesingFlow's
default. Set `disableBrutal` to `true` to use CUBIC instead.

See the [AesingFlow core integration manual](https://github.com/ASTRACAT2022/AesingFlow/blob/main/docs/core-integration-manual.md)
for operational and verification guidance.

## Remnawave inbound with standard Xray TLS

The inbound uses the registered `aesingflow` transport because Xray's stream
model does not accept `network: "udp"`. TLS is still the single standard Xray
TLS layer: do not add certificate or key fields to AesingFlow settings.

```json
{
  "tag": "AESINGFLOW-4433",
  "listen": "0.0.0.0",
  "port": 4433,
  "protocol": "aesingflow",
  "settings": {
    "clients": [{
      "id": "REMNAWAVE_USER_UUID",
      "token": "INDIVIDUAL_SECRET_TOKEN",
      "email": "REMNAWAVE_USER_UUID",
      "level": 0
    }],
    "maxStreams": 256,
    "brutalBps": 250000000
  },
  "streamSettings": {
    "network": "aesingflow",
    "security": "tls",
    "tlsSettings": {
      "serverName": "de1.node.astracat.ru",
      "minVersion": "1.3",
      "alpn": ["aesingflow"],
      "certificates": [{
        "keyFile": "/var/lib/remnawave/configs/xray/ssl/privkey.key",
        "certificateFile": "/var/lib/remnawave/configs/xray/ssl/fullchain.pem"
      }]
    }
  }
}
```

At startup, the Node accepts only `.key` private keys and `.pem` certificate
chains below `/var/lib/remnawave/configs/xray/ssl/`. It rejects traversal and
escaping symlinks, broad key permissions, malformed PEM, mismatched keys,
expired or non-server certificates, a missing SAN for `serverName`, and any
TLS version other than 1.3. Validation failures stop the inbound; no fallback
certificate or insecure mode is created.

Mount the same Remnawave certificate directory read-only on the Node. The
official Remnawave Node delivery flow retains these internal paths; never put
the key or certificate contents into an API response, subscription, database,
or `aesingflow://` link. After replacing the mounted files, perform the normal
controlled Node reload so Xray's standard certificate loader validates and
uses the new pair.
