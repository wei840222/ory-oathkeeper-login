package server

import "errors"

var (
	ErrInvalidSession = errors.New("invalid session")
)

type ErrorRes struct {
	Error string `json:"error"`
}
