package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Feed struct {
	Title   string   `xml:"title"`
	Entries []*Entry `xml:"entry"`
}

type Entry struct {
	Title   string    `xml:"title"`
	Summary string    `xml:"summary"`
	Updated time.Time `xml:"udpated"`
}

func main() {
	urlFl := flag.String("url", "https://bbs.archlinux.org/extern.php?action=feed&type=atom", "Feed URL")
	dbUserFl := flag.String("dbuser", "bb", "Database user name")
	dbPassFl := flag.String("dbpass", "bb", "Database user password")
	dbNameFl := flag.String("dbname", "bb", "Database name")
	repeatFl := flag.Int("repeat", 1, "Repeat data")

	flag.Parse()

	cred := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", *dbUserFl, *dbPassFl, *dbNameFl)
	db, err := sqlx.Connect("postgres", cred)
	if err != nil {
		log.Fatalf("cannot connect to database: %s", err)
	}
	defer db.Close()

	resp, err := http.Get(*urlFl)
	if err != nil {
		log.Fatalf("cannot fetch resource: %s", err)
	}
	defer resp.Body.Close()

	upload(resp.Body, db, *repeatFl)
}

func upload(resource io.Reader, db *sqlx.DB, repeat int) {
	var feed Feed
	if err := xml.NewDecoder(resource).Decode(&feed); err != nil {
		log.Fatalf("cannot decode resource: %s", err)
	}

	tx, err := db.Beginx()
	if err != nil {
		log.Fatalf("cannot start transaction :%s", err)
	}
	defer tx.Rollback()

	for i := 0; i < repeat; i++ {
		for _, e := range feed.Entries {
			var tid uint
			err = tx.Get(&tid, `
			INSERT INTO topics (title, author_id, created, updated, category_id)
			VALUES ($1, $2, $3, $3, 1)
			RETURNING topic_id
		`, e.Title, 1, e.Updated)
			if err != nil {
				log.Fatalf("cannot insert topic: %s", err)
			}
			_, err = tx.Exec(`
			INSERT INTO messages (topic_id, author_id, content, created)
			VALUES ($1, $2, $3, $4)
		`, tid, 1, e.Summary, e.Updated)
			if err != nil {
				log.Fatalf("cannot insert message: %s", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatalf("cannot commit transaction: %s", err)
	}
}

func sanitizeHTML(s string) string {
	return html.EscapeString(s)
}
