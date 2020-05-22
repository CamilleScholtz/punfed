package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"text/tabwriter"

	urljoin "github.com/shimohq/go-url-join"
)

type handler struct {
	Config *config
	User   string
}

func (h *handler) authenticate(w http.ResponseWriter, r *http.Request) error {
	k := key{r.FormValue("user"), r.FormValue("pass")}
	for _, ck := range h.Config.AcceptedKeys {
		if ck == k {
			h.User = r.FormValue("user")
			return nil
		}
	}

	return fmt.Errorf("Forbidden")
}

func (h *handler) upload(w http.ResponseWriter, r *http.Request) error {
	fl := r.MultipartForm.File["files[]"]

	for i, fh := range fl {
		if fh.Size > h.Config.MaxFileSize {
			return fmt.Errorf("Payload Too Large")
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

		o, err := os.Create(path.Join(h.getWritePath(), fn))
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

		fmt.Fprintln(w, urljoin.Join(h.Config.Root, h.Config.ServePath, fn))
	}

	return nil
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
			fmt.Fprintln(t, urljoin.Join("https://", h.Config.Root, h.Config.
				ServePath, f.Serve)+"\t"+f.Orig)
		}

		if i != len(s.Dates)-1 {
			fmt.Fprintln(t)
		}
	}

	return t.Flush()
}

func main() {
	c, err := parseConfig()
	if err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		h := &handler{Config: c}

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", 405)
			return
		}

		// TODO: Do I need to use `MaxFileSize` here?
		if err := r.ParseMultipartForm(h.Config.MaxFileSize); err != nil {
			http.Error(w, "Payload Too Large", 413)
			return
		}

		if err := h.authenticate(w, r); err != nil {
			http.Error(w, "Forbidden", 403)
			return
		}

		if r.FormValue("function") == "view" {
			err = h.view(w, r)
		} else {
			err = h.upload(w, r)
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	})
	if err := http.ListenAndServe(c.Listen, nil); err != nil {
		log.Fatalln(err)
	}
}
