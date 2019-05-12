package punfed

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"text/tabwriter"

	"github.com/mholt/caddy/caddyhttp/httpserver"
)

type handler struct {
	Next   httpserver.Handler
	Config config
	User   string
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
			h.User = r.FormValue("user")
			return nil
		}
	}

	return fmt.Errorf("incorrect key")
}

func (h *handler) view(w http.ResponseWriter, r *http.Request) error {
	s, err := h.unstore()
	if err != nil {
		return err
	}

	t := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	for _, d := range s.Dates {
		fmt.Fprintln(t, d.Date.Format("2006-01-02"))
		for _, f := range d.Files {
			fmt.Fprintln(t, "https://"+path.Join(h.Config.Key, h.Config.Serve,
				f.Serve)+"\t"+f.Orig)
		}
		fmt.Fprintln(t)
	}

	return t.Flush()
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

		fn, err := h.generateFilename(f, fh.Filename)
		if err != nil {
			return err
		}

		o, err := os.Create(path.Join(h.getSaveDir(), fn))
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

		if err := h.store(fn, fh.Filename); err != nil {
			log.Println(err)
			return err
		}

		w.Header().Add("Location", path.Join(h.Config.Serve, fn))
		fmt.Fprintln(w, "https://"+path.Join(h.Config.Key, h.Config.Serve, fn))
	}

	return nil
}
