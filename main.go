package main

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"text/template"
)

//go:embed templates/*.html
var embedded embed.FS
var templates = make(map[string]*template.Template)

func main() {
	base, err := embedded.ReadFile("templates/base.html")
	if err != nil {
		panic(err)
	}
	templates["base"] = template.New("base")
	templates["base"].Parse(string(base))

	fs, err := embedded.ReadDir("templates")
	if err != nil {
		panic(err)
	}
	for _, f := range fs {
		name := f.Name()
		if name == "base.html" {
			continue
		}

		data, err := embedded.ReadFile(fmt.Sprintf("templates/%s", name))
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

	http.Handle("/", handle(render))
	http.ListenAndServe(":4120", nil)
}

func handle(next http.HandlerFunc, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(methods) < 1 {
			methods = []string{"GET", "HEAD"}
		}

		for _, m := range methods {
			if m == r.Method {
				next(w, r)
				return
			}
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
	})
}

func render(w http.ResponseWriter, r *http.Request) {
	p := "sites" + r.URL.Path
	info, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			renderGeneric(w, http.StatusNotFound)
		} else {
			renderGeneric(w, http.StatusInternalServerError)
		}
		return
	}

	if !info.IsDir() {
		err = renderFile(w, p)
	} else {
		if !strings.HasSuffix(p, "/") {
			w.Header().Set("location", r.URL.Path+"/")
			w.WriteHeader(http.StatusPermanentRedirect)
			return
		}
		err = renderDir(w, p)
	}

	if err != nil {
		panic(err)
	}
}

func renderGeneric(w http.ResponseWriter, code int) error {
	w.WriteHeader(code)
	return templates["generic.html"].Execute(w, http.StatusText(code))
}

func renderFile(w io.Writer, path string) error {
	d, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return templates["single.html"].Execute(w, string(d))
}

func renderDir(w io.Writer, path string) error {
	list, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	return templates["list.html"].Execute(w, list)
}
