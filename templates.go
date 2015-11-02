package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var slugrx = regexp.MustCompile("[^a-z0-9]+")

func slugify(s string) string {
	return strings.Trim(slugrx.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{
	"slugify": slugify,
}).Parse(`

{{define "page-header"}}
<!DOCTYPE html>
<html>
<head>
{{end}}


{{define "page-error"}}
	{{template "page-header" .}}
	</head>
	<body>
		{{.Text}}
	</body>
	</html>
{{end}}


{{define "page-topic-list"}}
	{{template "page-header" .}}
	</head>
	<body>
		<a href="/t/">Create topic</a>
		<hr>
		{{if .Topics}}
			<div>
				<a href="./">&laquo; first page</a> | <a href="./?off={{.NextPageOff}}">next page &raquo;</a>
			</div>
			{{range .Topics}}
				<div>
					<a href="/t/{{.TopicID}}/{{.Title|slugify}}">{{.Title}}</a>
					{{.Replies}}
				</div>
			{{end}}
			<div>
				<a href="./">&laquo; first page</a> | <a href="./?off={{.NextPageOff}}">next page &raquo;</a>
			</div>
		{{else}}
			<div>
				no topics
			</div>
		{{end}}
	</body>
</html>
{{end}}


{{define "page-create-topic"}}
	{{template "page-header" .}}
	</head>
	<body>
		<form action="." method="POST" enctype="multipart/form-data">
			<div>
				<input type="text" name="title" required>
			</div>
			<div>
				<textarea name="content" required></textarea>
			</div>
			<button type="submit">Create</button>
		</form>
	</body>
</html>
{{end}}


{{define "page-message-list"}}
	{{template "page-header" .}}
	</head>
	<body>
		<h1>
			{{.Topic.Title}}
			<small>{{.Topic.Replies}}</small>
		</h1>

		{{range .Messages}}
			<div id="message-{{.MessageID}}">
				<strong>{{.User.Name}}</strong>
				{{.Message.Content}}
				{{.Message.Created}}
			</div>
		{{end}}

		<form action="." method="POST" enctype="multipart/form-data">
			<div>
				<textarea name="content" required></textarea>
			</div>
			<button type="submit">Post</button>
		</form>
	</body>
</html>
{{end}}
`))

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
	renderTo(w, "page-error", ctx)
}

func Render400(w http.ResponseWriter, text string) {
	ctx := errcontext{
		Code: http.StatusBadRequest,
		Text: text,
	}
	w.WriteHeader(http.StatusBadRequest)
	renderTo(w, "page-error", ctx)
}

func Render404(w http.ResponseWriter, text string) {
	ctx := errcontext{
		Code: http.StatusBadRequest,
		Text: text,
	}
	w.WriteHeader(http.StatusNotFound)
	renderTo(w, "page-error", ctx)
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
		if err := renderTo(&b, "page-error", ctx); err != nil {
			panic(err)
		}
	}
	w.WriteHeader(code)
	b.WriteTo(w)
}
