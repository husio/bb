package main

import (
	"errors"
	"fmt"
	"net/http"
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
		Title       string
		TitleErr    string
		Category    uint
		CategoryErr string
		Categories  []*Category
		Content     string
		ContentErr  string
	}

	if r.Method == "GET" {
		if cats, err := NewStore(DB(ctx)).Categories(); err != nil {
			Render500(w, err)
		} else {
			c.Categories = cats
			Render(w, http.StatusOK, "page_create_topic", c)
		}
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
	if raw := r.FormValue("category"); raw == "" {
		c.CategoryErr = "Category is required"
	} else {
		if cat, err := strconv.Atoi(r.FormValue("category")); err == nil {
			c.Category = uint(cat)
		} else {
			c.CategoryErr = "Invalid category"
		}
	}

	if c.TitleErr != "" || c.ContentErr != "" || c.CategoryErr != "" {
		if cats, err := NewStore(DB(ctx)).Categories(); err != nil {
			Render500(w, err)
		} else {
			c.Categories = cats
			Render(w, http.StatusBadRequest, "page_create_topic", c)
		}
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
	topic, err := store.CreateTopic(c.Title, uid, c.Category, now)
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
	turl := fmt.Sprintf("/t/%d/%s", topic.TopicID, topic.Slug())
	http.Redirect(w, r, turl, http.StatusFound)
}

func handleListTopics(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	store := NewStore(DB(ctx))

	p := NewSimplePaginator(time.Now())
	if sec, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
		p.Current = int(sec)
	}

	if t, err := store.LastTopicUpdated(time.Unix(int64(p.Current), 0)); err != nil {
		Render500(w, err)
		return
	} else if checkLastModified(w, r, t) {
		return
	}

	topics, err := store.Topics(time.Unix(int64(p.Current), 0), p.Limit())
	if err != nil {
		Render500(w, err)
		return
	}

	// if there are less topics than the page size, then this is the last page
	if len(topics) == PageSize {
		p.Next = int(topics[len(topics)-1].Updated.Unix())
	}

	c := struct {
		Topics     []*TopicWithUserCategory
		Pagination *SimplePaginator
	}{
		Topics:     topics,
		Pagination: p,
	}
	Render(w, http.StatusOK, "page_topic_list", c)
}

func handleCreateMessage(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	uid, ok := CurrentUserID(r)
	if !ok {
		// TODO - redirect to authentication page, but remember form content
		Render500(w, errors.New("not implemented"))
		return
	}

	tid, err := strconv.Atoi(Param(ctx, "topicid"))
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

	tx, err := DB(ctx).Beginx()
	if err != nil {
		Render500(w, err)
		return
	}
	defer tx.Rollback()

	store := NewStore(tx)

	t, err := store.TopicByID(uint(tid))
	if err != nil {
		if err == ErrNotFound {
			Render404(w, "Topic does not exist")
		} else {
			Render500(w, err)
		}
		return
	}

	m, err := store.CreateMessage(t.TopicID, uid, content, time.Now())
	if err != nil {
		Render500(w, err)
		return
	}

	if err := tx.Commit(); err != nil {
		Render500(w, err)
		return
	}

	murl := fmt.Sprintf(
		"/t/%d/%s?page=%d#m%d",
		t.TopicID, t.Topic.Slug(), t.Pages(), m.MessageID)
	http.Redirect(w, r, murl, http.StatusFound)
}

func handleListTopicMessages(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	tx, err := DB(ctx).Beginx()
	if err != nil {
		panic(err)
	}
	defer tx.Rollback()

	store := NewStore(tx)

	topicID, err := strconv.Atoi(Param(ctx, "topicid"))
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

	if checkLastModified(w, r, topic.Updated) {
		return
	}

	p := NewPaginator(r.URL.Query(), int(topic.Replies+1))
	messages, err := store.TopicMessages(topic.TopicID, p.Offset(), p.Limit())
	if err != nil {
		Render500(w, err)
		return
	}

	type MessageWithUserPos struct {
		*Message
		*User
		CollectionPos int // position number in messages collection
	}

	emsgs := make([]*MessageWithUserPos, 0, len(messages))
	for i, m := range messages {
		emsgs = append(emsgs, &MessageWithUserPos{
			CollectionPos: i + (p.CurrentPage()-1)*int(p.PageSize()) + 1,
			Message:       &m.Message,
			User:          &m.User,
		})
	}

	c := struct {
		Topic     *TopicWithUserCategory
		Messages  []*MessageWithUserPos
		Paginator *Paginator
	}{
		Topic:     topic,
		Messages:  emsgs,
		Paginator: p,
	}
	Render(w, http.StatusOK, "page_message_list", c)
}

func handleUserDetails(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func handleListCategories(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

var httpNoCache = devmode()

func checkLastModified(w http.ResponseWriter, r *http.Request, modtime time.Time) bool {
	if httpNoCache {
		return false
	}
	// https://golang.org/src/net/http/fs.go#L273
	ms, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since"))
	if err == nil && modtime.Before(ms.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return false
}
