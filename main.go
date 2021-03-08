package main

import (
	"embed"
	"fmt"
	"net/http"
	"text/template"
)

//go:embed templates/*.html
var templatesDir embed.FS
var templates = make(map[string]*template.Template)

func main() {
	base, err := templatesDir.ReadFile("templates/base.html")
	if err != nil {
		panic(err)
	}
	templates["base"] = template.New("base")
	templates["base"].Parse(string(base))

	fs, err := templatesDir.ReadDir("templates")
	if err != nil {
		panic(err)
	}
	for _, f := range fs {
		name := f.Name()
		if name == "base.html" {
			continue
		}

		data, err := templatesDir.ReadFile(fmt.Sprintf("templates/%s", name))
		if err != nil {
			panic(err)
		}

		templates[name], err = templates["base"].Clone()
		if err != nil {
			panic(err)
		}
		_, err = templates[name].Parse(string(data))
		if err != nil {
			panic(err)
		}
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":4120", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}
