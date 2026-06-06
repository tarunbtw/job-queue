package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type DB struct {
	Conn *sql.DB
}

func New(path string) *DB {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		log.Fatal("failed to open db:", err)
	}
	if err := conn.Ping(); err != nil {
		log.Fatal("failed to ping db:", err)
	}
	d := &DB{Conn: conn}
	d.migrate()
	return d
}

func (d *DB) migrate() {
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id          TEXT PRIMARY KEY,
		type        TEXT NOT NULL,
		payload     TEXT NOT NULL,
		status      TEXT NOT NULL DEFAULT 'pending',
		attempts    INTEGER NOT NULL DEFAULT 0,
		max_attempts INTEGER NOT NULL DEFAULT 3,
		error       TEXT,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := d.Conn.Exec(schema)
	if err != nil {
		log.Fatal("migration failed:", err)
	}
	log.Println("database ready")
}