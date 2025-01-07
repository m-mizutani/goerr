package main

import (
	"net/http"

	"github.com/m-mizutani/goerr"
)

var (
	ErrTagSysError   = goerr.NewTag("system_error")
	ErrTagBadRequest = goerr.NewTag("bad_request")
)

func handleError(w http.ResponseWriter, err error) {
	if goErr := goerr.Unwrap(err); goErr != nil {
		switch {
		case goErr.HasTag(ErrTagSysError):
			w.WriteHeader(http.StatusInternalServerError)
		case goErr.HasTag(ErrTagBadRequest):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	_, _ = w.Write([]byte(err.Error()))
}

func someAction() error {
	if _, err := http.Get("http://example.com/some/resource"); err != nil {
		return goerr.Wrap(err, "failed to get some resource", goerr.T(ErrTagSysError))
	}
	return nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := someAction(); err != nil {
			handleError(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// #nosec
	http.ListenAndServe(":8090", nil)
}
