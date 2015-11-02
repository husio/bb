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
	<link rel="stylesheet" href="//necolas.github.io/normalize.css/3.0.2/normalize.css">
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
		<div class="">
			<div class="">
				<a href="/t/">Create topic</a>
			</div>

			<div class="">
				{{if .Topics}}
					<div class="">
						{{template "simple-pagination" .Pagination}}
					</div>

					{{range .Topics}}
						<div class="">
							<a href="/t/{{.TopicID}}/{{.Title|slugify}}">{{.Title}}</a>
							{{.Replies}}
							<a href="/t/{{.TopicID}}/{{.Title|slugify}}?page={{.Pages}}">last page</a>
							by <em>{{.User.Name}}</em>
							{{.Created}} &bull; {{.Updated}}
						</div>
					{{end}}

					<div class="">
						{{template "simple-pagination" .Pagination}}
					</div>
				{{else}}
					<div class="">
						no topics
					</div>
				{{end}}
			</div>
		</div>
	</body>
</html>
{{end}}


{{define "page-create-topic"}}
	{{template "page-header" .}}
	</head>
	<body>
		<div class="">
			<form action="." method="POST" enctype="multipart/form-data" class="">
				<fieldset>
					<label for="title">Title</label>
					<input type="text" name="title" id="title" class="" required>
					<label for="content">Content</label>
					<textarea name="content" id="content" class="" required></textarea>
					<button type="submit">Create</button>
				</fieldset>
			</form>
		</div>
	</body>
</html>
{{end}}


{{define "page-message-list"}}
	{{template "page-header" .}}
	</head>
	<body>
		<div class="">
			<div class="">
				<a href="/">Topics</a> &raquo;
				{{.Topic.Title}}
			</div>

			<div class="">
				{{template "pagination" .Paginator}}
			</div>

			{{range .Messages}}
				<div id="message-{{.MessageID}}" class="">
					<strong>{{.User.Name}}</strong>
					{{.Message.Content}}
					{{.Message.Created}}
				</div>
			{{end}}

			<div  class="">
				<form action="." method="POST" enctype="multipart/form-data">
					<div>
						<textarea name="content" required></textarea>
					</div>
					<button type="submit">Post</button>
				</form>
			</div>
		</div>
	</body>
</html>
{{end}}


{{define "pagination"}}
	{{if .IsFirst}}
		<span>&laquo; first</span>
	{{else}}
		<a href="./?page={{.FirstPage}}">&laquo; first</a>
	{{end}}
	{{if .HasPrev}}
		<a href="./?page={{.PrevPage}}">&lsaquo; previous</a>
	{{else}}
		<span>&lsaquo; previous</span>
	{{end}}
	|
	{{if .HasNext}}
		<a href="./?page={{.NextPage}}">next &rsaquo;</a>
	{{else}}
		<span>next &rsaquo;</span>
	{{end}}
	{{if .IsLast}}
		<span>last &raquo;</span>
	{{else}}
		<a href="./?page={{.LastPage}}">last &raquo;</a>
	{{end}}
{{end}}



{{define "simple-pagination"}}
	{{if .IsFirst}}
		<span>&laquo; first</span>
	{{else}}
		<a href="./">&laquo; first</a>
	{{end}}
	|
	{{if .HasNext}}
		<a href="./?page={{.NextPage}}">next &rsaquo;</a>
	{{else}}
		<span>next &rsaquo;</span>
	{{end}}
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
