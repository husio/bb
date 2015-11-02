package main

import "time"

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
