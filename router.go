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
	"github.com/labstack/echo/v4/middleware"
)

func server() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.CORS())

	e.GET("/", listSites)
	e.GET("/:site", getSite)
	e.GET("/:site/:section", getSection)
	e.GET("/:site/:section/*", getFile)

	return e
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

	prefix := fmt.Sprintf("sites/%s/%s/", name, section.Path)
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

	path := fmt.Sprintf("sites/%s/%s/%s", name, section.Path, c.Param("*"))
	stats, err := os.Stat(path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if stats.IsDir() {
		files, err := getFileNames(path)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"kind": "dir", "data": files})
	}

	file, err := os.Open(path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"kind": "file", "data": string(data)})
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
