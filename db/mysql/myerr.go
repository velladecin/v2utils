package mysql

import (
    "fmt"
)

type UnhandledDbTypeErr struct {
    ttype string
}

func (e *UnhandledDbTypeErr) Error() string {
    msg := "Unknown DB type"

    if e.ttype != "" {
        msg += fmt.Sprintf(": %s", e.ttype)
    }

    return msg
}
