package main

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
)

type User struct {
	UserID uint64 `db:"user_id"`
	Login  string `db:"login"`
}

type Category struct {
	CategoryID  uint   `db:"category_id"`
	Name        string `db:"name"`
	Description string `db:"description"`
	TopicsCount uint   `db:"topics_count"`
	Color       uint   `db:"color"`
}

func (c *Category) ColorHex() string {
	return fmt.Sprintf("%.6x", 0xFFFFFF&c.Color)
}

func (c *Category) SetColorHex(x string) {
	panic("not implemented")
}

type Topic struct {
	TopicID    uint      `db:"topic_id"`
	Title      string    `db:"title"`
	AuthorID   uint      `db:"author_id"`
	CategoryID uint      `db:"category_id"`
	Created    time.Time `db:"created"`
	Updated    time.Time `db:"updated"`
	Replies    uint      `db:"replies"`
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

type TopicWithUserCategory struct {
	Topic
	User
	Category
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
