package main

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/ghodss/yaml"
)

type templateData struct {
	Title   string
	Request *http.Request
	Data    interface{}
	Site    site
}

type site struct {
	Sections  []section
	Dir       string
	templates map[string]frontmatter
}

type section struct {
	Label     string
	Path      string
	Match     string
	Templates []string
}

type frontmatter struct {
	Label        string
	HideBody     bool
	DisplayField string
	Blocks       []block
}

type block struct {
	Kind        string
	Label       string
	Name        string
	Description string
	Config      map[string]interface{}
}

//go:embed templates/*.html
var embedded embed.FS
var templates = make(map[string]*template.Template)

func main() {
	err := parseTemplates()
	if err != nil {
		log.Fatalf("Failed to parse templates %s", err.Error())
	}

	http.Handle("/", handle(render))
	err = runServer()
	if err != nil {
		log.Fatal(err)
	}
}

func parseTemplates() error {
	base, err := embedded.ReadFile("templates/base.html")
	if err != nil {
		return err
	}
	templates["base"] = template.New("base")
	templates["base"].Parse(string(base))

	fs, err := embedded.ReadDir("templates")
	if err != nil {
		return err
	}
	for _, f := range fs {
		name := f.Name()
		if name == "base.html" {
			continue
		}

		data, err := embedded.ReadFile(fmt.Sprintf("templates/%s", name))
		if err != nil {
			return err
		}

		templates[name], err = templates["base"].Clone()
		if err != nil {
			return err
		}
		_, err = templates[name].Parse(string(data))
		if err != nil {
			return err
		}
	}

	return nil
}

func runServer() error {
	server := http.Server{
		Addr:    ":4120",
		Handler: http.DefaultServeMux,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Println("Server started on port " + server.Addr[1:])
	<-done
	log.Println("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
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

	d := templateData{
		Title:   strings.TrimSuffix(p, "/"),
		Request: r,
	}

	splitURL := strings.Split(r.URL.Path, "/")
	if len(splitURL[1]) > 0 {
		s := site{
			Dir:       splitURL[1],
			templates: make(map[string]frontmatter),
		}

		b, err := os.ReadFile(fmt.Sprintf("sites/%s/.cms/config.yaml", s.Dir))
		if err != nil {
			renderGeneric(w, http.StatusInternalServerError)
			return
		}

		err = yaml.Unmarshal(b, &s)
		if err != nil {
			renderGeneric(w, http.StatusInternalServerError)
			return
		}

		fs, err := os.ReadDir(fmt.Sprintf("sites/%s/.cms/templates", s.Dir))
		if err != nil {
			renderGeneric(w, http.StatusInternalServerError)
			return
		}

		for _, f := range fs {
			name := f.Name()
			b, err := os.ReadFile(fmt.Sprintf("sites/%s/.cms/templates/%s", s.Dir, name))
			if err != nil {
				renderGeneric(w, http.StatusInternalServerError)
				return
			}
			key := strings.TrimSuffix(name, path.Ext(name))
			t := new(frontmatter)
			yaml.Unmarshal(b, t)
			s.templates[key] = *t
		}

		d.Site = s
	}

	var buffer bytes.Buffer
	if !info.IsDir() {
		err = d.renderFile(&buffer, p)
	} else {
		if !strings.HasSuffix(p, "/") {
			w.Header().Set("location", r.URL.Path+"/")
			w.WriteHeader(http.StatusPermanentRedirect)
			return
		}
		err = d.renderDir(&buffer, p)
	}

	if err != nil {
		renderGeneric(w, http.StatusInternalServerError)
		log.Println(err)
	} else {
		w.Write(buffer.Bytes())
	}
}

func renderGeneric(w http.ResponseWriter, code int) error {
	w.WriteHeader(code)
	return templates["generic.html"].Execute(w, http.StatusText(code))
}

func (d *templateData) renderFile(w io.Writer, path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	d.Data = string(f)

	return templates["single.html"].Execute(w, d)
}

func (d *templateData) renderDir(w io.Writer, path string) error {
	list, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	d.Data = list

	return templates["list.html"].Execute(w, d)
}
