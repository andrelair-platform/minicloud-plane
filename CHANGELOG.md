# Changelog

## [0.1.2](https://github.com/andrelair-platform/minicloud-plane/compare/minicloud-plane-v0.1.1...minicloud-plane-v0.1.2) (2026-08-14)


### Bug Fixes

* **website:** correct sidebars.ts brace syntax ([6477d41](https://github.com/andrelair-platform/minicloud-plane/commit/6477d414a46c1419f38a194b610f7ea890bd64e2))

## [0.1.1](https://github.com/andrelair-platform/minicloud-plane/compare/minicloud-plane-v0.1.0...minicloud-plane-v0.1.1) (2026-08-14)


### Features

* add Prometheus instrumentation — http_requests_total + http_request_duration_seconds ([ac73d5c](https://github.com/andrelair-platform/minicloud-plane/commit/ac73d5cd984c99f4f59bc43087d08975dff1331f))
* **catalog:** add Backstage catalog-info.yaml with plane.io/project-id ([263e186](https://github.com/andrelair-platform/minicloud-plane/commit/263e1862ffa3c98fa9c89f287e10493c0748d4ef))
* **catalog:** add Backstage catalog-info.yaml with plane.io/project-id annotation ([23a253b](https://github.com/andrelair-platform/minicloud-plane/commit/23a253b12cb6fea76287bea64f0430946d114393))
* **catalog:** add OpenAPI spec + providesApis relationship ([5444c21](https://github.com/andrelair-platform/minicloud-plane/commit/5444c216ef3323fb478d35188b305076f1cde9ba))
* initial minicloud-plane Go integration service ([f98feab](https://github.com/andrelair-platform/minicloud-plane/commit/f98feab9d352967128cf47c890e1a4c701744b62))
* OTel tracing + Prometheus exemplars (Gaps C+E) ([4134991](https://github.com/andrelair-platform/minicloud-plane/commit/41349913b36c6afd9b065f7fec5f0bc942a697fa))


### Bug Fixes

* API integration fixes + webhook pipeline operational ([9f03750](https://github.com/andrelair-platform/minicloud-plane/commit/9f03750bea896dd0393d563ed24be4a90b4e6554))
* **ci:** add OCI Accept header to Harbor pre-flight manifest check ([cdd6642](https://github.com/andrelair-platform/minicloud-plane/commit/cdd6642c4e5d51a90dc04760e8e736479e13d71d))
* **ci:** bump gitops via PR + admin auto-merge instead of direct push ([c8f9281](https://github.com/andrelair-platform/minicloud-plane/commit/c8f9281ae429240c94fdce2b5131904f609d604b))
* **ci:** bump Go 1.23→1.25 to fix CVE-2025-68121 (stdlib TLS) ([32bed8a](https://github.com/andrelair-platform/minicloud-plane/commit/32bed8a18d1ea1110c00292d32e15a31f4ddcc40))
* **ci:** bump overlays/dev on main push, not overlays/prod ([72a4aac](https://github.com/andrelair-platform/minicloud-plane/commit/72a4aac0ecb36f482b3b09fb6c51d36f4f5b8ca2))
* **ci:** use setup-kustomize action instead of flaky install script ([bf7c97f](https://github.com/andrelair-platform/minicloud-plane/commit/bf7c97f5562167847d6d0285b79e8722e83b9965))
* **metrics:** instrument /health to seed prometheus counters via probes ([83d5555](https://github.com/andrelair-platform/minicloud-plane/commit/83d5555166ffdd6ee46e54da1d4240f2282c18bf))

## Changelog

All notable changes to minicloud-plane are documented here.

This file is maintained by [release-please](https://github.com/googleapis/release-please).
