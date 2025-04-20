# Commands for local development

## Prerequisites

- [Go](https://go.dev/) 1.23+

## Go Commands

### Run the program

```bash
go run main.go
```

### Run the whole test in `app/tests` directory

```bash
go test ./app/tests
```

### Run the test in `app/tests` directory with verbose output

```bash
go test -v ./app/tests
```

### Run the specific test in `app/tests` directory

For example:

```bash
go test -run ./app/tests/arona_ai_test.go
```
