package main

import (
	"math"
	"regexp"
	"strings"
	"time"
)

type User struct {
	UserID uint64 `db:"user_id"`
	Name   string `db:"name"`
}

type Topic struct {
	TopicID  uint      `db:"topic_id"`
	Title    string    `db:"title"`
	AuthorID uint      `db:"author_id"`
	Created  time.Time `db:"created"`
	Updated  time.Time `db:"updated"`
	Replies  uint      `db:"replies"`
}

var slugrx = regexp.MustCompile("[^a-z0-9-]+")

func (t *Topic) Slug() string {
	s := t.Title
	if len(s) > 140 {
		s = s[:100]
	}
	s = slugrx.ReplaceAllString(strings.ToLower(s), "-")
	return strings.Trim(s, "-")
}

func (t *Topic) Pages() uint {
	return uint(math.Ceil(float64(t.Replies+1) / float64(PageSize)))
}

type TopicWithUser struct {
	Topic
	User
}

type Message struct {
	MessageID uint      `db:"message_id"`
	AuthorID  uint      `db:"author_id"`
	TopicID   uint      `db:"topic_id"`
	Content   string    `db:"content"`
	Created   time.Time `db:"created"`
}

type MessageWithUser struct {
	Message
	User
}
