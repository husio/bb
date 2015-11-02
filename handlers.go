package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
)

func handleCreateTopic(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	uid, ok := CurrentUserID(r)
	if !ok {
		// TODO - redirect to authentication page, but remember form content
		Render500(w, errors.New("not implemented"))
		return
	}

	var c struct {
		Title      string
		TitleErr   string
		Content    string
		ContentErr string
	}

	if r.Method == "GET" {
		Render(w, http.StatusOK, "page-create-topic", c)
		return
	}

	if err := r.ParseMultipartForm(2 << 20); err != nil {
		Render400(w, err.Error())
		return
	}
	c.Content = strings.TrimSpace(r.FormValue("content"))
	c.Title = strings.TrimSpace(r.FormValue("title"))

	if len(c.Title) < 3 {
		c.TitleErr = "Title must be at least 3 characters long"
	}
	if len(c.Title) > 200 {
		c.TitleErr = "Title must not be longer than 200 characters"
	}
	if len(c.Content) < 3 {
		c.ContentErr = "Content must be at least 3 characters long"
	}
	if len(c.Content) > 10000 {
		c.ContentErr = "Content must be shorter than 10000 characters"
	}
	if c.TitleErr != "" || c.ContentErr != "" {
		Render(w, http.StatusBadRequest, "page-create-topic", c)
		return
	}

	tx, err := DB(ctx).Beginx()
	if err != nil {
		Render500(w, err)
		return
	}
	defer tx.Rollback()
	store := NewStore(tx)
	now := time.Now()
	topic, err := store.CreateTopic(c.Title, uid, now)
	if err != nil {
		Render500(w, err)
		return
	}
	if _, err := store.CreateMessage(topic.TopicID, uid, c.Content, now); err != nil {
		Render500(w, err)
		return
	}
	if err := tx.Commit(); err != nil {
		Render500(w, err)
		return
	}
	turl := fmt.Sprintf("/t/%d/%s", topic.TopicID, slugify(topic.Title))
	http.Redirect(w, r, turl, http.StatusFound)
}

func handleListTopics(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	store := NewStore(DB(ctx))

	lim, off := paginate(r.URL.Query(), 50)
	topics, err := store.Topics(lim, off)
	if err != nil {
		Render500(w, err)
		return
	}
	c := struct {
		Topics []*Topic
	}{
		Topics: topics,
	}
	Render(w, http.StatusOK, "page-topic-list", c)
}

func handleCreateMessage(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	uid, ok := CurrentUserID(r)
	if !ok {
		// TODO - redirect to authentication page, but remember form content
		Render500(w, errors.New("not implemented"))
		return
	}

	tid, err := strconv.Atoi(Param(ctx, "topic"))
	if err != nil || tid < 0 {
		Render404(w, "Topic does not exist")
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	if len(content) < 3 {
		Render400(w, "Message too short")
		return
	}
	if len(content) > 20000 {
		Render400(w, "Message too long")
		return
	}

	store := NewStore(DB(ctx))
	m, err := store.CreateMessage(uint(tid), uid, content, time.Now())
	if err != nil {
		Render404(w, err.Error())
		return
	}

	// TODO redirect to message page
	murl := fmt.Sprintf("/t/%d/#message-%d", m.TopicID, m.MessageID)
	http.Redirect(w, r, murl, http.StatusFound)
}

func handleTopicMessages(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tx, err := DB(ctx).Beginx()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	store := NewStore(tx)

	topicID, err := strconv.Atoi(Param(ctx, "topic"))
	if err != nil || topicID < 0 {
		Render404(w, "Topic does not exist")
		return
	}
	topic, err := store.TopicByID(uint(topicID))
	if err == ErrNotFound {
		Render404(w, "Topic does not exist")
		return
	}
	if err != nil {
		Render500(w, err)
		return
	}

	lim, off := paginate(r.URL.Query(), 50)
	messages, err := store.TopicMessages(topic.TopicID, lim, off)
	if err != nil {
		Render500(w, err)
		return
	}

	c := struct {
		Topic    *Topic
		Messages []*MessageWithUser
	}{
		Topic:    topic,
		Messages: messages,
	}
	Render(w, http.StatusOK, "page-message-list", c)
}

func paginate(q url.Values, pageSize uint) (limit, offset uint) {
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	return pageSize, uint(page-1) * pageSize
}