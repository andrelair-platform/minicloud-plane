---
id: intro
title: Overview
sidebar_label: Overview
slug: /
---

# minicloud Plane

**Level 4 integration service** for Plane CE — Go microservice bridging the Plane project management API with the minicloud platform via a webhook-to-NATS event bus and a REST API consumed by Backstage.

## Responsibility

| In scope | Out of scope |
|---|---|
| Plane REST API client (issues, cycles, states, members) | Plane CE deployment (minicloud-gitops) |
| Webhook receiver — validates + publishes to NATS JetStream | NATS broker configuration |
| REST API for Backstage catalog integration | Backstage plugin wiring (minicloud-backstage) |
| GitOps overlay path management (minicloud-1) | |

## Stack

| Concern | Choice |
|---|---|
| Language | Go 1.25 |
| HTTP router | chi v5 |
| Messaging | NATS JetStream |
| Plane API | REST v1 (API key auth) |
| Container | `distroless/static-debian12:nonroot` |
| Registry | `harbor.10.0.0.200.nip.io/library/minicloud-plane` |

## Event flow

```
Plane webhook POST
        │
        ▼
 minicloud-plane
  ┌──────────────┐
  │  Validate    │
  │  signature   │
  └──────┬───────┘
         │ publish
         ▼
  NATS JetStream
  subject: plane.issues.*

         │ also serves
         ▼
  Backstage REST API
  GET /issues?project=...
```

## Links

- [GitHub repository](https://github.com/andrelair-platform/minicloud-plane)
- [Plane instance](https://plane.devandre.sbs)
- [Platform documentation](https://andrelair-platform.github.io/minicloud-platform-docs/)
