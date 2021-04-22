package main

import (
	"errors"
	"fmt"
	"io"
	"mime"
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

	prefix := fmt.Sprintf("sites/%s/%s", name, section.Path)
	paths, err := doublestar.Glob(prefix + "/" + section.Match + section.Extension)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	files := make([]os.FileInfo, 0, len(paths))
	existing := make(map[string]bool)
	for _, path := range paths {
		name := strings.TrimPrefix(path, prefix)
		key := strings.Split(name, string(os.PathSeparator))[1]
		if _, ok := existing[name]; ok {
			continue
		}
		existing[key] = true

		p, err := os.Stat(prefix + "/" + key)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		files = append(files, p)
	}

	sort.Slice(files, func(i, j int) bool {
		return fileLess(files[i], files[j])
	})

	list, err := renameFileList(prefix, files)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"label": section.Label,
		"files": list,
	})
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

	p := fmt.Sprintf("sites/%s/%s/%s", name, section.Path, c.Param("*"))
	stats, err := os.Stat(p)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	if stats.IsDir() {
		files, err := getFileNames(p)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusOK, map[string]interface{}{"kind": "dir", "data": files})
	}

	file, err := os.Open(p)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	ext := path.Ext(p)
	if ext != ".md" {
		return c.Blob(http.StatusOK, mime.TypeByExtension(ext), data)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{"kind": ext, "data": string(data)})
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
		return fileLess(files[i], files[j])
	})

	return renameFileList(dirPath, files)
}

func fileLess(a, b os.FileInfo) bool {
	// Directories first
	isDir := a.IsDir()
	if isDir != b.IsDir() {
		return isDir
	}

	// Then sort by last mod time
	if a.ModTime() != b.ModTime() {
		return a.ModTime().After(b.ModTime())
	}

	// Then sort by name
	aName := a.Name()
	bName := b.Name()
	for i := 0; i < minInt(len(aName), len(bName)); i++ {
		if aName[i] == bName[i] {
			continue
		}
		return aName[i] < bName[i]
	}

	return len(aName) <= len(bName)
}

func renameFileList(dirPath string, files []os.FileInfo) ([]ListEntry, error) {
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
