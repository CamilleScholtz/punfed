package punfed

import (
	"fmt"
	"io"
	"log"
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
	// TODO: Scope doesn't work
	log.Println(r.URL.Path)
	log.Println(h.Config.Scope)
	if r.Method != http.MethodPost || !httpserver.Path(r.URL.Path).Matches(
		h.Config.Scope) {
		return h.Next.ServeHTTP(w, r)
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		return http.StatusBadRequest, err
	}

	if err := h.key(w, r); err != nil {
		return http.StatusUnauthorized, err
	}

	if err := h.file(w, r); err != nil {
		return http.StatusBadRequest, err
	}

	return http.StatusCreated, nil
}

func (h *handler) key(w http.ResponseWriter, r *http.Request) error {
	k := key{r.Form["user"][0], r.Form["pass"][0]}
	for _, ck := range h.Config.Keys {
		if ck == k {
			return nil
		}
	}

	return fmt.Errorf("incorrect key")
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

		fmt.Fprintln(w, path.Join(h.Config.Dest, fn))
	}

	return nil
}
