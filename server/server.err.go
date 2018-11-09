package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dhalturin/release-manager/data"
)

// Err struct
type Err struct {
	Data    Errors
	Pointer string
	Meta    string
}

// Errors struct
type Errors struct {
	List []ErrList `json:"errors"`
}

// ErrList struct
type ErrList struct {
	Status int       `json:"status"`
	Code   int       `json:"code"`
	Detail string    `json:"detail"`
	Source ErrSource `json:"source"`
}

// ErrSource struct
type ErrSource struct {
	Pointer string `json:"pointer"`
	Meta    string `json:"meta"`
}

func (e *Err) get(codes []int) *Err {
	for _, code := range codes {
		Error := e.find(code)

		e.push(Error.Status, code, Error.Detail)
	}

	return e
}

func (e *Err) find(code int) data.ErrorStruct {
	return data.ErrorList[code]
}

func (e *Err) push(status int, code int, detail string) *Err {
	e.Data.List = append(e.Data.List, ErrList{
		status,
		code,
		detail,
		ErrSource{e.Pointer, e.Meta},
	})

	return e
}

func (e *Err) print(w http.ResponseWriter) {
	w.Write([]byte(e.join()))
}

func (e *Err) join() string {
	errorList := []string{}

	for _, j := range e.Data.List {
		errorList = append(errorList, fmt.Sprintf("> #%d: %s", j.Code, j.Detail))
	}

	return strings.Join(errorList, "\n")
}
