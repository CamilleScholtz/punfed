package punfed

import (
	"os"
	"strconv"
	"strings"

	"github.com/mholt/caddy"
)

type config struct {
	Scope string
	Dest  string

	Len int

	Keys []key
}

type key struct {
	User string
	Pass string
}

func parseConfig(c *caddy.Controller) (*config, error) {
	cfg := &config{}

	for c.Next() {
		switch c.Val() {
		case "scope":
			if !c.NextArg() {
				return cfg, c.ArgErr()
			}

			cfg.Scope = c.Val()
		case "destination":
			if !c.NextArg() {
				return cfg, c.ArgErr()
			}

			i, err := os.Stat(c.Val())
			if err != nil {
				return cfg, c.Err(err.Error())
			}
			if !i.IsDir() {
				return cfg, c.ArgErr()
			}

			cfg.Dest = c.Val()
		case "keys":
			if !c.NextArg() {
				return cfg, c.ArgErr()
			}

			for _, s := range strings.Split(c.Val(), ",") {
				k := strings.SplitN(s, ":", 2)
				cfg.Keys = append(cfg.Keys, key{k[0], k[1]})
			}
		case "filename_length":
			if !c.NextArg() {
				return cfg, c.ArgErr()
			}

			l, err := strconv.ParseUint(c.Val(), 10, 32)
			if err != nil {
				return cfg, c.Err(err.Error())
			}

			cfg.Len = int(l)
		}
	}

	return cfg, nil
}
