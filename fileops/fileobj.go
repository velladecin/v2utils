package fileops
import (
    "syscall"
    "io/fs"
    "errors"
)
const (
    // status
    File_std = iota
    File_new
    File_cut
    File_mis
    File_chg
)

type FileObj interface {
    Path() string
    Ino() uint64
    Exists() bool
    Comms() chan Notify
}
type Notify interface {
    Path() string
    Status() uint8
    Data() any
}

func stat(path string) syscall.Stat_t {
    var stat syscall.Stat_t
    err := syscall.Stat(path, &stat)
    if err != nil {
        if ! errors.Is(err, fs.ErrNotExist) {
            panic(err)
        }
    }
    return stat
}
