package fileops
import (
    "fmt"
    "os"
    "io"
    "time"
)
const (
    // buffer
    buff_min = 1024
    buff_std = 1024*1024
    buff_max = 1024*1024*1024

    // milliseconds
    loop_zzzZZzz = 750
)

// Config
type TailConf func(*fileTail)
func BufferSize(buffer int) func(*fileTail) {
    if buffer < buff_min || buffer > buff_max {
        panic(fmt.Sprintf("Tail buffer overflow: %d .. %d", buff_min, buff_max))
    }
    return func(ft *fileTail) {
        ft.buff = buffer
    }
}
func Reopen(b bool) func(*fileTail) {
    return func(ft *fileTail) {
        ft.reopen = b
    }
}

// Chunk
type Tchunk struct {
    path string
    status uint8
    s string
}
func (tc Tchunk) Path() string {
    return tc.path
}
func (tc Tchunk) Status() uint8 {
    return tc.status
}
func (tc Tchunk) Data() any {
    return tc.s
}

// TODO do a first byte to be able to check file overwrite within size??
// C tail somehow knows when echo "bla" > file.txt is done even when the text is same.. :?

// Tail
type fileTail struct {
    path string
    fh *os.File
    ino uint64
    size, pos int64
    buff int
    reopen bool // -f vs -F
    status uint8
    comms chan Notify
}
func (ft *fileTail) Path() string { return ft.path }
func (ft *fileTail) Ino() uint64 { return ft.ino }
func (ft *fileTail) Comms() chan Notify { return ft.comms }
func (ft *fileTail) Exists() bool {
    if ft.ino == 0 && ft.size == 0 {
        return false
    }
    return true
}
func (ft *fileTail) close() {
    ft.fh.Close()
    ft.ino = 0
    ft.size = 0
    ft.pos = 0
}
func (ft *fileTail) updateInode() {
    stat := stat(ft.path)
    ft.ino = stat.Ino
    ft.size = stat.Size
    //fmt.Println(ft.fh.Fd())
}
func (ft *fileTail) openFile() {
    // TODO what info does *File provide,
    //      can we update the inode from it?
    fh, err := os.Open(ft.path)
    if err != nil {
        panic(err)
    }
    ft.fh = fh
    ft.updateInode()
}
func (ft *fileTail) seekFileStart()      { ft.seekFile(0, os.SEEK_SET) }
func (ft *fileTail) seekFileEnd()        { ft.seekFile(0, os.SEEK_END) }
func (ft *fileTail) seekFileSet(i int64) { ft.seekFile(i, os.SEEK_SET) }
func (ft *fileTail) seekFile(offset int64, whence int) {
    pos, err := ft.fh.Seek(offset, whence)
    if err != nil {
        panic(err)
    }
    ft.pos = pos
}
func (ft *fileTail) readFile(bytes []byte) int {
    p, err := ft.fh.Read(bytes)
    if err != nil && err != io.EOF {
        panic(err)
    }
    ft.pos += int64(p)
    ft.seekFileSet(ft.pos)
    return p
}

func NewTail(path string, conf ...TailConf) FileObj {
    ft := &fileTail{path, nil, 0, 0, 0, buff_std, false, File_std, make(chan Notify)}
    ft.updateInode()
    //stat := stat(path)
    //ft := &fileTail{path, nil, stat.Ino, 0, 0, buff_std, false, File_std, make(chan Notify)}
    if ! ft.Exists() {
        panic("No such file or directory: " + ft.path)
    }
    for _, tconf := range conf {
        tconf(ft)
    }

    go func(ft *fileTail) {
        ft.openFile()
        defer ft.close()
        ft.seekFileEnd()
        ft.status = File_std

        var ino uint64
        var size int64
        var bytes = make([]byte, ft.buff)
        for {
            ino = ft.ino
            size = ft.size
            ft.updateInode()
            ft.status = File_std

            // missing file continues to be missing :)
            if ino == ft.ino && ino == 0 {
                time.Sleep(time.Duration(loop_zzzZZzz) * time.Millisecond)
                continue
            }

            // check if ino has changed
            // check size only if ino not changed

            if ino == ft.ino {
                // truncated
                if size > ft.size {
                    ft.status = File_cut
                    ft.seekFileStart()
                }
            } else {
                // not found
                if ft.ino == 0 {
                    fmt.Println("disappeared")
                    ft.status = File_mis
                    ft.close()
                    // TODO - send smth to nofify of file going missing?
                    ft.comms <- Tchunk{ft.path, ft.status, ""}
                    continue
                }

                // ino has changed
                // old
                ft.close()
                // new
                ft.openFile()
                defer ft.close()
                ft.seekFileStart()
                ft.status = File_new
            }

            n := ft.readFile(bytes)
            if n == 0 {
                time.Sleep(time.Duration(loop_zzzZZzz) * time.Millisecond)
                continue
            }

            ft.comms <- Tchunk{ft.path, ft.status, string(bytes[:n])} 
            // zero out the bytes slice to reset it for next read
            // used to be in a goroutine but that felt somehow risky..
            for i, j := 0, n-1; i<=j; i, j = i+1, j-1 {
                bytes[i], bytes[j] = 0, 0
            }
            time.Sleep(time.Duration(loop_zzzZZzz) * time.Millisecond)
        }
    }(ft)
    
    return ft
}
