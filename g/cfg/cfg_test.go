package cfg

import (
	"os"
	"io/ioutil"
	"testing"
)

// avoids polluting test messing with project own configuration
// testdata is ignored by go build
func init() {
	err := os.Chdir("testdata")
	if err != nil {
		panic(err)
	}
}

func TestPaths(t *testing.T) {

	err := os.RemoveAll("config")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	_ = os.Unsetenv("APP_ENV")

	// #1. test error case: no config/ dir
	paths, err := Paths()
	if paths != nil {
		t.Log("paths must be nil, there are no config/ dir")
		t.Fail()
	}
	if err == nil {
		t.Log("err can't be nil, there are no config/ dir")
		t.Fail()
	}

	// #2. test config.default w/o APP_ENV
	err = os.Mkdir("config", 0755)
	if err != nil {
		t.Logf("error when creating config/: %s", err)
		t.Fail()
	}
	err = ioutil.WriteFile("config/config.default", []byte{}, 0644)
	if err != nil {
		t.Logf("error when creating config/config.default: %s", err)
		t.Fail()
	}

	paths, err = Paths()
	if paths == nil {
		t.Log("paths must NOT be nil, there is config/config.default")
		t.Fail()
	}
	if len(paths) != 1 && paths[0] != "config/config.default" {
		t.Log("paths must be {`config/config.default`}")
		t.Fail()
	}
	if err != nil {
		t.Log("err must be nil, there is config/config.default")
		t.Fail()
	}

	// #3. wrong APP_ENV location
	err = os.Setenv("APP_ENV", "aws")
	defer os.Unsetenv("APP_ENV")
	if err != nil {
		t.Logf("Can't set env variable APP_ENV: %s", err)
		t.Fail()
	}
	paths, err = Paths()
	if paths != nil {
		t.Log("paths must be nil, APP_ENV points to wrong file")
		t.Fail()
	}
	if err == nil {
		t.Log("paths must not be nil, APP_ENV points to wrong file")
		t.Fail()
	}

	// #4. Created APP_ENV file
	err = ioutil.WriteFile("config/config.aws", []byte{}, 0644)
	if err != nil {
		t.Logf("error when creating config/config.aws: %s", err)
		t.Fail()
	}
	paths, err = Paths()
	if paths == nil {
		t.Log("paths must not be nil, APP_ENV points to right file")
		t.Fail()
	}
	if len(paths) != 2 && paths[0] != "config/config.default" &&
		paths[1] != "config/config.aws" {
		t.Log("paths must be {`config/config.default`, `config/config.aws`}")
		t.Fail()
	}
	if err != nil {
		t.Log("err must be nil, there is config/config.default")
		t.Fail()
	}

	err = os.RemoveAll("config")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

}

type TestConf struct {
	Verbose bool	`yaml:"verbose"`
	Jobs int	`yaml:"jobs"`
	Endpoint string `yaml:"endpoint"`
}

func TestLoad(t *testing.T) {
	err := os.RemoveAll("config")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	err = os.Mkdir("config", 0755)
	if err != nil {
		t.Logf("error when creating config/: %s", err)
		t.Fail()
	}

	// single file
	err = ioutil.WriteFile("config/config.default", []byte("verbose: true\njobs: 42\nendpoint: ${ENDPOINT}\n"), 0644)
	if err != nil {
		t.Logf("error when creating config/config.default: %s", err)
		t.Fail()
	}

	var conf TestConf
	err = Load(&conf)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if conf.Jobs != 42 {
		t.Log("conf.Jobs MUST be 42")
		t.Fail()
	}
	if conf.Endpoint != "" {
		t.Log("conf.Endpoint MUST be empty, ENDPOINT variable is not defined")
		t.Fail()
	}
	// TODO: assert the rest

	// second file
	err = ioutil.WriteFile("config/config.unittest", []byte("jobs: 11\n"), 0644)
	if err != nil {
		t.Logf("error when creating config/config.unittest: %s", err)
		t.Fail()
	}

	// #2 NO APP_ENV + ENDPOINT
	os.Setenv("ENDPOINT", "end://point")
	defer os.Unsetenv("APP_ENV")
	err = Load(&conf)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if conf.Jobs != 42 {
		t.Log("conf.Jobs MUST be 42, APP_ENV is empty")
		t.Fail()
	}
	if conf.Endpoint != "end://point" {
		t.Log("conf.Endpoint MUST be 'end://point', according ENDPOINT variable")
		t.Fail()
	}

	// #3 WITH APP_ENV
	os.Setenv("APP_ENV", "unittest")
	defer os.Unsetenv("APP_ENV")
	err = Load(&conf)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if conf.Jobs != 11 {
		t.Log("conf.Jobs MUST be 11, APP_ENV is `unittest`")
		t.Fail()
	}

	err = os.RemoveAll("config")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}

type PostgresConf struct {
	Endpoint string
	Debug bool
}

type TestConfNested struct {
	Verbose bool	`yaml:"verbose"`
	Jobs int	`yaml:"jobs"`
	Postgres	PostgresConf
}

func TestLoadNested(t *testing.T) {

	err := os.RemoveAll("config")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	err = os.Mkdir("config", 0755)
	if err != nil {
		t.Logf("error when creating config/: %s", err)
		t.Fail()
	}

	// single file, nested
	b := []byte("verbose: false\njobs: 1\npostgres:\n  endpoint: ${ENDPOINT}\n  debug: true\n")
	err = ioutil.WriteFile("config/config.default", b, 0644)
	if err != nil {
		t.Logf("error when creating config/config.default: %s", err)
		t.Fail()
	}

	os.Setenv("ENDPOINT", "://point")
	var conf TestConfNested
	err = Load(&conf)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if conf.Postgres.Endpoint != "://point" {
		t.Log("postgres:endpoint MUST BE ://point")
		t.Fail()
	}

	err = os.RemoveAll("config")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
}
