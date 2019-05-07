package punfed

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

type handler struct {
	Next   httpserver.Handler
	Config config
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int,
	error) {
	if r.Method != http.MethodPost || !httpserver.Path(r.URL.Path).Matches(
		h.Config.Scope) {
		return h.Next.ServeHTTP(w, r)
	}

	if err := r.ParseMultipartForm(h.Config.Max); err != nil {
		return http.StatusInternalServerError, err
	}

	if err := h.key(w, r); err != nil {
		return http.StatusUnauthorized, err
	}

	if r.FormValue("function") == "view" {
		if err := h.view(w, r); err != nil {
			return http.StatusInternalServerError, err
		}
		return http.StatusOK, nil
	}

	if err := h.upload(w, r); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusCreated, nil
}

func (h *handler) key(w http.ResponseWriter, r *http.Request) error {
	k := key{r.FormValue("user"), r.FormValue("pass")}
	for _, ck := range h.Config.Keys {
		if ck == k {
			return nil
		}
	}

	return fmt.Errorf("incorrect key")
}

func (h *handler) view(w http.ResponseWriter, r *http.Request) error {
	fl, err := ioutil.ReadDir(path.Join(h.Config.Save, r.FormValue("user")))
	if err != nil {
		return err
	}

	for _, f := range fl {
		fmt.Fprintln(w, "https://"+path.Join(h.Config.Key, h.Config.Serve, f.
			Name()))
	}

	return nil
}

func (h *handler) upload(w http.ResponseWriter, r *http.Request) error {
	fl := r.MultipartForm.File["files[]"]
	for i, fh := range fl {
		if fh.Size > h.Config.Max {
			return fmt.Errorf("file too large")
		}

		f, err := fl[i].Open()
		if err != nil {
			return err
		}
		defer f.Close()

		fn, err := generateFilename(h.Config.Len, f, fh)
		if err != nil {
			return err
		}

		o, err := os.Create(path.Join(h.Config.Save, r.FormValue("user"), fn))
		if err != nil {
			return err
		}
		defer o.Close()

		if _, err := io.Copy(o, f); err != nil {
			return err
		}
		if err := o.Sync(); err != nil {
			return err
		}

		w.Header().Add("Location", path.Join(h.Config.Serve, fn))
		fmt.Fprintln(w, "https://"+path.Join(h.Config.Key, h.Config.Serve, fn))
	}

	return nil
}
