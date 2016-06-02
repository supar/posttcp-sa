package main

import (
	"io/ioutil"
	"os"
	"testing"
)

type dbCfgMock struct {
	Conf, Valid   string
	ValidInterval int
	ValidLimit    float64
}

var t_toml = []dbCfgMock{
	dbCfgMock{
		Conf: `
# Some comment
[database]
type = "mysql"
user = "any"
name = "dbname"
host = "some.host"

[score]
interval = 77
`,
		Valid:         "any@tcp(some.host:3306)/dbname",
		ValidInterval: 77,
		ValidLimit:    0.22,
	},
	dbCfgMock{
		Conf: `
# Some comment
[database]
type = "mysql"
name = "dbname"
user = "any"
password = "good"
port = 3336
host = "some.host"
  [database.parameters]
  charset = "utf8"
  loc = "US/Pacific"
  parseTime = "true"

[score]
  limit = 0.13
`,
		Valid:         "any:good@tcp(some.host:3336)/dbname?charset=utf8&loc=US%2FPacific&parseTime=true",
		ValidInterval: 60,
		ValidLimit:    0.13,
	},
}

// Helper: create temporary file
func getTempFile() (fpath string, err error) {
	var (
		file *os.File
	)

	if file, err = ioutil.TempFile("", "test_config.toml"); err != nil {
		return
	}

	fpath = file.Name()
	file.Close()

	return
}

// Helper: Write temporary content
func setTempConf(fpath, content string) (err error) {
	var file *os.File

	if file, err = os.OpenFile(fpath, os.O_WRONLY, 0666); err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)

	return
}

func TestConfig_FileNotexists(t *testing.T) {
	var (
		err error
	)

	if _, err = NewConfig("/this/is/not/valid/File/check"); err == nil {
		t.Fatal("Expected error on none valid file, but got nothing")
	}
}

func TestConfig_DirInsteadFile(t *testing.T) {
	var (
		dir string
		err error
	)

	if dir, err = ioutil.TempDir("", ""); err != nil {
		t.Fatalf("Can't get temporary dir, error: %s", err.Error())
	}

	if _, err = NewConfig(dir); err == nil {
		t.Fatal("Expected error message, but got nothing")
	}
}

func TestConfig_LoadFile(t *testing.T) {
	var (
		fpath string
		err   error
		cfg   *Config
	)

	if fpath, err = getTempFile(); err != nil {
		t.Fatalf("Cannot create temporary file: %s", err.Error())
	}
	defer os.Remove(fpath)

	for _, v := range t_toml {
		if err = setTempConf(fpath, v.Conf); err != nil {
			t.Fatalf("Cannot write to temporary file: %s", err.Error())
		}

		if cfg, err = NewConfig(fpath); err != nil {
			t.Fatalf("Expected to open file %s, but got error: %s", fpath, err.Error())
		}

		if err = cfg.Parse(); err != nil {
			t.Errorf("Unexpected error %s", err.Error())
		}

		if dsn := cfg.Database.DSN(); v.Valid != dsn {
			t.Errorf("Expected %s, but got %s", v.Valid, dsn)
		}

		os.Truncate(fpath, 0)
	}
}

func Test_GetEmptyScoreInterval(t *testing.T) {
	var (
		cfg          = &Config{}
		interval int = 60
	)

	if v := cfg.GetScoreInterval(); v != interval {
		t.Errorf("Expected score interval %d, but got %d", interval, v)
	}
}

func Test_GetEmptyScoreLinit(t *testing.T) {
	var (
		cfg           = &Config{}
		limit float64 = 0.22
	)

	if v := cfg.GetScoreLimit(); v != limit {
		t.Errorf("Expected score interval %f, but got %f", limit, v)
	}
}

func Test_GetScoreInterval(t *testing.T) {
	var (
		fpath string
		err   error
		cfg   *Config
	)

	if fpath, err = getTempFile(); err != nil {
		t.Fatalf("Cannot create temporary file: %s", err.Error())
	}
	defer os.Remove(fpath)

	for _, v := range t_toml {
		if err = setTempConf(fpath, v.Conf); err != nil {
			t.Fatalf("Cannot write to temporary file: %s", err.Error())
		}

		if cfg, err = NewConfig(fpath); err != nil {
			t.Fatalf("Expected to open file %s, but got error: %s", fpath, err.Error())
		}

		if err = cfg.Parse(); err != nil {
			t.Errorf("Unexpected error %s", err.Error())
		}

		if interval := cfg.GetScoreInterval(); v.ValidInterval != interval {
			t.Errorf("Expected score interval %d, but got %d", v.ValidInterval, interval)
		}

		os.Truncate(fpath, 0)
	}
}

func Test_GetScoreLimit(t *testing.T) {
	var (
		fpath string
		err   error
		cfg   *Config
	)

	if fpath, err = getTempFile(); err != nil {
		t.Fatalf("Cannot create temporary file: %s", err.Error())
	}
	defer os.Remove(fpath)

	for _, v := range t_toml {
		if err = setTempConf(fpath, v.Conf); err != nil {
			t.Fatalf("Cannot write to temporary file: %s", err.Error())
		}

		if cfg, err = NewConfig(fpath); err != nil {
			t.Fatalf("Expected to open file %s, but got error: %s", fpath, err.Error())
		}

		if err = cfg.Parse(); err != nil {
			t.Errorf("Unexpected error %s", err.Error())
		}

		if limit := cfg.GetScoreLimit(); v.ValidLimit != limit {
			t.Errorf("Expected score interval %f, but got %f", v.ValidLimit, limit)
		}

		os.Truncate(fpath, 0)
	}
}
