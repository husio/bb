package main

import (
	"bytes"
	"html/template"
	"io"
	"log"
	"net/http"
)

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{}).Parse(`

{{define "page-header"}}
<!DOCTYPE html>
<html lang="en">
<html>
<head>
    <meta charset="utf-8">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta http-equiv="x-ua-compatible" content="ie=edge">
	<link rel="stylesheet" href="https://cdn.rawgit.com/twbs/bootstrap/v4-dev/dist/css/bootstrap.css">
{{end}}


{{define "page-error"}}
	{{template "page-header" .}}
	</head>
	<body>
		<div class="container-fluid">
			<div class="row">
				<div class="col-md-12">
					<div class="alert alert-danger" role="alert">
						{{.Text}}
					</div>
				</div>
			</div>
		</div>
	</body>
	</html>
{{end}}


{{define "page-topic-list"}}
	{{template "page-header" .}}
	</head>
	<body>
		<div class="container-fluid">
			<div class="row">
				<div class="col-md-12">
					<a href="/t/">Create topic</a>
				</div>
			</div>

			{{if .Topics}}
				<div class="row">
					<div class="col-md-12">
						{{template "simple-pagination" .Pagination}}
					</div>
				</div>

				{{range .Topics}}
					<div class="row">
						<div class="row">
							<div class="col-md-12">
								<strong>
									<a href="/t/{{.TopicID}}/{{.Slug}}/">{{.Title}}</a>
								</strong>
							</div>
						</div>
						<div class="row">
							<div class="col-md-12">
								By {{.User.Name}} - <span class="text-muted">{{.Replies}} replies - X views</span>
								<a href="/t/{{.TopicID}}/{{.Slug}}/?page={{.Pages}}">last page</a>
								<span class="pull-right">{{.Updated.Format "_2 Jan 2006"}}</span>
							</div>
						</div>
					</div>
				{{end}}

				<div class="row">
					<div class="col-md-12">
						{{template "simple-pagination" .Pagination}}
					</div>
				</div>
			{{else}}
				<div class="row">
					<div class="col-md-12">
						no topics
					</div>
				</div>
			{{end}}
		</div>
	</body>
</html>
{{end}}


{{define "page-create-topic"}}
	{{template "page-header" .}}
	</head>
	<body>
		<div class="container-fluid">
			<div class="row">
				<div class="col-md-12">
					<form action="." method="POST" enctype="multipart/form-data" class="">
						<fieldset class="form-group">
							<label for="title">Title</label>
							<input class="form-control" type="text" name="title" id="title" class="" required>
						</fieldset>
						<fieldset class="form-group">
							<label for="content">Content</label>
							<textarea class="form-control" name="content" id="content" class="" required></textarea>
						</fieldset>
						<div class="pull-right">
							<a href="/" class="btn btn-link" type="button">Back to main page</a>
							<button class="btn btn-primary" type="submit">Submit</button>
						</div>
					</form>
				</div>
			</div>
		</div>
	</body>
</html>
{{end}}


{{define "page-message-list"}}
	{{template "page-header" .}}
	</head>
	<body>
		<div class="container-fluid">
			<div class="row">
				<div class="col-md-12">
					<a href="/">Topics</a> &raquo;
					{{.Topic.Title}}
				</div>
			</div>

			<div class="row">
				<div class="col-md-12">
						{{template "pagination" .Paginator}}
				</div>
			</div>

			{{range .Messages}}
				<div class="row">
					<div class="col-md-12" id="m{{.MessageID}}">
						<a href="./#m{{.MessageID}}">#{{.CollectionPos}}</a>
						<strong>{{.User.Name}}</strong>
						{{.Message.Content}}
						{{.Message.Created}}
					</div>
				</div>
			{{end}}

			<div class="row">
				<div class="col-md-12">
						{{template "pagination" .Paginator}}
				</div>
			</div>

			{{if .Paginator.IsLast}}
				<div class="row">
					<div class="col-md-12">
						<form action="." method="POST" enctype="multipart/form-data">
							<fieldset class="form-group">
								<textarea class="form-control" name="content" required></textarea>
							</fieldset>
							<button class="btn btn-primary-outline btn-sm pull-right" type="submit">Submit</button>
						</form>
					</div>
				</div>
			{{else}}
				<div class="row">
					<div class="col-md-4 col-md-offset-4 alert alert-info center">
						Go to <a href="?page={{.Paginator.LastPage}}">last page</a> to comment.
					</div>
				</div>
			{{end}}

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
