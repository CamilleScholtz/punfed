package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// config is a stuct with all config values. See `runtime/config/config.toml`
// for more information about these values.
type config struct {
	// Root URL.
	Root string

	// URL & port to listen to.
	Listen string

	// The maximum filesize.
	MaxFileSize int64

	// Target directory on disk that serves as upload destination.
	WritePath string

	// Uploaded files can be gotten back from here.
	ServePath string

	// The lenght of generated random filenames.
	RandomFilenameLenght int

	// The accepted username & password combinations.
	AcceptedKeys []key
}

type key struct {
	User string
	Pass string
}

// parseConfig parses a toml config.
func parseConfig() (*config, error) {
	c := &config{}

	if _, err := toml.DecodeFile("/etc/caddy/upload.toml", c); err != nil {
		return nil, fmt.Errorf("config %s: %s", "/etc/caddy/upload.toml", err)
	}

	// Convert `MaxFileSize` to MB.
	c.MaxFileSize = c.MaxFileSize << 20

	return c, nil
}
