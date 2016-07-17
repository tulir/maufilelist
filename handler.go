// mauFileList - A configurable file listing system
// Copyright (C) 2016 Tulir Asokan

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	log "maunium.net/go/maulogger"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// DirConfig is the config of a single directory
type DirConfig struct {
	DirectoryName string     `json:"directory-name"`
	FieldNames    []string   `json:"field-names"`
	FieldData     []string   `json:"field-data"`
	EnableBackBtn bool       `json:"enable-back-button"`
	DirectoryList ObjectList `json:"directory-list"`
	FileList      ObjectList `json:"file-list"`
}

// ObjectList is a file or directory list config
type ObjectList struct {
	Enable    bool     `json:"enable"`
	Parsing   []string `json:"parsing"`
	FieldData []string `json:"field-data"`
}

var cachedFormats = make(map[string]*template.Template)
var cachedConfigs = make(map[string]*DirConfig)

func handle(w http.ResponseWriter, r *http.Request) {
	var vd = config.MainDomain
	host := r.Header.Get("host")
	if len(host) > 0 {
		host = strings.ToLower(host)
		vdom, ok := config.VirtualDomains[host]
		if ok {
			vd = vdom
		}
	}

	var root = vd.Root
	var path = r.URL.RequestURI()

	for req, subroot := range vd.Subroots {
		if strings.HasPrefix(path, req) {
			root = subroot
			path = path[len(req):]
			break
		}
	}

	format, err := loadFormat(root, path)
	if err != http.StatusOK {
		w.WriteHeader(err)
		return
	}

	cfg, err := loadConfig(root, path)
	if err != http.StatusOK {
		w.WriteHeader(err)
		return
	}

	listFiles(w, r, cfg, format, filepath.Join(root, path))
}

func listFiles(w http.ResponseWriter, r *http.Request, cfg *DirConfig, format *template.Template, dir string) {

}

func loadFormat(root, path string) (*template.Template, int) {
	var formatPath = findFile(root, path, ".mfl-format.gohtml")
	if len(formatPath) == 0 {
		log.Warnln("No format file found for", filepath.Join(root, path))
		return nil, http.StatusNotFound
	}

	format, ok := cachedFormats[formatPath]
	if !ok {
		var err error
		format, err = template.New(formatPath).ParseFiles(formatPath)
		if err != nil {
			log.Errorln("Failed to load template at", formatPath+":", err)
			return nil, http.StatusInternalServerError
		}
		cachedFormats[formatPath] = format
	}
	return format, http.StatusOK
}

func loadConfig(root, path string) (*DirConfig, int) {
	var configPath = findFile(root, path, ".mfl.json")
	if len(configPath) == 0 {
		return nil, http.StatusNotFound
	}

	cfg, ok := cachedConfigs[configPath]
	if !ok {
		data, err := ioutil.ReadFile(configPath)
		if err != nil {
			log.Errorln("Failed to read config at", configPath+":", err)
			return nil, http.StatusInternalServerError
		}

		cfg = &DirConfig{}
		err = json.Unmarshal(data, &config)
		if err != nil {
			log.Errorln("Failed to unmarshal config at", configPath+":", err)
			return nil, http.StatusInternalServerError
		}
	}
	return cfg, http.StatusOK
}

func findFile(root, path, fileName string) string {
	var dir = filepath.Join(root, path)
	path = filepath.Join(dir, ".mfl.gohtml")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	for len(dir) <= len(root) {
		if strings.IndexRune(dir, '/') < 0 {
			return ""
		}
		dir = dir[:strings.LastIndex(dir, "/")]
		var path = filepath.Join(path, fileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
