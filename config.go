package punfed

import (
	"os"
	"strconv"
	"strings"

	"github.com/mholt/caddy"
)

type config struct {
	Scope string
	Save  string
	Serve string

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
		if !c.NextArg() {
			return cfg, c.ArgErr()
		}

		switch c.Val() {
		case "scope":
			cfg.Scope = c.Val()
		case "save":
			i, err := os.Stat(c.Val())
			if err != nil {
				return cfg, c.Err(err.Error())
			}
			if !i.IsDir() {
				return cfg, c.ArgErr()
			}

			cfg.Save = c.Val()
		case "serve":
			cfg.Serve = c.Val()
		case "keys":
			for _, s := range strings.Split(c.Val(), ",") {
				k := strings.SplitN(s, ":", 2)
				cfg.Keys = append(cfg.Keys, key{k[0], k[1]})
			}
		case "filename_length":
			l, err := strconv.ParseUint(c.Val(), 10, 32)
			if err != nil {
				return cfg, c.Err(err.Error())
			}

			cfg.Len = int(l)
		}
	}

	return cfg, nil
}
