package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://careergps:secret@localhost:5432/careergps?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect failed: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ping failed: %v\n", err)
		os.Exit(1)
	}

	for _, f := range []string{"seed/skills.sql", "seed/companies.sql"} {
		sql, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", f, err)
			os.Exit(1)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			fmt.Fprintf(os.Stderr, "exec %s: %v\n", f, err)
			os.Exit(1)
		}
		fmt.Printf("seeded %s\n", f)
	}

	fmt.Println("seed complete")
}
