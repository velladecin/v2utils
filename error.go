package v2utils

import (
    "fmt"
)

type Error struct {
    err, ctx string
}

func NewError() *Error {
    return &Error{}
}

func (e *Error) Error() string {
    if e.ctx == "" {
        e.ctx = "v2utils"
    }

    return fmt.Sprintf("%s: %s", e.ctx, e.err)
}

func (e *Error) Err(r string) *Error { e.err = r; return e }
func (e *Error) Ctx(c string) *Error { e.ctx = c; return e }

func Err(e string)       *Error { return NewError().Err(e) }
func ErrCtx(e, c string) *Error { return NewError().Err(e).Ctx(c) }
