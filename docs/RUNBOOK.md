# Infra & Deployment Runbook

## Topology

- **App**: Fly.io app `user-service-weathered-pine-7843`, org `ruben-tanahara-902`, region `dfw` (Dallas — closest available to Mexico, Fly has no MX region).
- **Config**: `fly.toml` at repo root. `internal_port = 5001`, `auto_stop_machines = true` (scales to zero when idle).
- **Database**: MongoDB Atlas, project `user_service-prod`, cluster `Cluster0`, database `user_service`. DB user `rubenmtztanahara_db_user`, scoped `readWrite` on that DB only.
- **Secrets**: `JWT_SECRET`, `MONGO_URI`, `MONGO_DB` — set via `fly secrets`, never committed. Prod `JWT_SECRET` is distinct from the local `.env` dev value.

## Deploy

```bash
fly deploy --app user-service-weathered-pine-7843
```

Builds from `Dockerfile`, pushes to Fly's registry, rolls out with zero-downtime (2 machines).

## Secrets

```bash
fly secrets list --app user-service-weathered-pine-7843   # names + digests only
fly secrets set KEY="value" --app user-service-weathered-pine-7843
```

Setting a secret stages it; it takes effect on the next `fly deploy` or `fly apps restart`.

## Health checks

```bash
curl https://user-service-weathered-pine-7843.fly.dev/health/live    # process up
curl https://user-service-weathered-pine-7843.fly.dev/health/ready   # process up + Mongo reachable
```

## Status & logs

```bash
fly status --app user-service-weathered-pine-7843          # machine states
fly logs --app user-service-weathered-pine-7843 --no-tail  # snapshot, not a stream
fly logs --app user-service-weathered-pine-7843            # streams, doesn't exit — background it
```

## Restart (without a new deploy)

```bash
fly apps restart user-service-weathered-pine-7843
```

Use after fixing something external to the app image — e.g. an Atlas network access change.

## Known failure mode: TLS handshake errors from Atlas

**Symptom**: `fly logs` shows `remote error: tls: internal error` from all three Atlas shard
hosts, `/health/ready` never comes up, machines crash-loop.

**Cause**: Fly has no static egress IP. If the connecting IP isn't in Atlas's Network Access
list, Atlas rejects at the TLS layer rather than with a normal auth error.

**Fix**: Atlas UI → Network Access → Add IP Address → `0.0.0.0/0` ("Allow Access from
Anywhere"). Safe here — Atlas still enforces TLS plus the scoped DB user/password. Then
`fly apps restart <app>`.

## Scaling

Fly defaults new apps to 2 machines per process group (HA — zero-downtime rolling deploys,
failover if one crashes). `min_machines_running = 0` in `fly.toml` only allows scale-to-zero
on idle, it does not reduce the machine count.

```bash
fly scale count 1 --app user-service-weathered-pine-7843 --yes   # single machine
fly scale count 2 --app user-service-weathered-pine-7843 --yes   # back to HA
```

Running on 1 machine loses zero-downtime deploys and instant failover — acceptable for a
low-traffic single-service API optimizing for cost.

## Rollback

```bash
fly releases --app user-service-weathered-pine-7843        # list past releases
fly deploy --image <previous-image-ref> --app user-service-weathered-pine-7843
```

## Change Log

- 2026-08-26: Initial deploy — Fly.io (`dfw`) + MongoDB Atlas (`Cluster0`). Hit and resolved
  the Atlas TLS/network-access failure mode documented above.
