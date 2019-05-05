package punfed

import (
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/mholt/caddy"
)

type config struct {
	Key   string
	Scope string
	Save  string
	Serve string
	Len   int
	Keys  []key
}

type key struct {
	User string
	Pass string
}

func parseConfig(c *caddy.Controller) (*config, error) {
	cfg := &config{
		Key: c.Key,
		Len: 4,
	}

	for c.Next() {
		switch c.Val() {
		case "scope":
			if !c.NextArg() {
				return cfg, c.ArgErr()
			}

			cfg.Scope = c.Val()
		case "save":
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

			cfg.Save = c.Val()
		case "serve":
			if !c.NextArg() {
				return cfg, c.ArgErr()
			}

			cfg.Serve = c.Val()
		case "keys":
			if !c.NextArg() {
				return cfg, c.ArgErr()
			}

			for _, s := range strings.Split(c.Val(), " ") {
				k := strings.SplitN(s, ":", 2)
				cfg.Keys = append(cfg.Keys, key{k[0], k[1]})

				if err := os.MkdirAll(path.Join(cfg.Save, k[0]), os.
					ModePerm); err != nil {
					return cfg, c.ArgErr()
				}
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
