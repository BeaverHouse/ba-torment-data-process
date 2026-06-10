## About this CLI

A deployable batch CLI that builds BA Torment data from external sources and uploads it to Postgres DB.  
It is a single Cobra binary `batorment`; each task is a subcommand.

### Subcommands

1. `update-from-schaledb` — sync student/present data from SchaleDB into Postgres.
2. `process-raid` — turn raw party data (DuckDB) into refined party/filter/summary
   data with verified video references, and upload it to Postgres.
3. `total-analysis` — aggregate analysis across raids (by assault / by student).
4. `generate-student-grid-image` — render student grid images (run manually as needed).

Run locally with `go run . <subcommand>` (loads `.env` in the local env).

### Container pipeline

`docker-entrypoint.sh` runs the scheduled batch in order, failing fast on any step:

```
1) update-from-schaledb
2) process-raid
3) total-analysis
```

`generate-student-grid-image` is not part of the batch.

### Build & deploy

Single binary, multi-stage Dockerfile. Debian is used for the runtime, since DuckDB requires `glibc`.

The `Build and Push Docker Image` GitHub
workflow (`.github/workflows/docker-build.yml`, manual) builds multi-arch image and pushes to GHCR.

### Codebase

Being a CLI, it has no HTTP/MCP layer: `cmd/` (Cobra commands) → `internal/logic/` (domain logic) → `internal/db/postgres/` (sqlc-generated).  
Errors go through go-common `errorhandle` package; SQL lives in `internal/db/postgres/sql/` (`sqlc generate`).  
The shared standards are enforced by Austin's harness rules.
