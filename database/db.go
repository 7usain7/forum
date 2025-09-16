package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() {
	var err error

	DB, err = sql.Open("sqlite3", "database/forum.db")
	if err != nil {
		log.Fatal(err)
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)
	DB.SetConnMaxLifetime(0)

	if _, err = DB.Exec(`
	PRAGMA foreign_keys = ON;
	PRAGMA journal_mode = WAL;
	PRAGMA synchronous = NORMAL;
	`); err != nil {
		log.Fatal(err)
	}

	const schema = `
CREATE TABLE IF NOT EXISTS users(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT NOT NULL UNIQUE,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions(
  id TEXT PRIMARY KEY,
  user_id INTEGER NOT NULL,
  expires_at DATETIME NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS categories(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS posts(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_categories(
  post_id INTEGER NOT NULL,
  category_id INTEGER NOT NULL,
  PRIMARY KEY(post_id, category_id),
  FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS comments(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  post_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  body TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS likes(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  target_type TEXT NOT NULL CHECK(target_type IN ('post','comment')),
  target_id INTEGER NOT NULL,
  like_type INTEGER NOT NULL CHECK(like_type IN (-1,1)),
  UNIQUE(user_id, target_type, target_id),
  FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);
`

	if _, err := DB.Exec(schema); err != nil {
		log.Fatal(err)
	}

	if err := ensureBodyColumns(); err != nil {
		log.Fatal(err)
	}

	// Clean up expired sessions
	go func() {
		t := time.NewTicker(30 * time.Minute)
		defer t.Stop()
		for range t.C {
			DB.Exec(`DELETE FROM sessions WHERE expires_at < CURRENT_TIMESTAMP`)
		}
	}()
}

func CloseDB() {
	if DB != nil {
		_ = DB.Close()
	}
}

func ensureBodyColumns() error {
	if err := ensureColumn("posts", "body", "content"); err != nil {
		return err
	}
	if err := ensureColumn("comments", "body", "content"); err != nil {
		return err
	}
	return nil
}

func ensureColumn(table, want, legacy string) error {
	exists, err := columnExists(table, want)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if legacy != "" {
		legacyExists, err := columnExists(table, legacy)
		if err != nil {
			return err
		}
		if legacyExists {
			if _, err := DB.Exec(fmt.Sprintf(`ALTER TABLE %s RENAME COLUMN %s TO %s`, table, legacy, want)); err == nil {
				return nil
			}
			if _, err := DB.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s TEXT`, table, want)); err != nil {
				return err
			}
			_, err = DB.Exec(fmt.Sprintf(`UPDATE %s SET %s = %s WHERE %s IS NULL`, table, want, legacy, want))
			return err
		}
	}

	_, err = DB.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s TEXT`, table, want))
	return err
}

func columnExists(table, column string) (bool, error) {
	rows, err := DB.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid     int
			name    string
			ctype   string
			notnull int
			dflt    sql.NullString
			pk      int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}

	return false, rows.Err()
}
