# Documentation

## What is this repo?

This repo is for processing BA Torment data from various external sources.

1. Download data from external sources.
   - Student, Item, and specific static data from [Schale DB](https://schaledb.com/).
   - Party data from [Plana Stats](https://www.plana-stats.com/).
     - The raw form of party data is DuckDB.
   - Self-implemented Youtube video analysis pipeline by Gemini.
2. Process the data and store it in database.
   - Party data with compressed, refined form.
   - Filter data from party data, with various conditions.
   - Summary data of individual assaults.
   - Student mapping JSON (with search keywords too)
   - Total analysis, by assaults or by students.

Most of the processed data is uploaded to Supabase Storage, but some data is stored in Supabase Database.

## Data structure

1. [DuckDB data](./data-duckdb.md)
2. BA Torment data
   - [Party data](./data-batorment-party.md)
   - [Filter data](./data-batorment-filter.md)
   - [Summary data](./data-batorment-summary.md)
   - Total analysis: WIP
   - Student mapping JSON: WIP

## Concept & Deployment

See [Concept](./concept.md) for the CLI's subcommands, container pipeline, and build/deploy.

## Environment variables

Environment variables are configured in [deploy.yaml](../deploy.yaml).  
The remote environment variables are stored in external secret manager, and are loaded via personal CLI.  
When you want to run this code, set up the environment variables manually with following description.
