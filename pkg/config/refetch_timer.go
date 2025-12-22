package config

import "time"

type RefetchTimer struct {
	Enabled  bool          `yaml:"enabled" default:"true"`
	Interval time.Duration `yaml:"interval" default:"24h"`
}
