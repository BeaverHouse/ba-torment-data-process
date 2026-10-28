## About this CLI

A deployable batch CLI that builds BA Torment data from external sources and saves
it to Postgres and public object storage.
It is a single Cobra binary `batorment`; each task is a subcommand.

### Subcommands

1. `update-from-schaledb` — sync student/present data from SchaleDB into Postgres.
2. `process-raid` — turn raw party data (DuckDB) into refined party/filter/summary
   data with verified video references, and upload JSON files to object storage.
3. `total-analysis` — aggregate analysis across raids (by assault / by student).

Jobs receive settings through environment variables. `env.yaml` selects production;
`env.test.yaml` selects the separate test Supabase database and storage.

### Test backend

The analyzer, collector and web app use the same test database and public bucket.
The database schema remains `batorment_v3`. The collector's `FILE_MANAGER_URL`
points to the local file manager on port 8081, which must use its test backend.
`SUPABASE_PUBLIC_S3_CDN_URL` selects the bucket used to read cached party JSON.
The analyzer uses that same CDN; the web app uses `NEXT_PUBLIC_CDN_URL`.

SchaleDB and raw raid DuckDB files remain shared public inputs. Run
`update-from-schaledb`, `process-raid` and `total-analysis` against the test backend
to refresh their outputs. Wiki content keeps its existing shared source.

Video metadata and reviewed analysis cannot be rebuilt from those inputs. Copy
`contents`, `youtube_metadata` and `youtube_analysis` from production when seeding
an empty test database, preserving IDs and resetting the two video ID sequences.
Copy public objects under `batorment/` as the matching initial storage snapshot.
Leave the analysis queue, feedback and evaluation run history empty.

### Container pipeline

`docker-entrypoint.sh` runs the scheduled batch in order, failing fast on any step:

```
1) update-from-schaledb
2) process-raid
3) total-analysis
```

### Build & deploy

Single binary, multi-stage Dockerfile. Debian is used for the runtime, since DuckDB requires `glibc`.

The `Build and Push Docker Image` GitHub
workflow (`.github/workflows/docker-build.yml`, manual) builds multi-arch image and pushes to GHCR.

### Codebase

Being a CLI, it has no HTTP/MCP layer: `cmd/` (Cobra commands) → `internal/logic/` (domain logic) → `internal/db/postgres/` (sqlc-generated).  
Errors go through go-common `errorhandle` package; SQL lives in `internal/db/postgres/sql/` (`sqlc generate`).  
The shared standards are enforced by Austin's harness rules.
