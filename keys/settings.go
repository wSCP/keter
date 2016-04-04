package keys

import (
	"fmt"
	"os"
	"time"
)

type Settings struct {
	Env         string
	Path        string
	File        string
	ChainExpiry time.Duration
	Verbose     bool
}

func NewSettings(env, path, file string, expiry time.Duration, verbose bool) *Settings {
	return &Settings{env, path, file, expiry, verbose}
}

func DefaultSettings(k *Keys) error {
	if k.Settings == nil {
		k.Settings = &Settings{
			"XDG_CONFIG_HOME",
			"keter/keterrc",
			"",
			2 * time.Second,
			false,
		}
	}
	return nil
}

func SetSettings(env, path, file string, expiry time.Duration, verbose bool) Config {
	return config{
		50,
		func(k *Keys) error {
			k.Settings = NewSettings(env, path, file, expiry, verbose)
			return nil
		},
	}
}

func (s *Settings) LoadPath() string {
	if s.File != "" {
		return s.File
	}
	var pth string
	configHome := os.Getenv(s.Env)
	if configHome != "" {
		pth = fmt.Sprintf("%s/%s", configHome, s.Path)
	} else {
		pth = fmt.Sprintf("%s/%s/%s", os.Getenv("HOME"), ".config", s.Path)
	}
	return pth
}
