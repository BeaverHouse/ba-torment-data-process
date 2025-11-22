# Docker Deployment Guide

This guide explains how to build and run the BA Torment data processing pipeline using Docker.

## Overview

The Docker image includes both data processing commands:
1. **update_from_schaledb**: Updates student data from SchaleDB
2. **process_raid**: Processes raid data from DuckDB with integrated video references and filters

Both commands are executed sequentially in a single container run.

## Prerequisites

- Docker installed on your system
- Environment variables configured (see below)

## Quick Start

### Build the image:

```bash
docker build -t ba-torment-data-process:latest .
```

### Run the container:

```bash
docker run --rm \
  -e POSTGRES_HOST=your-db-host \
  -e POSTGRES_PORT=5432 \
  -e POSTGRES_USER=your-db-user \
  -e POSTGRES_PASSWORD=your-db-password \
  -e POSTGRES_DB=your-db-name \
  -e AWS_REGION=us-east-1 \
  -e AWS_ACCESS_KEY_ID=your-access-key \
  -e AWS_SECRET_ACCESS_KEY=your-secret-key \
  -e AWS_S3_BUCKET=your-bucket-name \
  -e BATORMENT_DUCKDB_REMOTE_URL=https://your-cloudfront-url \
  -e BA_ANALYZER_SERVICE_API_KEY=your-api-key \
  -e GO_ENV=production \
  ba-torment-data-process:latest
```

Or load from `.env` file:

```bash
docker run --rm --env-file .env ba-torment-data-process:latest
```

## Image Details

### Build Stages

The Dockerfile uses multi-stage builds for optimization:

1. **Builder stage** (golang:1.25.2-bookworm):
   - Downloads Go modules
   - Builds both binaries
   - Uses Debian Bookworm for glibc compatibility with DuckDB

2. **Runtime stage** (debian:bookworm-slim):
   - Minimal Debian base with glibc
   - Only includes runtime dependencies and compiled binaries
   - Optimized for size while maintaining DuckDB compatibility

### Entrypoint Script

The `docker-entrypoint.sh` script:
- Runs `update_from_schaledb` first
- Then runs `process_raid`
- Provides clear logging and error handling
- Exits with appropriate error codes on failure

## Configuration

### Environment Variables

All environment variables are configured via:
- `.env` file (for Docker Compose)
- `-e` flags (for Docker CLI)
- Cloud provider secrets manager (for production)

See [Environment Variables](./README.md#environment-variables) for detailed descriptions.

### Raid Configuration

To modify which raids to process, edit the `raids` array in `cmd/process_raid/main.go` before building the image:

```go
var (
	raids = []string{
		"3S27-1",
		"3S27-3",
		"3S27-4",
	}
)
```

Then rebuild the image.

## Production Deployment

### CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: Deploy BA Data Process

on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC
  workflow_dispatch:

jobs:
  process:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: docker build -t ba-data-process .

      - name: Run processing
        env:
          POSTGRES_HOST: ${{ secrets.POSTGRES_HOST }}
          POSTGRES_USER: ${{ secrets.POSTGRES_USER }}
          POSTGRES_PASSWORD: ${{ secrets.POSTGRES_PASSWORD }}
          # ... other secrets
        run: |
          docker run --rm --env-file <(env) ba-data-process
```

### Cloud Run (GCP)

```bash
# Build and push to Google Container Registry
gcloud builds submit --tag gcr.io/PROJECT_ID/ba-data-process

# Deploy to Cloud Run
gcloud run jobs create ba-data-process \
  --image gcr.io/PROJECT_ID/ba-data-process \
  --set-env-vars POSTGRES_HOST=... \
  --region us-central1

# Execute the job
gcloud run jobs execute ba-data-process
```

### AWS ECS

```bash
# Build and push to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com
docker tag ba-data-process:latest ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/ba-data-process:latest
docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/ba-data-process:latest

# Create and run ECS task with environment variables from Systems Manager Parameter Store
```

## Troubleshooting

### Common Issues

**Build fails with "go.mod not found"**
- Ensure you're running the build command from the project root

**Container exits immediately**
- Check environment variables are set correctly
- View logs: `docker logs CONTAINER_ID`

**Database connection fails**
- Verify database host is accessible from Docker network
- For localhost databases, use `host.docker.internal` instead of `localhost`

**Out of memory errors**
- Increase Docker memory limit in Docker Desktop settings
- Or use `docker run --memory=4g ...` to allocate more memory

### Debug Mode

Run container with shell to debug:

```bash
docker run --rm -it --entrypoint /bin/sh ba-torment-data-process:latest
```

Then manually run commands:
```sh
/app/bin/update_from_schaledb
/app/bin/process_raid
```

## Image Maintenance

### Rebuilding

After code changes:

```bash
docker build --no-cache -t ba-torment-data-process:latest .
```

### Cleanup

Remove old images and containers:

```bash
# Remove all stopped containers
docker container prune

# Remove unused images
docker image prune -a

# Remove everything (containers, images, volumes, networks)
docker system prune -a
```

## Performance Considerations

- **Build time**: Multi-stage build takes 2-5 minutes depending on network speed
- **Runtime**: Depends on number of raids to process (typically 5-15 minutes)
- **Image size**: ~80-120MB (compressed), ~200-300MB (uncompressed)
- **Memory usage**: Peak ~500MB-1GB during DuckDB processing
- **Note**: Uses Debian base instead of Alpine for DuckDB glibc compatibility

## Security

- Never commit `.env` files with sensitive data
- Use secrets management in production (AWS Secrets Manager, GCP Secret Manager, etc.)
- Regularly update base images for security patches
- Consider using distroless images for production
