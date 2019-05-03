package punfed

import (
	"os"
	"strconv"

	"github.com/mholt/caddy"
)

type config struct {
	Scope string
	Dest  string

	Len int

	//Keys []key
}

func parseConfig(c *caddy.Controller) (*config, error) {
	cfg := &config{}

	for c.Next() {
		for c.NextBlock() {
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
	}

	return cfg, nil
}
