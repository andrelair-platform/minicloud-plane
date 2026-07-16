# minicloud-plane

[![CI](https://github.com/andrelair-platform/minicloud-plane/actions/workflows/ci.yml/badge.svg)](https://github.com/andrelair-platform/minicloud-plane/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.25-blue)](https://go.dev)
[![Supply chain: cosign](https://img.shields.io/badge/supply%20chain-cosign%20signed-green)](https://github.com/sigstore/cosign)

> Level 4 integration service for [Plane CE](https://plane.so) on the minicloud enterprise Kubernetes platform. Exposes a REST proxy API that Backstage consumes via its proxy plugin, and a webhook→NATS bridge that publishes Plane events to the cluster messaging layer for downstream consumers.

**Live:** [https://plane.devandre.sbs](https://plane.devandre.sbs)

---

## Table of Contents

- [What it does](#what-it-does)
- [Architecture](#architecture)
- [API Reference](#api-reference)
- [NATS events](#nats-events)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [CI/CD Pipeline](#cicd-pipeline)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

---

## What it does

The service has two independent responsibilities on a single HTTP server (`0.0.0.0:8080`):

| Responsibility | Path | Consumer |
|---|---|---|
| **REST proxy** | `GET /api/projects`, `GET /api/projects/{id}/issues` | Backstage `@internal/plugin-minicloud-plane` via `/api/proxy/minicloud-plane/` |
| **Webhook bridge** | `POST /webhook` | Plane CE webhook → NATS subject `plane.<workspace>.<event>.<action>` |
| **Health check** | `GET /health` | Kubernetes liveness probe, ArgoCD sync check |

The REST proxy keeps Backstage decoupled from Plane's raw API — the token lives in the service, not in the frontend. The webhook bridge makes Plane events available to any NATS subscriber in the cluster without polling.

---

## Architecture

```
Plane CE (plane.devandre.sbs)
    │
    ├─ [REST]     ←── GET /api/v1/workspaces/{ws}/projects/
    │                  GET /api/v1/workspaces/{ws}/projects/{id}/issues/
    │                  ↑ internal/plane — Client with X-Api-Key auth, 15s timeout
    │
    └─ [Webhook]  ──► POST /webhook
                       ↑ internal/webhook — HMAC-SHA256 signature verification
                              │
                              ▼
                       internal/nats — Publisher
                              │
                              ▼
                       NATS (nats.messaging.svc.cluster.local:4222)
                       Subject: plane.<workspace>.<event>.<action>
                       e.g.    plane.andrelair.issue.created


Backstage (backstage namespace)
    │
    │  GET /api/proxy/minicloud-plane/api/projects/PT/issues
    ▼
minicloud-plane:8080  →  /api/projects/PT/issues
    │
    └─ internal/api — Handler proxies to Plane, returns JSON array
```

**Packages:**

| Package | Role |
|---|---|
| `cmd/server` | Entry point — wires env vars, constructs dependencies, starts HTTP server |
| `internal/plane` | Plane API client — `GET projects`, `GET issues`, `GET issue` with typed response structs |
| `internal/api` | HTTP handler for `/api/` — proxies to Plane client, resolves project identifier → UUID |
| `internal/webhook` | HTTP handler for `/webhook` — verifies HMAC-SHA256 signature, unmarshals event, calls publisher |
| `internal/nats` | NATS publisher — connects with auto-reconnect, publishes JSON payloads on `plane.<ws>.<event>.<action>` |

**Runtime:**

| Component | Detail |
|---|---|
| Language | Go 1.25 |
| Container | `gcr.io/distroless/static:nonroot` — no shell, no package manager, non-root UID |
| Registry | `harbor.10.0.0.200.nip.io/library/minicloud-plane` |
| Namespace | `minicloud-plane-dev` (dev) — staging/prod overlays in minicloud-gitops |
| ArgoCD app | `minicloud-plane-dev` |
| GitOps | Kustomize overlays in `minicloud-gitops/services/minicloud-plane/overlays/` |

---

## API Reference

Full OpenAPI 3.0 spec is registered in the Backstage catalog (`catalog-info.yaml` — `kind: API`). Quick reference:

### `GET /health`
```json
{"status": "ok", "service": "minicloud-plane"}
```

### `GET /api/projects`
Returns all projects in the configured Plane workspace.

```json
[
  {
    "id": "uuid",
    "name": "Platform",
    "identifier": "PT",
    "description": "...",
    "created_at": "2026-01-01T00:00:00Z"
  }
]
```

### `GET /api/projects/{projectId}/issues`
Returns up to 100 issues for the given project ID or identifier (e.g. `PT`).

```json
[
  {
    "id": "uuid",
    "sequence_id": 42,
    "name": "Add OIDC to Backstage",
    "priority": "high",
    "state": "uuid",
    "assignees": ["uuid"],
    "created_at": "2026-01-01T00:00:00Z"
  }
]
```

Both endpoints return `502 Bad Gateway` if the upstream Plane API is unreachable.

---

## NATS events

Every POST to `/webhook` that passes signature verification is published to NATS:

```
Subject:  plane.<workspace>.<event>.<action>
Examples: plane.andrelair.issue.created
          plane.andrelair.issue.updated
          plane.andrelair.cycle.created
```

Payload: the full Plane webhook JSON body, re-serialised.

**Subject pattern** lets consumers subscribe selectively:
- All issue events: `plane.andrelair.issue.*`
- All events: `plane.andrelair.>`
- Specific action: `plane.andrelair.issue.created`

If the NATS publish fails, the webhook handler still returns `200 OK` to Plane. Plane retries on non-2xx responses — returning 500 would create a retry loop for a transient NATS issue.

---

## Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | ≥ 1.25 | Matches `go.mod` |
| Docker / Podman | any | Only needed for local image builds |
| NATS server | ≥ 2.x | Only needed for webhook bridge testing |

---

## Getting Started

### Run locally

```bash
git clone https://github.com/andrelair-platform/minicloud-plane.git
cd minicloud-plane

export PLANE_URL=https://plane.devandre.sbs
export PLANE_TOKEN=<your-api-token>        # Plane Settings → API tokens
export PLANE_WORKSPACE=andrelair           # Plane workspace slug
export NATS_URL=nats://localhost:4222      # optional — defaults to in-cluster URL
export PLANE_WEBHOOK_SECRET=              # optional — leave empty to skip HMAC check

go run ./cmd/server
```

### Run tests

```bash
go test -v ./...
```

### Build the container image locally

```bash
podman build -f Containerfile -t minicloud-plane:local .
# or
docker build -f Containerfile -t minicloud-plane:local .
```

### Smoke test the REST proxy

```bash
# Health
curl http://localhost:8080/health

# List projects
curl http://localhost:8080/api/projects | jq .

# List issues for project PT
curl http://localhost:8080/api/projects/PT/issues | jq '.[].name'
```

---

## Configuration

All configuration is via environment variables. No config files.

| Variable | Required | Default | Description |
|---|---|---|---|
| `PLANE_URL` | yes | — | Base URL of the Plane CE instance (e.g. `https://plane.devandre.sbs`) |
| `PLANE_TOKEN` | yes | — | Plane API token — Settings → API tokens in Plane |
| `PLANE_WORKSPACE` | yes | — | Plane workspace slug (visible in all Plane URLs) |
| `NATS_URL` | no | `nats://nats.messaging.svc.cluster.local:4222` | NATS server URL |
| `PLANE_WEBHOOK_SECRET` | no | `""` | HMAC-SHA256 secret for Plane webhook signature verification; leave empty to skip |
| `PORT` | no | `8080` | HTTP listen port |

**Secrets in production** are managed by External Secrets Operator pulling from Vault at `secret/platform/minicloud-plane`. The Kubernetes Secret is named `minicloud-plane-secret` in the `minicloud-plane-dev` namespace.

### Backstage proxy config

The Backstage backend proxy is configured in `minicloud-gitops/helm-values/backstage-values.yaml`:

```yaml
proxy:
  endpoints:
    '/minicloud-plane':
      target: 'http://minicloud-plane.minicloud-plane-dev.svc.cluster.local:8080'
      changeOrigin: true
```

The `@internal/plugin-minicloud-plane` plugin in Backstage calls `/api/proxy/minicloud-plane/api/projects/{id}/issues` — all routing is done by the Backstage backend.

### Plane webhook setup

In Plane: **Settings → Webhooks → Add webhook**

| Field | Value |
|---|---|
| URL | `http://minicloud-plane.minicloud-plane-dev.svc.cluster.local:8080/webhook` |
| Secret | value of `PLANE_WEBHOOK_SECRET` |
| Events | issue (created, updated, deleted) — add more as needed |

---

## CI/CD Pipeline

Every push triggers `.github/workflows/ci.yml` — tests run first, build only proceeds on success:

```
push (any branch)
    │
    └─ test job: go test -v ./...
              │
              └─ build-and-push job (push events only):
                    │
                    ├─ 1. Connect to Tailscale (OAuth)
                    ├─ 2. Trust minicloud CA on runner
                    ├─ 3. docker build (Containerfile) → push to Harbor
                    ├─ 4. Trivy scan — fails on unfixed CRITICAL CVEs
                    ├─ 5. cosign sign (keyless — GitHub OIDC → Sigstore Fulcio)
                    ├─ 6. syft SBOM (CycloneDX JSON) — attached as OCI referrer
                    └─ bump-gitops job:
                          kustomize edit set image → GPG-signed commit to minicloud-gitops
                          overlay: dev | staging | prod (matches pushed branch)
```

**Branch behaviour:**

| Branch | Image tag | Trivy | Cosign | SBOM | GitOps overlay |
|---|---|---|---|---|---|
| `main` | `<sha>` | yes | yes | yes | `overlays/prod` |
| `staging` | `staging-<sha>` | yes | yes | no | `overlays/staging` |
| `dev` | `dev-<sha>` | yes | no | no | `overlays/dev` |

### Required secrets

All 7 secrets are **org-level on `andrelair-platform`** (visibility: all). New repos inherit them automatically.

| Secret | Purpose |
|---|---|
| `TS_OAUTH_CLIENT_ID` | Tailscale OAuth client ID — joins tailnet as `tag:ci` |
| `TS_OAUTH_SECRET` | Tailscale OAuth secret |
| `MINICLOUD_CA_CERT` | Self-signed CA PEM — trusts Harbor TLS |
| `HARBOR_USER` | Harbor registry username |
| `HARBOR_PASSWORD` | Harbor registry password |
| `GITOPS_TOKEN` | GitHub PAT (`repo` scope) for committing to `minicloud-gitops` |
| `GPG_PRIVATE_KEY` | Armored GPG private key for signing gitops commits (key ID `FD6D39D681DEFA34`) |

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Backstage Plane Issues tab shows no issues | Proxy not reaching service, or wrong project ID | Check pod: `kubectl logs -n minicloud-plane-dev -l app=minicloud-plane`; verify `plane.io/project-id` annotation on the catalog entity matches the Plane project identifier |
| `502 Bad Gateway` from REST proxy | Plane CE unreachable or token expired | Confirm `https://plane.devandre.sbs` is reachable from the pod; regenerate token in Plane Settings |
| Webhook returns `401 Unauthorized` | HMAC-SHA256 signature mismatch | Ensure `PLANE_WEBHOOK_SECRET` in the k8s secret matches the secret configured in Plane Settings → Webhooks |
| NATS publish warnings in logs | NATS server unreachable | The webhook still returns 200 to Plane; check NATS pod: `kubectl get pods -n messaging` |
| `NATS connect failed` on startup | NATS URL wrong or NATS not ready | Pod will crash-loop; ArgoCD will retry. Check `NATS_URL` env var and NATS pod status |
| Harbor pre-flight returns 404 | `build-push-action@v7` pushes an OCI image index; Harbor needs explicit Accept header | Accept header already set in `ci.yml` — verify it was not removed |
| Test job passes but image is wrong | `go test` runs without Tailscale; build-and-push runs after | Check `needs: test` is present on the `build-and-push` job in `ci.yml` |

---

## Contributing

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add issue state resolution (uuid → name)
fix: handle Plane API pagination beyond 100 issues
chore: bump go 1.25 → 1.26
ci: pin trivy action to v0.37.0
```

Branch strategy mirrors the rest of the platform:
- `dev` — direct push, CI builds `dev-<sha>` image, deploys to `overlays/dev`
- `staging` — PR required, cosign-signed, deploys to `overlays/staging`
- `main` — PR + GPG required, cosign + SBOM, deploys to `overlays/prod`

---

## License

[MIT](LICENSE) © andrelair-platform
