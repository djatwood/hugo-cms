package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v2"
)

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

func getFileNames(dirPath string) ([]ListEntry, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		// Directories first
		isDir := files[i].IsDir()
		if isDir != files[j].IsDir() {
			return isDir
		}

		// Then sort by last mod time
		if files[i].ModTime() != files[j].ModTime() {
			return files[i].ModTime().After(files[j].ModTime())
		}

		// Then sort by name
		aName := files[i].Name()
		bName := files[j].Name()
		for i := 0; i < minInt(len(aName), len(bName)); i++ {
			if aName[i] == bName[i] {
				continue
			}
			return aName[i] < bName[i]
		}

		return len(aName) <= len(bName)
	})

	list := make([]ListEntry, len(files))
	for i := range files {
		fileName := files[i].Name()

		name := fileName
		if files[i].IsDir() {
			name += "/"
		} else if path.Ext(fileName) == ".md" {
			file, err := os.ReadFile(dirPath + "/" + fileName)
			if err != nil {
				return nil, err
			}
			site := new(ConfigFile)
			err = yaml.Unmarshal(file, site)
			if err != nil {
				return nil, err
			}
			if len(site.Title) > 0 {
				name = site.Title
			}
		}

		list[i] = ListEntry{
			Name: name,
			Path: fileName,
		}
	}

	return list, nil
}
