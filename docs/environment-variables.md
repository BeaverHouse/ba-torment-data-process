# Environment variables

Use `.env` file for local development, and store it in secure location for production.

```
# .env

POSTGRES_HOST=<your-postgres-host>
POSTGRES_PORT=<your-postgres-port>
POSTGRES_USER=<your-postgres-user>
POSTGRES_PASSWORD=<your-postgres-password>
POSTGRES_DBNAME=<your-postgres-dbname>

YOUTUBE_API_KEY=<your-youtube-api-key>
BATORMENT_UPLOAD_URL=<your-batorment-upload-url>
BATORMENT_DOWNLOAD_URL=<your-batorment-download-url>
```

- `POSTGRES_*` are required for PostgreSQL connection.
- `YOUTUBE_API_KEY` is the API key for [YouTube Data API v3](https://developers.google.com/youtube/v3/docs).
- `BATORMENT_UPLOAD_URL` and `BATORMENT_DOWNLOAD_URL` are presigned URLs for [Oracle Object Storage](https://docs.oracle.com/en-us/iaas/Content/Object/Concepts/objectstorageoverview.htm).
