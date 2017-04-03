package punfed

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"blitznote.com/src/caddy.upload/protofile"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/pkg/errors"
)

const (
	// At this point this is an arbitrary number.
	reportProgressEveryBytes = 1 << 15
)

// Errors used in functions that resemble the core logic of this plugin.
const (
	errCannotReadMIMEMultipart coreUploadError = "Error reading MIME multipart payload"
	errFilenameConflict        coreUploadError = "Name-Name Conflict"
	errInvalidFilename         coreUploadError = "Invalid filename and/or path"
	errNoDestination           coreUploadError = "A destination is missing"
	errUnknownEnvelopeFormat   coreUploadError = "Unknown envelope format"
	errLengthInvalid           coreUploadError = "Field 'length' has been set, but is invalid"
	errNoKey                   coreUploadError = "No key provided"
	errInvalidKey              coreUploadError = "Invalid key provided"
	errNoFiles                 coreUploadError = "No files provided"
)

// coreUploadError is returned for errors that are not in a leaf method,
// that have no specialized error.
type coreUploadError string

// Error implements the error interface.
func (e coreUploadError) Error() string { return string(e) }

// Handler represents a configured instance of this plugin for uploads.
// If you want to use it outside of Caddy, then implement 'Next' as
// something with method ServeHTTP and at least the same member variables
// that you can find here.
type Handler struct {
	Next   httpserver.Handler
	Config HandlerConfiguration
}

// genRand returns some printable chars, meant to be used as
// randomized filenames.
func genRand(wantedLength uint32) string {
	suffix := make([]byte, wantedLength, wantedLength)
	// Most sources of randomness return full words; don't use N times rand.Int31().
	rand.Seed(time.Now().UnixNano())
	rand.Read(suffix)

	for idx, c := range suffix {
		c = (c % 36)
		if c <= 9 {
			c += 48 // 48–57 → 0–9
		} else {
			c += 87 // 97–122 → a–z
		}
		suffix[idx] = c
	}

	return string(suffix)
}

// ServeHTTP catches methods if meant for file manipulation, else is a passthrough.
// Directs HTTP methods and fields to the corresponding function calls.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	var (
		// A prefix we will need to replace with the target directory.
		sc   string
		conf *ScopeConfiguration
	)

	if r.Method == http.MethodPost {
		// Iterate over the scopes in the order they have been defined.
		for _, sc = range h.Config.PathScopes {
			if httpserver.Path(r.URL.Path).Matches(sc) {
				conf = h.Config.Scope[sc]
				goto inScope
			}
		}
	}
	return h.Next.ServeHTTP(w, r)

inScope:
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		return h.ServeMultipartUpload(w, r, sc, conf)
	}
	return http.StatusUnsupportedMediaType, errUnknownEnvelopeFormat
}

// ServeMultipartUpload is used on HTTP POST to explode a MIME Multipart envelope
// into one or more supplied files. They are then supplied to WriteOneHTTPBlob one by one.
func (h *Handler) ServeMultipartUpload(w http.ResponseWriter, r *http.Request, sc string, conf *ScopeConfiguration) (int, error) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		return http.StatusUnsupportedMediaType, errCannotReadMIMEMultipart
	}

	k := r.PostFormValue("key")
	if k == "" {
		return http.StatusUnauthorized, errNoKey
	}
	if k != conf.Key {
		return http.StatusUnauthorized, errInvalidKey
	}

	fl := r.MultipartForm.File["file"]
	if len(fl) == 0 {
		return http.StatusBadRequest, errNoFiles
	}

	for i, f := range fl {
		fr, err := f.Open()
		if err != nil {
			return http.StatusBadRequest, err
		}
		_, rv, err := h.WriteOneHTTPBlob(w, r, sc, conf, f.Filename, f.Header.Get("Content-Length"), fr)
		if err != nil {
			// Don't use the Filename here: it is controlled by the user.
			return rv, errors.Wrap(err, "MIME Multipart exploding failed on part "+strconv.Itoa(i+1))
		}
	}

	return http.StatusCreated, nil
}

// WriteOneHTTPBlob handles HTTP PUT (and HTTP POST without envelopes),
// writes one file to disk by adapting WriteFileFromReader to HTTP conventions.
func (h *Handler) WriteOneHTTPBlob(w http.ResponseWriter, r *http.Request, sc string, conf *ScopeConfiguration, fn, as string, fr io.Reader) (uint64, int, error) {
	eb, _ := strconv.ParseUint(as, 10, 64)
	if as != "" && eb <= 0 {
		return 0, http.StatusLengthRequired, errLengthInvalid
		// Usually 411 is used for the outermost element.
		// We don't require any length; but it must be valid if given.
	}

	if conf.FilenameLength > 0 {
		fn = genRand(conf.FilenameLength) + filepath.Ext(fn)
	} else {
		fn = genRand(4) + filepath.Ext(fn)
	}

	cb := conf.UploadProgressCallback
	if cb == nil {
		cb = noopUploadProgressCallback
	}
	bw, err := WriteFileFromReader(conf.WriteToPath, fn, fr, eb, cb)
	if err != nil {
		if os.IsExist(err) || strings.HasSuffix(err.Error(), "not a directory") {
			// 409.
			return 0, http.StatusConflict, errFilenameConflict
		}
		if bw > 0 && bw < eb {
			// 507: Insufficient storage.
			return bw, http.StatusInsufficientStorage, err
		}
		return bw, http.StatusInternalServerError, err
	}

	fmt.Fprintln(w, "https://"+r.Host+"/"+filepath.Join(strings.TrimPrefix(conf.WriteToPath, sc), fn))

	if bw < eb {
		// 202: Accepted (but not completed).
		return bw, http.StatusAccepted, nil
	}
	// 201: Created.
	return bw, http.StatusCreated, nil
}

// WriteFileFromReader implements an unit of work consisting of
// • creation of a temporary file,
// • writing to it,
// • discarding it on failure ('zap') or
// • its "emergence" ('persist') into observable namespace.
//
// If 'anticipatedSize' ≥ protofile.reserveFileSizeThreshold (usually 32 KiB)
// then disk space will be reserved before writing (by a ProtoFileBehaver).
//
// With uploadProgressCallback:
// The file has been successfully written if "error" remains 'io.EOF'.
func WriteFileFromReader(p, fn string, r io.Reader, as uint64, cb func(uint64, error)) (uint64, error) {
	wp, err := protofile.IntentNew(p, fn)
	if err != nil {
		return 0, err
	}
	w := *wp
	defer w.Zap()

	err = w.SizeWillBe(as)
	if err != nil {
		return 0, err
	}

	var bw uint64
	var n int64
	for err == nil {
		n, err = io.CopyN(w, r, reportProgressEveryBytes)
		if err == nil || err == io.EOF {
			bw += uint64(n)
			cb(bw, err)
		}
	}

	if err != nil && err != io.EOF {
		return bw, err
	}
	err = w.Persist()
	if err != nil {
		cb(bw, err)
	}
	return bw, err
}
