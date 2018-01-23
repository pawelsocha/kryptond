package config

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

// Config represent project configuration
var Cfg *Config

type Config struct {
	Database struct {
		Type     string `ini:"type"`
		Name     string `ini:"database"`
		Host     string `ini:"host"`
		User     string `ini:"user"`
		Password string `ini:"password"`
	}

	Mikrotik struct {
		Username string `ini:"username"`
		Password string `ini:"password"`
		Insecure bool   `ini:"insecure"`
		Workers  uint   `ini:"workers"`
	}
}

// New create config struct with data from ini file
func New(filename string) (*Config, error) {
	c := &Config{}
	if err := c.ReadConfig(filename); err != nil {
		return nil, err
	}

	Cfg = c
	return c, nil
}

// ReadConfig read configuration from INI file
func (c *Config) ReadConfig(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return err
	}

	cfg, err := ini.InsensitiveLoad(filename)
	if err != nil {
		return err
	}

	err = cfg.Section("database").MapTo(&c.Database)
	if err != nil {
		return err
	}

	if err = cfg.Section("mikrotik").MapTo(&c.Mikrotik); err != nil {
		return err
	}
	return nil
}

// GetDatabaseDSN prepare database connection uri using data from INI file.
func (c Config) GetDatabaseDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=10s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		3306,
		c.Database.Name,
	)
}
