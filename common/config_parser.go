package common

import (
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// ConfigParser is a wrapper for Viper
type ConfigParser struct {
	v *viper.Viper
	l *logrus.Logger
}

// NewConfigParser for a new parser with a given environemtn variable prefix
func NewConfigParser(l *logrus.Logger, prefix string) *ConfigParser {
	v := viper.New()
	v.SetConfigType("toml")

	v.SetEnvPrefix(prefix)
	v.AutomaticEnv()

	return &ConfigParser{
		v: v,
		l: l,
	}
}

// ReadFile reads a config file from a given path
func (p *ConfigParser) ReadFile(path string) error {
	p.v.SetConfigFile(path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		p.l.Info("config file not found, using defaults")
		return nil
	}

	err := p.v.ReadInConfig()
	if err != nil {
		return errors.Wrap(err, "failed to read configuration file")
	}

	return nil
}

// GetString returns a string from our config parser
func (p *ConfigParser) GetString(key string, fallback string) string {
	p.v.SetDefault(key, fallback)
	return p.v.GetString(key)
}

// GetInt64 returns an int64 from our config parser
func (p *ConfigParser) GetInt64(key string, fallback int64) int64 {
	p.v.SetDefault(key, fallback)
	return p.v.GetInt64(key)
}
