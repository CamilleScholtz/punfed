package punfed

import (
	"os"
	"strconv"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

// init configures the plugin on the Caddy webserver.
func init() {
	caddy.RegisterPlugin("punfed", caddy.Plugin{
		ServerType: "http",
		Action:     Setup,
	})
}

// Setup creates a new middleware with the given configuration.
func Setup(c *caddy.Controller) error {
	config, err := parseCaddyConfig(c)
	if err != nil {
		return err
	}

	site := httpserver.GetConfig(c)
	if site.TLS == nil || !site.TLS.Enabled {
		if c.Dispenser.File() == "Testfile" {
			goto pass
		}
		for _, host := range []string{"127.0.0.1", "localhost", "[::1]", "::1"} {
			if site.Addr.Host == host || strings.HasPrefix(site.Addr.Host, host) {
				goto pass
			}
		}

		for _, scopeConf := range config.Scope {
			if !scopeConf.AcknowledgedNoTLS {
				return c.Err("You are using plugin 'punfed' on a site without TLS.")
			}
		}
	}

pass:
	site.AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return &Handler{
			Next:   next,
			Config: *config,
		}
	})

	return nil
}

// ScopeConfiguration represents the settings of a scope (URL path).
type ScopeConfiguration struct {
	// Target directory on disk that serves as upload destination.
	WriteToPath string

	// Key contains the key value uploaders need to provide.
	Key string

	// UploadProgressCallback is called every so often
	// to report the total bytes written to a single file and the current error,
	// including 'io.EOF'.
	UploadProgressCallback func(uint64, error)

	// If false, this plugin returns HTTP Errors.
	// If true, passes the given request to the next middleware
	// which could respond with an Error of its own, poorly obscuring where this plugin is used.
	SilenceAuthErrors bool

	// The user must set a "flag of shame" for sites that don't use TLS with 'punfed'. (read-only)
	// This keeps track of whether said flags has been set.
	AcknowledgedNoTLS bool

	// Append '_' and a randomized suffix of that length.
	FilenameLength uint32
}

// HandlerConfiguration is the result of directives found in a 'Caddyfile'.
// Can be modified at runtime, except for values that are marked as 'read-only'.
type HandlerConfiguration struct {
	// Prefixes on which Caddy activates this plugin (read-only).
	// Order matters because scopes can overlap.
	PathScopes []string

	// Maps scopes (paths) to their own and potentially differently configurations.
	Scope map[string]*ScopeConfiguration
}

func parseCaddyConfig(c *caddy.Controller) (*HandlerConfiguration, error) {
	siteConfig := &HandlerConfiguration{
		PathScopes: make([]string, 0, 1),
		Scope:      make(map[string]*ScopeConfiguration),
	}

	for c.Next() {
		config := ScopeConfiguration{}
		config.UploadProgressCallback = noopUploadProgressCallback

		// Most likely only one path; but could be more.
		scopes := c.RemainingArgs()
		if len(scopes) == 0 {
			return siteConfig, c.ArgErr()
		}
		siteConfig.PathScopes = append(siteConfig.PathScopes, scopes...)

		for c.NextBlock() {
			key := c.Val()
			switch key {
			case "to":
				if !c.NextArg() {
					return siteConfig, c.ArgErr()
				}
				// Must be a directory.
				writeToPath := c.Val()
				finfo, err := os.Stat(writeToPath)
				if err != nil {
					return siteConfig, c.Err(err.Error())
				}
				if !finfo.IsDir() {
					return siteConfig, c.ArgErr()
				}
				config.WriteToPath = writeToPath
			case "key":
				if !c.NextArg() {
					return siteConfig, c.ArgErr()
				}
				config.Key = c.Val()
			case "silent_auth_errors":
				config.SilenceAuthErrors = true
			case "yes_without_tls":
				config.AcknowledgedNoTLS = true
			case "filename_length":
				if !c.NextArg() {
					return siteConfig, c.ArgErr()
				}
				l, err := strconv.ParseUint(c.Val(), 10, 32)
				if err != nil {
					return siteConfig, c.Err(err.Error())
				}
				config.FilenameLength = uint32(l)
			}
		}

		if config.WriteToPath == "" {
			return siteConfig, c.Errf("The destination path 'to' is missing")
		}
		if config.Key == "" {
			return siteConfig, c.Errf("The string 'key' is missing")
		}

		for idx := range scopes {
			siteConfig.Scope[scopes[idx]] = &config
		}
	}

	return siteConfig, nil
}

// noopUploadProgressCallback NOP-functor, set as default.
func noopUploadProgressCallback(bytesWritten uint64, err error) {
	// I want to become a closure that updates a data structure.
}
