package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
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
	s, err := h.unstore()
	if err != nil {
		return err
	}

	n := time.Now()
	if len(s.Dates) == 0 {
		s.Dates = append(s.Dates, date{n, []file{}})
	} else {
		l := s.Dates[len(s.Dates)-1]
		if n.Year() != l.Date.Year() || n.YearDay() != l.Date.YearDay() {
			s.Dates = append(s.Dates, date{n, []file{}})
		}
	}
	s.Dates[len(s.Dates)-1].Files = append(s.Dates[len(s.Dates)-1].Files, file{
		fn, ofn})

	ns, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(h.getStoreFile(), ns, 0666)
}

func (h *handler) unstore() (store, error) {
	s := store{}

	f, err := os.OpenFile(h.getStoreFile(), os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return s, err
	}
	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		return s, err
	}

	if len(d) > 0 {
		if err := json.Unmarshal(d, &s); err != nil {
			return s, err
		}
	}

	return s, nil
}
