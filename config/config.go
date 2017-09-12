// Package config config opts and config file
package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/jessevdk/go-flags"
	"gopkg.in/yaml.v2"
)

var Opts struct {
	Conf     string `long:"conf" default:"config.yml" description:"stuns config file"`
	LogLevel string `long:"log_level" default:"info" description:"log level"`
	LogFile  string `long:"log_file" default:"stdout" description:"log file"`
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func init() {
	parser := flags.NewParser(&Opts, flags.HelpFlag|flags.PassDoubleDash|flags.IgnoreUnknown)

	_, err := parser.Parse()
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(-1)
	}
}

type Server struct {
	Port     int    `yaml:"port"`
	Address  string `yaml:"bind"`
	Protocol string `yaml:"proto"`
	Pass     string `yaml:"pass"`
}

type Upstream struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type,omitempty"`
	Targets []string `yaml:"targets"`
	Hash    string   `yaml:"hash"`
}

type Settings struct {
	Upstreams []Upstream `yaml:"upstreams"`
	Servers   []Server   `yaml:"servers"`
}

func Load(filename string) (*Settings, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	cfg := &Settings{}
	err = yaml.Unmarshal(content, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (m *Upstream) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Upstream

	// These are the default values for a basic metric config
	rawUpstream := plain{
		Type: "static",
	}
	if err := unmarshal(&rawUpstream); err != nil {
		return err
	}

	//TODO: Check for valid types

	*m = Upstream(rawUpstream)
	return nil
}
