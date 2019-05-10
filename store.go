package punfed

import (
	"os"
	"time"

	"github.com/burntSushi/toml"
)

type store struct {
	Dates []date
}

type date struct {
	Date  time.Time
	Files []file
}

type file struct {
	Serve string
	Orig  string
}

func (h *handler) store(fn, ofn string) error {
	f, err := os.OpenFile(h.getStoreFile(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0644)
	if err != nil {
		return err
	}
	defer f.Close()

	s, err := h.unstore()
	if err != nil {
		return err
	}

	n := time.Now()
	l := s.Dates[len(s.Dates)-1]
	if n.Year() != l.Date.Year() || n.YearDay() != l.Date.YearDay() {
		s.Dates = append(s.Dates, date{n, []file{}})
	}
	s.Dates[len(s.Dates)-1].Files = append(s.Dates[len(s.Dates)-1].Files, file{
		fn, ofn})

	return toml.NewEncoder(f).Encode(s)
}

func (h *handler) unstore() (store, error) {
	s := store{}
	if _, err := toml.DecodeFile(h.getStoreFile(), &s); err != nil {
		return s, err
	}

	return s, nil
}
