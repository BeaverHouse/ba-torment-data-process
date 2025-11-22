# Commands

This document describes the available commands in the `cmd/` directory and their purposes.

## Available Commands

### 1. `process_raid` (Main Integrated Pipeline)

**Path:** `cmd/process_raid/`

**Purpose:** Complete integrated pipeline for processing raid data from DuckDB to final output with video references and filters.

**Process Flow:**
```
1. Parse DuckDB → Generate party data
2. Update video references → Match with verified YouTube analysis
3. Upload party data → Save to S3 with video refs
4. Create video filter → Generate filter from video analysis
5. Upload filters → lunatic, non-lunatic, basic filters
6. Generate summary → Process and upload summary data
```

**Usage:**
```bash
go run cmd/process_raid/main.go [--dry-run]
```

**Features:**
- Single command execution for complete data processing
- Memory-based data sharing between steps (no redundant S3 downloads)
- Graceful error handling (continues on non-critical failures)
- Detailed step-by-step logging

**Configuration:**
- Edit the `raids` array in `main.go` to specify which raids to process

---

### 2. `update_video_ref` (Legacy - Standalone)

**Path:** `cmd/update_video_ref/`

**Purpose:** Updates video references for existing party data by downloading from S3 and matching with verified YouTube analysis.

**Usage:**
```bash
go run cmd/update_video_ref/main.go [--dry-run]
```

**Note:** This is a legacy command. For new raid processing, use `process_raid` instead which integrates this functionality without S3 downloads.

**When to use:**
- Updating video refs for raids that were processed before the integrated pipeline
- Reprocessing video refs when new YouTube analysis data is available

---

### 3. `update_video_filter` (Legacy - Standalone)

**Path:** `cmd/update_video_filter/`

**Purpose:** Creates and uploads video filters based on verified YouTube analysis data from the API.

**Usage:**
```bash
go run cmd/update_video_filter/main.go [--dry-run]
```

**Note:** This is a legacy command. For new raid processing, use `process_raid` instead which integrates this functionality.

**When to use:**
- Updating video filters independently for existing raids
- Regenerating filters when video analysis data changes

---

### 4. `update_from_schaledb`

**Path:** `cmd/update_from_schaledb/`

**Purpose:** Updates student and present data from SchaleDB API to the database.

**Usage:**
```bash
go run cmd/update_from_schaledb/main.go
```

**Process:**
- Fetches latest student data from SchaleDB
- Fetches latest present/gift data from SchaleDB
- Updates PostgreSQL database with the fetched data

**When to use:**
- After new students are added to the game
- When SchaleDB data needs to be synced

---

## Command Comparison

### For New Raid Processing

**✅ Recommended: `process_raid`**
- Complete end-to-end processing
- Efficient memory usage
- Single command execution
- Automatic video ref integration

### For Updating Existing Data

**Video References Only:**
- Use `update_video_ref` if you need to update video refs for existing party data

**Video Filters Only:**
- Use `update_video_filter` if you need to regenerate video filters independently

**Student Database:**
- Use `update_from_schaledb` for SchaleDB data synchronization

---

## Flags

All raid processing commands support the following flag:

- `--dry-run`: Run the command without actually uploading data to S3
  - Useful for testing and validation
  - All processing steps are executed, but upload operations are skipped

**Example:**
```bash
go run cmd/process_raid/main.go --dry-run
```

---

## Migration Guide

### From `parse_duckdb` (deprecated)

The `parse_duckdb` command has been replaced by `process_raid`. The new command includes:
- All original DuckDB parsing functionality
- Integrated video reference updates
- Integrated video filter generation
- Better logging and error handling

**No code changes required** - just use the new command name:
```bash
# Old
go run cmd/parse_duckdb/main.go

# New
go run cmd/process_raid/main.go
```

### From separate commands

If you were running multiple commands separately:
```bash
# Old approach
go run cmd/parse_duckdb/main.go
go run cmd/update_video_ref/main.go
go run cmd/update_video_filter/main.go
```

You can now use a single command:
```bash
# New approach
go run cmd/process_raid/main.go
```

This is more efficient because:
- No S3 downloads needed between steps
- Data stays in memory throughout the pipeline
- Fewer API calls and network operations
- Automatic coordination between steps
