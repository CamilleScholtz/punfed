package punfed

import (
	"io"
	"net/http"
	"os"
	"path"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

type handler struct {
	Next   httpserver.Handler
	Config config
}

// TODO: Check for POST Postform/Form
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int,
	error) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return 0, err
	}

	if err := h.key(w, r); err != nil {
		return 0, err
	}

	if err := h.file(w, r); err != nil {
		return 0, err
	}

	return 0, nil
}

func (h *handler) key(w http.ResponseWriter, r *http.Request) error {
	/*k := key{r.Form["user"][0], r.Form["pass"][0]}
	for _, ck := range config.Keys {
		if ck == k {
			return nil
		}
	}

	return fmt.Errorf("incorrect key")*/
	return nil
}

func (h *handler) file(w http.ResponseWriter, r *http.Request) error {
	fl := r.MultipartForm.File["files[]"]
	for i, fh := range fl {
		f, err := fl[i].Open()
		if err != nil {
			return err
		}
		defer f.Close()

		fn, err := generateFilename(h.Config.Len, f, fh)
		if err != nil {
			return err
		}

		n, err := os.Create(path.Join(h.Config.Dest, fn))
		if err != nil {
			return err
		}
		defer n.Close()

		if _, err := io.Copy(n, f); err != nil {
			return err
		}

		//fmt.Fprintln(w, path.Join(config.URL, fn))
	}

	return nil
}
