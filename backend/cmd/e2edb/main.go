// Command e2edb creates or drops a disposable authorization-test database.
// It intentionally refuses names outside its dedicated prefix.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

const safePrefix = "go_pos_authz_e2e_"

func main() {
	action := flag.String("action", "create", "create or drop")
	name := flag.String("name", "", "disposable database name")
	flag.Parse()

	if !strings.HasPrefix(*name, safePrefix) || len(*name) <= len(safePrefix) {
		log.Fatalf("refusing unsafe database name %q; expected prefix %q", *name, safePrefix)
	}
	if *action != "create" && *action != "drop" {
		log.Fatalf("unsupported action %q", *action)
	}

	_ = godotenv.Load()
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"), os.Getenv("DB_SSLMODE"))
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	quoted := pgx.Identifier{*name}.Sanitize()
	if *action == "create" {
		var exists bool
		if err := conn.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname=$1)`, *name).Scan(&exists); err != nil {
			log.Fatal(err)
		}
		if exists {
			log.Fatalf("database %q already exists; refusing to reuse a possibly non-disposable target", *name)
		}
		_, err = conn.Exec(ctx, "CREATE DATABASE "+quoted)
	} else {
		_, _ = conn.Exec(ctx, `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()`, *name)
		_, err = conn.Exec(ctx, "DROP DATABASE IF EXISTS "+quoted)
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%sd disposable database %s", *action, *name)
}
