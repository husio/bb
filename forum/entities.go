package forum

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

func (u *User) Slug() string {
	return slugify(u.Login)
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

func (c *Category) Slug() string {
	return slugify(c.Name)
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

func (t *Topic) Slug() string {
	return slugify(t.Title)
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

const maxSlugLen = 140

func slugify(s string) string {
	s = slugrx.ReplaceAllString(strings.ToLower(s), "-")
	if len(s) > maxSlugLen {
		s = s[:maxSlugLen]
	}
	return strings.Trim(s, "-")
}

var slugrx = regexp.MustCompile("[^a-z0-9-]+")
