package main

import (
	"log"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"mime/multipart"

	"github.com/h2non/filetype"
	"github.com/jmcvetta/randutil"
)

func generateFilename(f multipart.File, h *multipart.FileHeader) (string,
	error) {
	r, err := randutil.AlphaString(config.FilenameLength)
	if err != nil {
		return "", nil
	}

	t, err := filetype.MatchReader(f)
	if err != nil {
		return "", err
	}
	if t == filetype.Unknown {
		t.Extension = path.Ext(h.Filename)
	} else {
		t.Extension = "." + t.Extension
	}

	return r + t.Extension, nil
}

func keyHandler(w http.ResponseWriter, r *http.Request) error {
	k := key{r.Form["user"][0], r.Form["pass"][0]}
	for _, ck := range config.Keys {
		if ck == k {
			return nil
		}
	}

	return fmt.Errorf("incorrect key")
}

func fileHandler(w http.ResponseWriter, r *http.Request) error {
	fl := r.MultipartForm.File["files[]"]
	for i, h := range fl {
		f, err := fl[i].Open()
		if err != nil {
			return err
		}
		defer f.Close()

		fn, err := generateFilename(f, h)
		if err != nil {
			return err
		}

		n, err := os.Create(path.Join(config.Dest, fn))
		if err != nil {
			return err
		}
		defer n.Close()

		if _, err := io.Copy(n, f); err != nil {
			return err
		}

		fmt.Fprintln(w, path.Join(config.URL, fn))
	}

	return nil
}

// TODO: Check for POST Postform/Form
func handler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		log.Println(err)
		return
	}

	if err := keyHandler(w, r); err != nil {
		fmt.Fprintln(w, err)
		return
	}

	if err := fileHandler(w, r); err != nil {
		log.Println(err)
		return
	}
}

func main () {
	log.Println("Starting...")

	if err := parseConfig(); err != nil {
		log.Fatalln(err)
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(config.URL, nil)
}
