# Refactoring backlog

Patterns found in the code that are worth fixing. Each entry names where the
pattern is and what goes wrong. Remove an entry once its fix is merged.

## Errors

- **Log and return.** `schaledb/students.go` logs a failed upload with WARN and
  then returns the same error, so the run reports it twice and at the wrong
  level. A function that returns an error should not also log it.
- **A nil result that hides its cause.** `analysis.DownloadPartyData` logs and
  returns nil on any failure. `party.processCachedRaid` then reports every
  failure as "not found", a Supabase timeout included. Return the error instead.
- **Ignored flag errors.** `cmd/process_raid.go` and `cmd/total_analysis.go`
  discard the error from `cmd.Flags().GetBool` and `GetInt`.
- **Usage on runtime errors.** Cobra prints the command's usage after any
  error a `RunE` returns, so a failed upload ends with a help text. Set
  `SilenceUsage` on the root command.

## Partial runs

- **One portrait aborts the student sync.** In `ParseSchaleDBStudents`, a failed
  portrait upload returns at once. The students written before it stay updated,
  and the student maps are never uploaded. Decide which steps are fatal and
  report the rest at the end.

## Context and time limits

- **`context.Background()` everywhere.** No call receives the command's context,
  so a run cannot be cancelled as a whole. Pass `cmd.Context()` down from each
  command.
- **Unbounded startup.** go-common's `postgres.InitFromEnv` pings the database
  with no deadline. A database that accepts the connection and then stops
  answering still hangs the job before the first query.
- **Fixed sleeps.** `storage.UploadFile` sleeps two seconds after every upload,
  and `GetDataFromURL` sleeps two seconds between retries, so every portrait in
  the student sync adds two seconds whether Supabase needs the pause or not.

## Structure

- **Configuration read at package load.** `analysis.PartyDataBaseURL` reads the
  environment when the package is loaded. A value set later in the process is
  never seen.
- **Misnamed type.** `JapaneseStudentInfo` also holds the English and Chinese
  names.
- **Stale documentation.** `docs/README.md` points at `deploy.yaml`, which no
  longer exists, and describes a Gemini video analysis this repository does not
  run.
