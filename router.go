package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v2"
)

func server() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.GET("/", listSites)
	e.GET("/:site", getSite)
	e.GET("/:site/:section", getSection)
	e.GET("/:site/:section/*", getFile)

	return e
}

func listSites(c echo.Context) error {
	dir, err := os.Open("sites")
	if err != nil {
		return err
	}

	sites, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, sites)
}

func parseSite(name string) (*site, error) {
	dir := "sites/" + name
	_, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	file, err := os.ReadFile(dir + "/.cms/config.yaml")
	if err != nil {
		return nil, err
	}

	s := site{Dir: dir}

	err = yaml.Unmarshal(file, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func getSite(c echo.Context) error {
	name := c.Param("site")
	s, err := parseSite(name)
	if errors.Is(err, os.ErrNotExist) {
		return c.JSON(http.StatusNotFound, name+" not found")
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, s)
}

func getSection(c echo.Context) error {
	name := c.Param("site")
	s, err := parseSite(name)
	if errors.Is(err, os.ErrNotExist) {
		return c.JSON(http.StatusNotFound, name+" not found")
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	section, ok := s.Sections[c.Param("section")]
	if !ok {
		return c.JSON(http.StatusNotFound, "section not found")
	}

	prefix := fmt.Sprintf("%s/%s/", s.Dir, section.Path)
	files, err := doublestar.Glob(prefix + section.Match + section.Extension)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	existing := make(map[string]bool)
	for i := 0; i < len(files); i++ {
		name := strings.TrimPrefix(files[i], prefix)
		trimmedName := strings.Split(name, string(os.PathSeparator))[0]

		if name != trimmedName {
			trimmedName += "/"
		}

		if _, ok := existing[trimmedName]; !ok {
			files[i] = trimmedName
			existing[trimmedName] = true
		} else {
			copy(files[i:], files[i+1:])
			files = files[:len(files)-1]
			i--
		}
	}

	return c.JSON(http.StatusOK, files)
}

func getFile(c echo.Context) error {
	name := c.Param("site")
	s, err := parseSite(name)
	if errors.Is(err, os.ErrNotExist) {
		return c.JSON(http.StatusNotFound, name+" not found")
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	section, ok := s.Sections[c.Param("section")]
	if !ok {
		return c.JSON(http.StatusNotFound, "section not found")
	}

	path := fmt.Sprintf("%s/%s/%s", s.Dir, section.Path, c.Param("*"))
	stats, err := os.Stat(path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if stats.IsDir() {
		files, err := getFileNames(path)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, files)
	}

	file, err := os.Open(path)

	data, err := io.ReadAll(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.String(http.StatusOK, string(data))
}

func getFileNames(path string) ([]string, error) {
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		isDir := files[i].IsDir()
		if isDir != files[j].IsDir() {
			return isDir
		}

		aName := files[i].Name()
		bName := files[j].Name()

		for i := 0; i < min(len(aName), len(bName)); i++ {
			if aName[i] < bName[i] {
				return true
			}
			if aName[i] > bName[i] {
				return false
			}
		}

		return len(aName) <= len(bName)
	})

	names := make([]string, len(files))
	for i := range files {
		names[i] = files[i].Name()
		if files[i].IsDir() {
			names[i] += "/"
		}
	}

	return names, nil
}

func min(a, b int) int {
	if b < a {
		return b
	}
	return a
}
