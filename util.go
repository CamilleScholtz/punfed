package punfed

import (
	"mime/multipart"
	"path"

	"github.com/h2non/filetype"
	"github.com/jmcvetta/randutil"
)

func generateFilename(l int, f multipart.File, h *multipart.FileHeader) (string,
	error) {
	r, err := randutil.AlphaString(l)
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
