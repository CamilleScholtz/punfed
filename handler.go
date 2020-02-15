package punfed

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"text/tabwriter"
)

type handler struct {
	Next   http.Handler
	Config *ScopeConfiguration
	Scope  string
	User   string
}

// NewHandler creates a new instance of this plugin's upload handler.
func NewHandler(s string, c *ScopeConfiguration, n http.Handler) (
	*handler, error) {
	h := handler{
		Next:   n,
		Config: c,
		Scope:  s,
	}

	if n == nil {
		h.Next = http.NotFoundHandler()
	}

	return &h, nil
}

// ServeHTTP handles any uploads, else defers the request to the next handler.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var cn bool

	httpCode, err := h.serveHTTP(w, r, h.Scope, func(w http.ResponseWriter,
		r *http.Request) (int, error) {
		cn = true
		return 0, nil
	})

	if cn {
		h.Next.ServeHTTP(w, r)
		return
	}

	if httpCode >= 400 && err != nil {
		http.Error(w, err.Error(), httpCode)
	} else {
		w.WriteHeader(httpCode)
	}
}

func (h *handler) serveHTTP(w http.ResponseWriter, r *http.Request, s string,
	nf func(http.ResponseWriter, *http.Request) (int, error)) (
	int, error) {
	//|| !httpserver.Path(r.URL.Path).Matches(s)

	if r.Method != http.MethodPost {
		return nf(w, r)
	}

	if err := r.ParseMultipartForm(h.Config.MaxFilesize); err != nil {
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
	for _, ck := range h.Config.AcceptedKeys {
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
	for i, d := range s.Dates {
		fmt.Fprintln(t, d.Date.Format("* 2006-01-02"))

		for _, f := range d.Files {
			fmt.Fprintln(t, "https://"+path.Join(h.Config.ServePath, f.Serve)+
				"\t"+f.Orig)
		}

		if i != len(s.Dates)-1 {
			fmt.Fprintln(t)
		}
	}

	return t.Flush()
}

func (h *handler) upload(w http.ResponseWriter, r *http.Request) error {
	fl := r.MultipartForm.File["files[]"]
	for i, fh := range fl {
		if fh.Size > h.Config.MaxFilesize {
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
			return err
		}

		w.Header().Add("Location", path.Join(h.Config.ServePath, fn))
		fmt.Fprintln(w, "https://"+path.Join(h.Config.ServePath, fn))
	}

	return nil
}
