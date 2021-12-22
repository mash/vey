package http

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return h(w, r)
}

func Wrap(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.ServeHTTP(w, r)
		if err != nil {
			var er Error
			if !errors.As(err, &er) {
				er = Error{
					Code: http.StatusInternalServerError,
					Msg:  http.StatusText(http.StatusInternalServerError),
					Err:  err,
				}
			}
			Log.Error(err)
			_ = WriteJSON(w, er.Code, er)
		}
	})
}

func WrapF(f func(http.ResponseWriter, *http.Request) error) http.Handler {
	return Wrap(HandlerFunc(f))
}

type BodyFunc func(http.ResponseWriter, *http.Request, Body) error

// AcceptJSON parses the request body into Body struct.
// Be careful that, the request body is EOF when later trying to read it again.
// AcceptJSON responds and doesn't call h, if JSON decoding to Body struct fails.
func AcceptJSON(h BodyFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		// Max 1MB body
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)

		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()

		var body Body
		if err := dec.Decode(&body); err != nil {
			return Error{
				Code: http.StatusBadRequest,
				Msg:  "json decoding failed",
				Err:  err,
			}
		}

		return h(w, r, body)
	}
}

func WriteJSON(w http.ResponseWriter, statusCode int, out interface{}) error {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return enc.Encode(out)
}
