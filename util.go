package punfed

import (
	"io"
	"mime/multipart"
	"path"

	"github.com/h2non/filetype"
	"github.com/jmcvetta/randutil"
)

func (h *handler) getSaveDir() string {
	return path.Join(h.Config.ServePath, h.User)
}

func (h *handler) getStoreFile() string {
	return path.Join(h.getSaveDir(), ".punfed.json")
}

func (h *handler) generateFilename(f multipart.File, fn string) (string,
	error) {
	r, err := randutil.AlphaString(h.Config.RandomFilenameLenght)
	if err != nil {
		return "", nil
	}

	t, err := filetype.MatchReader(f)
	if err != nil {
		return "", err
	}
	if t == filetype.Unknown {
		t.Extension = path.Ext(fn)
	} else {
		t.Extension = "." + t.Extension
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	return r + t.Extension, nil
}
