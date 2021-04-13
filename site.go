package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
)

func listSites(c echo.Context) error {
	dir, err := os.Open("sites")
	if err != nil {
		return err
	}
	defer dir.Close()

	paths, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	sites := make([]ListEntry, len(paths))
	for i, p := range paths {
		f, err := os.ReadFile(fmt.Sprintf("sites/%s/config.toml", p))
		if err != nil {
			return err
		}

		config := new(ConfigFile)
		err = toml.Unmarshal(f, config)
		if err != nil {
			return err
		}

		sites[i] = ListEntry{
			Name: config.Title,
			Path: p,
		}
	}

	return c.JSON(http.StatusOK, sites)
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

	s := new(site)
	err = yaml.Unmarshal(file, s)
	if err != nil {
		return nil, err
	}

	return s, nil
}
