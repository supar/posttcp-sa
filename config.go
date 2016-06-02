package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/supar/dsncfg"
	"os"
)

var (
	NAME = "postmy-rbl"

	CONFIGFILE,
	VERSION,
	BUILDDATE string

	PrintVersion bool

	CONSOLELOG = LevelDebug
)

// Programm configuration object
type Config struct {
	ConfFile string `tomp:"-"`
	Score    *Score
	Server   string
	Log      map[string]LogAdapter
	Database *dsncfg.Database `toml:"database"`
}

type LogAdapter struct {
	File  string `json:"filename"`
	Level int    `json:"level"`
}

type Score struct {
	Interval int
	Limit    float64
}

//
func init() {
	if NAME == "" {
		NAME = "posttcp-sa"
	}

	flag.StringVar(&CONFIGFILE, "C", "./"+NAME+".toml", "Configuration file path required")
	flag.IntVar(&CONSOLELOG, "v", 0, "Console verbose level output, default 0 - off, 7 - debug")
	flag.BoolVar(&PrintVersion, "V", false, "Print version")
}

func NewConfig(filePath string) (conf *Config, err error) {
	var (
		file os.FileInfo
	)

	if file, err = os.Stat(filePath); os.IsNotExist(err) || file.IsDir() {
		return nil, fmt.Errorf("Can't open file: %s", filePath)
	}

	conf = &Config{
		ConfFile: filePath,
		Database: &dsncfg.Database{},
	}

	return
}

func (this *Config) Parse() (err error) {
	if _, err = toml.DecodeFile(this.ConfFile, this); err != nil {
		return
	}

	if err = this.Database.Init(); err != nil {
		return
	}

	if this.Log == nil {
		this.Log = make(map[string]LogAdapter)
	}

	if _, ok := this.Log["console"]; !ok || CONSOLELOG > 0 {
		this.Log["console"] = LogAdapter{Level: CONSOLELOG}
	}

	return nil
}

func (this *Config) GetLogAdapterJSON(name string) (cfg string, err error) {
	var cfg_tmp []byte

	if val, ok := this.Log[name]; !ok {
		err = fmt.Errorf("Unknown log adapter `%s`", name)
	} else {
		cfg_tmp, err = json.Marshal(val)
		cfg = string(cfg_tmp)
	}

	return
}

func (this *Config) GetScoreInterval() int {
	if this.Score == nil || this.Score.Interval == 0 {
		return 60
	}

	return this.Score.Interval
}

func (this *Config) GetScoreLimit() float64 {
	if this.Score == nil || this.Score.Limit == 0 {
		return 0.22
	}

	return this.Score.Limit
}
