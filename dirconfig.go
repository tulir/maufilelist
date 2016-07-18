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
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// DirConfig is the config of a single directory
type DirConfig struct {
	Path          string            `json:"-"`
	DirectoryName string            `json:"directory-name"`
	FieldNames    []string          `json:"field-names"`
	EnableBackBtn bool              `json:"enable-back-button"`
	DirectoryList FieldInstructions `json:"directory-list"`
	FileList      FieldInstructions `json:"file-list"`
}

// FieldInstructions is a file or directory list config
type FieldInstructions struct {
	Enabled      bool             `json:"enabled"`
	ParsingRaw   []string         `json:"parsing"`
	Parsing      []*regexp.Regexp `json:"-"`
	FieldDataRaw []string         `json:"field-data"`
	FieldData    [][]FieldData    `json:"-"`
}

// FieldDataType is the type of a FieldData object
type FieldDataType int

// Allowed FieldDataTypes
const (
	TypeLiteral    FieldDataType = iota
	TypeName       FieldDataType = iota
	TypeLastChange FieldDataType = iota
	TypeArg        FieldDataType = iota
)

// FieldData contains the type and data of a list entry definition field
type FieldData struct {
	Type FieldDataType
	Data interface{}
}

// GetData gets the data for all the fields
func (instructions FieldInstructions) GetData(file os.FileInfo) []string {
	var args []string
	for _, parser := range instructions.Parsing {
		if !parser.MatchString(file.Name()) {
			continue
		}
		args = parser.FindStringSubmatch(file.Name())
	}

	var values = make([]string, len(instructions.FieldData))
	for i, fieldDataList := range instructions.FieldData {
		values[i] = instructions.getFieldData(file, args, fieldDataList)
	}
	return values
}

func (instructions FieldInstructions) getFieldData(file os.FileInfo, args []string, fieldDataList []FieldData) string {
	var buf bytes.Buffer
	for _, fd := range fieldDataList {
		switch fd.Type {
		case TypeName:
			buf.WriteString(file.Name())
		case TypeLastChange:
			format, ok := fd.Data.(string)
			if !ok || len(format) == 0 {
				format = "02.01.2006"
			}
			buf.WriteString(file.ModTime().Format(format))
		case TypeLiteral:
			text, _ := fd.Data.(string)
			buf.WriteString(text)
		case TypeArg:
			n, _ := fd.Data.(int)
			if len(args) > n && n >= 0 {
				buf.WriteString(args[n])
			}
		}
	}
	return buf.String()
}

// Parse does all necessary parsing of raw data
func (cfg *DirConfig) Parse() error {
	var err error

	cfg.DirectoryList.Parsing, err = cfg.parseParsingData(cfg.DirectoryList.ParsingRaw)
	if err != nil {
		return fmt.Errorf("Directory name %s", err)
	}

	cfg.FileList.Parsing, err = cfg.parseParsingData(cfg.FileList.ParsingRaw)
	if err != nil {
		return fmt.Errorf("File name %s", err)
	}

	cfg.DirectoryList.FieldData, err = cfg.parseFieldData(cfg.DirectoryList.FieldDataRaw)
	if err != nil {
		return fmt.Errorf("Parsing directory %s", err)
	}

	cfg.FileList.FieldData, err = cfg.parseFieldData(cfg.FileList.FieldDataRaw)
	if err != nil {
		return fmt.Errorf("Parsing file %s", err)
	}

	return nil
}

func (cfg *DirConfig) parseParsingData(data []string) ([]*regexp.Regexp, error) {
	parsed := make([]*regexp.Regexp, len(data))
	for i, parser := range data {
		var err error
		parsed[i], err = regexp.Compile(parser)
		if err != nil {
			return nil, fmt.Errorf("parser #%d in %s is invalid: %s", i+1, cfg.Path, err)
		}
	}
	return parsed, nil
}

func (cfg *DirConfig) parseFieldData(definitions []string) ([][]FieldData, error) {
	var data = make([][]FieldData, len(definitions))
	for defN, def := range definitions {
		i := 0
		dataList := make([]FieldData, 1)
	DefParser:
		for i < len(def)-1 {
			var fd FieldData
			var err error
			if def[i] == ' ' {
				i++
				continue DefParser
			} else if def[i] == '$' {
				fd, i, err = parseArg(i, def)
			} else if def[i] == '`' {
				fd, i, err = parseLiteral(i, def)
			} else {
				fd, i, err = parseParam(i, def)
			}
			if err != nil {
				return data, fmt.Errorf("field data entry #%d failed: %s", defN, err)
			}
			dataList = append(dataList, fd)
		}
		data[defN] = dataList
	}
	return data, nil
}

func parseArg(i int, def string) (FieldData, int, error) {
	argEnd := strings.IndexAny(def[i+1:], "`$ ")
	if argEnd == -1 {
		argEnd = len(def[i:])
	} else {
		argEnd++
	}

	argEnd += len(def[:i])

	argN, err := strconv.Atoi(def[i+1 : argEnd])
	if err != nil {
		return FieldData{}, argEnd + 1, fmt.Errorf("Invalid argument number %s at column %d", def[i+1:argEnd], i+1)
	}

	return FieldData{Type: TypeArg, Data: argN}, argEnd + 1, nil
}

func parseLiteral(i int, def string) (FieldData, int, error) {
	for {
		literalEnd := strings.IndexRune(def[i+1:], '`')
		if literalEnd == -1 {
			return FieldData{}, len(def), fmt.Errorf("Unterminated literal after start at column %d", i)
		}

		literalEnd += len(def[:i]) + 1
		if def[literalEnd-1] == '\\' {
			continue
		}
		return FieldData{Type: TypeLiteral, Data: def[i+1 : literalEnd]}, literalEnd + 1, nil
	}
}

func parseParam(i int, def string) (FieldData, int, error) {
	argEnd := strings.IndexAny(def[i:], "`$ ")
	if argEnd == -1 {
		argEnd = len(def)
	}

	argEnd += len(def[:i])

	param := def[i:argEnd]
	var arg string
	if strings.ContainsRune(param, ':') {
		parts := strings.Split(param, ":")
		param = parts[0]
		arg = strings.Join(parts[1:], ":")
	}
	var fd FieldData
	switch param {
	case "file-name":
		fd.Type = TypeName
	case "last-change":
		fd.Type = TypeLastChange
		if len(arg) > 0 {
			// TODO format in arg
		}
	default:
		return FieldData{}, argEnd + 1, fmt.Errorf("Unknown data key %s", param)
	}
	return fd, argEnd, nil
}
