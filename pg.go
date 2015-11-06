package main

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/net/context"

	"github.com/lib/pq"
)

func WithPG(ctx context.Context, credentials string) (context.Context, error) {
	db, err := sqlx.Connect("postgres", credentials)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, "db:connection", db), nil
}

func DB(ctx context.Context) *sqlx.DB {
	return ctx.Value("db:connection").(*sqlx.DB)
}

func NewStore(c dbconn) *store {
	return &store{db: c}
}

type store struct {
	db dbconn
}

type dbconn interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

func (s *store) UserByID(userID uint) (*User, error) {
	var u User
	err := s.db.Get(&u, `SELECT * FROM users WHERE user_id = $1`, userID)
	return &u, transformErr(err)
}

func (s *store) LastTopicUpdated(updatedGte time.Time) (time.Time, error) {
	var t time.Time
	err := s.db.Get(&t, `
		SELECT updated FROM topics WHERE updated < $1
		ORDER BY updated DESC
		LIMIT 1
	`, updatedGte)
	return t, transformErr(err)
}

func (s *store) Topics(updatedGte time.Time, limit uint) ([]*TopicWithUserCategory, error) {
	var topics []*TopicWithUserCategory
	err := s.db.Select(&topics, `
		SELECT t.*, u.*, c.*
		FROM topics t
			INNER JOIN users u ON t.author_id = u.user_id
			INNER JOIN categories c ON t.category_id = c.category_id
		WHERE t.updated < $1
		ORDER BY t.updated DESC LIMIT $2
	`, updatedGte, limit)
	return topics, transformErr(err)
}

func (s *store) CreateTopic(title string, author, category uint, now time.Time) (*Topic, error) {
	var t Topic
	err := s.db.Get(&t, `
		INSERT INTO topics (title, author_id, category_id, created, updated, replies)
		VALUES ($1, $2, $3, $4, $4, 0)
		RETURNING *
	`, title, author, category, now)
	return &t, transformErr(err)
}

func (s *store) TopicByID(topicID uint) (*TopicWithUserCategory, error) {
	var t TopicWithUserCategory
	err := s.db.Get(&t, `
		SELECT t.*, u.*, c.*
		FROM topics t
			INNER JOIN users u ON t.author_id = u.user_id
			INNER JOIN categories c ON t.category_id = c.category_id
		WHERE topic_id = $1
		LIMIT 1
		`, topicID)
	return &t, transformErr(err)
}

func (s *store) TopicMessages(topicID uint, offset, limit uint) ([]*MessageWithUser, error) {
	var messages []*MessageWithUser
	err := s.db.Select(&messages, `
		SELECT m.*, u.*
		FROM messages m
			INNER JOIN users u ON m.author_id = u.user_id
		WHERE m.topic_id = $1
		ORDER BY m.created ASC OFFSET $2 LIMIT $3
	`, topicID, offset, limit)
	return messages, transformErr(err)
}

func (s *store) CreateMessage(topic, author uint, content string, now time.Time) (*Message, error) {
	var m Message
	err := s.db.Get(&m, `
		INSERT INTO messages (topic_id, author_id, content, created)
		VALUES ($1, $2, $3, $4)
		RETURNING *
	`, topic, author, content, now)
	return &m, transformErr(err)
}

func (s *store) Categories() ([]*Category, error) {
	var cats []*Category
	err := s.db.Select(&cats, `SELECT * FROM categories LIMIT 1000`)
	return cats, transformErr(err)
}

var (
	ErrConflict = errors.New("conflict")
	ErrNotFound = errors.New("not found")
)

func transformErr(err error) error {
	if err == nil {
		return nil
	}
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err, ok := err.(*pq.Error); ok && err.Code == "23505" {
		return ErrConflict
	}
	return err
}
