package main

import (
	"embed"
	"errors"
	"fmt"
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

	http.Handle("/", handle(listSites))
	http.ListenAndServe(":4120", nil)
}

func handle(next http.HandlerFunc, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(methods) < 1 && r.Method == "GET" {
			next(w, r)
			return
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

func listSites(w http.ResponseWriter, r *http.Request) {
	p := "sites" + r.URL.Path
	info, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			panic(err)
		}
		return
	}

	if info.IsDir() {
		if !strings.HasSuffix(p, "/") {
			w.Header().Set("location", r.URL.Path+"/")
			w.WriteHeader(http.StatusPermanentRedirect)
			return
		}

		list, err := os.ReadDir(p)
		if err != nil {
			panic(err)
		}
		err = templates["list.html"].Execute(w, list)
		if err != nil {
			panic(err)
		}
	} else {
		d, err := os.ReadFile(p)
		if err != nil {
			panic(err)
		}

		err = templates["single.html"].Execute(w, string(d))
		if err != nil {
			panic(err)
		}
	}
}
