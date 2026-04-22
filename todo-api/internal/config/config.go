package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	CookieSecure bool
}

func Load() Config {
	c := Config{
		Port:         "9090",
		CookieSecure: false,
	}
	if v := os.Getenv("PORT"); v != "" {
		c.Port = v
	}
	if v := os.Getenv("COOKIE_SECURE"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			c.CookieSecure = b
		}
	}
	return c
}
