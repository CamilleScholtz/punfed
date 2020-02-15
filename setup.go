package punfed

// ScopeConfiguration represents the settings of a scope (URL path).
type ScopeConfiguration struct {
	// The maximum filesize.
	MaxFilesize int64

	// Target directory on disk that serves as upload destination.
	WritePath string

	// Uploaded files can be gotten back from here.
	ServePath string

	// The lenght of generated random filenames.
	RandomFilenameLenght int

	// The accepted username & password combinations.
	AcceptedKeys []key
}

type key struct {
	User string
	Pass string
}

// NewDefaultConfiguration creates a new default configuration.
func NewDefaultConfiguration(targetDirectory string) *ScopeConfiguration {
	cfg := ScopeConfiguration{
		MaxFilesize:          2000 << 20,
		RandomFilenameLenght: 4,
	}

	return &cfg
}
