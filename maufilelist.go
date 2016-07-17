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
	"fmt"
	"io/ioutil"
	flag "maunium.net/go/mauflag"
	log "maunium.net/go/maulogger"
	"net/http"
)

// Config is the main configuration
type Config struct {
	IP             string          `json:"ip"`
	Port           int             `json:"port"`
	DefaultFormat  string          `json:"default-format"`
	MainDomain     VDom            `json:"main-domain"`
	VirtualDomains map[string]VDom `json:"virtual-domains"`
}

// VDom is a virtual domain
type VDom struct {
	Root     string            `json:"root"`
	Subroots map[string]string `json:"subroots"`
}

var config Config

var confPath = flag.Make().Default("/etc/mfl/config.json").ShortKey("c").LongKey("config").String()
var logPath = flag.Make().Default("/var/log/mfl").ShortKey("l").LongKey("logs").String()
var debug = flag.Make().Default("false").ShortKey("d").LongKey("debug").Bool()

func main() {
	flag.Parse()

	log.Init()
	log.Fileformat = func(now string, i int) string { return fmt.Sprintf("%[3]s/%[1]s-%02[2]d.log", now, i, *logPath) }
	if *debug {
		log.PrintLevel = 0
	}
	log.Debugln("Logger initialized.")

	log.Debugln("Loading config...")
	data, err := ioutil.ReadFile(*confPath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}
	log.Debugln("Config loaded!")

	http.HandleFunc("/", handle)
	log.Infof("Listening on %s:%d", config.IP, config.Port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.IP, config.Port), nil)
}
