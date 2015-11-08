package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

var tmpl interface {
	ExecuteTemplate(io.Writer, string, interface{}) error
}

var tNoCache = devmode()

func loadTemplates() error {
	const tmplglob = "assets/templates/*html"

	var err error
	if tNoCache {
		tmpl, err = newDynamicTemplateLoader(tmplglob)
	} else {
		t := template.New("").Funcs(tmplFuncs)
		tmpl, err = t.ParseGlob(tmplglob)
	}
	return err
}

type dynamicTemplateLoader struct {
	mu   sync.Mutex
	glob string
	t    *template.Template
}

func newDynamicTemplateLoader(glob string) (*dynamicTemplateLoader, error) {
	t, err := template.New("").Funcs(tmplFuncs).ParseGlob(glob)
	if err != nil {
		return nil, err
	}
	dl := &dynamicTemplateLoader{
		glob: glob,
		t:    t,
	}
	go dl.hotUpdate()
	return dl, nil
}

func (dl *dynamicTemplateLoader) hotUpdate() {
	dir := dl.glob
	for len(dir) > 0 && dir[len(dir)-1] != '/' {
		dir = dir[:len(dir)-1]
	}
	if len(dir) == 0 {
		panic("invalid directory")
	}

	lastMod := time.Now()
	for {
		time.Sleep(2 * time.Second)

		f, err := os.Stat(dir)
		if err != nil {
			log.Printf("cannot stat %q directory: %s", dir, err)
			continue
		}

		if mtime := f.ModTime(); mtime.After(lastMod) {
			dl.mu.Lock()
			t := template.New("").Funcs(tmplFuncs)
			if t, err := t.ParseGlob(dl.glob); err != nil {
				log.Printf("cannot parse templates: %s", err)
			} else {
				dl.t = t
			}
			dl.mu.Unlock()
			lastMod = mtime
		}
	}
}

func (dl *dynamicTemplateLoader) ExecuteTemplate(w io.Writer, name string, ctx interface{}) error {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	return dl.t.ExecuteTemplate(w, name, ctx)
}

func renderTo(w io.Writer, name string, context interface{}) error {
	return tmpl.ExecuteTemplate(w, name, context)
}

type errcontext struct {
	Code int
	Text string
}

func Render500(w http.ResponseWriter, err error) {
	log.Printf("error: %s", err)
	ctx := errcontext{
		Code: http.StatusInternalServerError,
		Text: http.StatusText(http.StatusInternalServerError),
	}
	w.WriteHeader(http.StatusInternalServerError)
	renderTo(w, "page_error", ctx)
}

func Render400(w http.ResponseWriter, text string) {
	ctx := errcontext{
		Code: http.StatusBadRequest,
		Text: text,
	}
	w.WriteHeader(http.StatusBadRequest)
	renderTo(w, "page_error", ctx)
}

func Render404(w http.ResponseWriter, text string) {
	ctx := errcontext{
		Code: http.StatusBadRequest,
		Text: text,
	}
	w.WriteHeader(http.StatusNotFound)
	renderTo(w, "page_error", ctx)
}

func Render(w http.ResponseWriter, code int, name string, context interface{}) {
	var b bytes.Buffer
	if err := renderTo(&b, name, context); err != nil {
		log.Printf("cannot render %q template: %s", name, err)
		code = http.StatusInternalServerError
		b.Reset()
		ctx := errcontext{
			Code: code,
			Text: http.StatusText(code),
		}
		if err := renderTo(&b, "page_error", ctx); err != nil {
			panic(err)
		}
	}
	w.WriteHeader(code)
	b.WriteTo(w)
}

var tmplFuncs = template.FuncMap{
	"markdown": markdown,
}

func markdown(s string) template.HTML {
	unsafe := blackfriday.MarkdownCommon([]byte(s))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	return template.HTML(html)
}
