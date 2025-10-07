package main

import (
	"fmt"
	"log"

	logic_duckdb "ba-torment-data-process/internal/logic/duckdb"
)

func main() {
	if err := logic_duckdb.ParseDuckDB(); err != nil {
		log.Fatal(fmt.Errorf("failed to parse duckdb: %w", err))
	}
	fmt.Println("Successfully parsed DuckDB data")
}
