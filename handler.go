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
	"io/ioutil"
	log "maunium.net/go/maulogger"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateData is the data given to templates
type TemplateData struct {
	Directory  string
	FieldNames []string
	Files      [][]string
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

	var dir = filepath.Join(root, path)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else if os.IsPermission(err) {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	format, errCode := loadFormat(root, path)
	if errCode != http.StatusOK {
		w.WriteHeader(errCode)
		return
	}

	cfg, errCode := loadConfig(root, path)
	if errCode != http.StatusOK {
		w.WriteHeader(errCode)
		return
	}

	listFiles(w, r, cfg, format, dir, files)
}

func listFiles(w http.ResponseWriter, r *http.Request, cfg *DirConfig, format *template.Template, dir string, files []os.FileInfo) {
	var templCfg = TemplateData{
		Directory:  cfg.DirectoryName,
		FieldNames: cfg.FieldNames,
		Files:      make([][]string, len(files)),
	}

	listFilesByFieldInstructions(cfg.DirectoryList, true, &templCfg, files)
	listFilesByFieldInstructions(cfg.FileList, false, &templCfg, files)

	err := format.Execute(w, templCfg)
	if err != nil {
		panic(err)
	}
}

func listFilesByFieldInstructions(list FieldInstructions, dir bool, templCfg *TemplateData, files []os.FileInfo) {
	if list.Enabled {
		for i, file := range files {
			if strings.HasPrefix(file.Name(), ".") || (file.IsDir() && dir) || (!file.IsDir() && !dir) {
				continue
			}

			fileData := list.GetData(file)
			if fileData == nil {
				i--
			} else {
				templCfg.Files[i] = fileData
			}
		}
	}
}

func loadFormat(root, path string) (*template.Template, int) {
	var formatPath = findFile(root, path, ".mfl-format.gohtml")
	if len(formatPath) == 0 {
		log.Warnln("No format file found for", filepath.Join(root, path))
		return nil, http.StatusNotFound
	}

	format, ok := cachedFormats[formatPath]
	if !ok {
		format = template.New(formatPath[len(root) : strings.LastIndex(formatPath, "/")+1])

		data, err := ioutil.ReadFile(formatPath)
		if err != nil {
			log.Errorln("Failed to read template at", formatPath+":", err)
			return nil, http.StatusInternalServerError
		}

		_, err = format.Parse(string(data))
		if err != nil {
			log.Errorln("Failed to parse template at", formatPath+":", err)
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
		err = json.Unmarshal(data, cfg)
		if err != nil {
			log.Errorln("Failed to unmarshal config at", configPath+":", err)
			return nil, http.StatusInternalServerError
		}

		err = cfg.Parse()
		if err != nil {
			log.Errorln("Failed to parse raw data in config at", configPath+":", err)
			return nil, http.StatusInternalServerError
		}

		cachedConfigs[configPath] = cfg
	}
	return cfg, http.StatusOK
}

func findFile(root, path, fileName string) string {
	var dir = filepath.Join(root, path)
	path = filepath.Join(dir, fileName)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	for len(dir) > len(root) {
		if strings.IndexRune(dir[:len(dir)-1], '/') < 0 {
			return ""
		}
		dir = dir[:strings.LastIndex(dir[:len(dir)-1], "/")]
		var path = filepath.Join(dir, fileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
