package v2utils

import (
    "regexp"
    "bufio"
    "os"
    "time"
    "syscall"
)

// test
const (
    T_START = iota
    T_CUR
    T_END

    seek_set = T_START
    seek_cur = T_CUR
    seek_end = T_END
)

type FileInfo struct {
    Path string
    Inode uint64
    Mode uint32 
    Uid uint32
    Gid uint32
    Size int64
    Atime, Mtime, Ctime int64
}

var comment = regexp.MustCompile(`^\s*#|^\/\/`)
var emptyLn = regexp.MustCompile(`^\s*$`)

func ReadFileClean  (path string)                       ([]string, error) { return ReadFileNoMatch(path, []*regexp.Regexp{comment, emptyLn}) }
func ReadFileMatch  (path string, rgx []*regexp.Regexp) ([]string, error) { return readFile(path, rgx, true) }
func ReadFileNoMatch(path string, rgx []*regexp.Regexp) ([]string, error) { return readFile(path, rgx, false) }
func readFile(path string, rgx []*regexp.Regexp, match bool) ([]string, error) {
    lines, err := ReadFile(path)
    if err != nil {
        return lines, err
    }

    wanted := make([]string, 0)

    L_line: for _, line := range lines {
        L_rgx: for _, r := range rgx {
            ok := r.MatchString(line)

            // positive match
            if match {
                if ok {
                    wanted = append(wanted, line)
                    continue L_line
                }

                continue L_rgx
            }

            // negative match
            if ok {
                continue L_line
            }
        }

        // no rgx matched
        // keep this line in case of negative match

        if ! match {
            wanted = append(wanted, line)
        }
    }

    return wanted, nil
}

func ReadFile(path string) ([]string, error) {
    var lines = []string{}

    fi, err := os.Open(path)
    if err != nil {
        return lines, err
    }
    defer fi.Close()

    r := bufio.NewScanner(fi)
    for r.Scan() {
        lines = append(lines, r.Text())
    }

    err = r.Err()
    if err != nil {
        lines = []string{} // remove any remnants
    }

    return lines, err
}

func Stat(path string) (FileInfo, error) {
    var finfo FileInfo = FileInfo{Path: path}

    stat, err := os.Stat(path)
    if err != nil {
        return finfo, err
    }

    return getLocalFileInfo(stat)
}

func StatF(f *os.File) (FileInfo, error) {
    var finfo FileInfo = FileInfo{Path: f.Name()}

    stat, err := f.Stat()
    if err != nil {
        return finfo, err
    }

    return getLocalFileInfo(stat)
}

func getLocalFileInfo(fi os.FileInfo) (FileInfo, error) {
    var finfo FileInfo = FileInfo{Path: fi.Name()}
    var s *syscall.Stat_t = fi.Sys().(*syscall.Stat_t)

    finfo.Inode = s.Ino
    finfo.Mode = s.Mode
    finfo.Uid = s.Uid
    finfo.Gid = s.Gid
    finfo.Size = fi.Size()
    finfo.Atime = s.Atim.Sec
    finfo.Mtime = s.Mtim.Sec
    finfo.Ctime = s.Ctim.Sec

    return finfo, nil
}

func openAndSeek(path string, seekfrom int) *os.File {
    fi, err := os.Open(path)
    if err != nil {
        panic(err)
    }

    _, err = fi.Seek(0, seekfrom)
    if err != nil {
        panic(err)
    }

    return fi
}

func Tail(path string, from int) chan []string {
    stat, err := Stat(path)
    if err != nil {
        panic(err)
    }

    size := stat.Size
    inode := stat.Inode
    ch := make(chan []string)

    go func() {
        fi := openAndSeek(path, from)
        defer fi.Close()

        for {
            // make sure we actually have a file before moving forward
            // stat only returns 'no such file' error (see docs for "os")

            var finfo FileInfo
            var err error

            for {
                finfo, err = Stat(path)
                if err != nil {
                    ch <- []string{err.Error()}

                    time.Sleep(time.Second)
                    continue
                }

                break
            }

            if finfo.Size < size || finfo.Inode != inode {
                // smth happened, eg: logrotate
                fi.Close()

                fi = openAndSeek(path, seek_set)
                defer fi.Close()

                inode = finfo.Inode
            }

            size = finfo.Size
            var lines []string

            r := bufio.NewScanner(fi)
            for r.Scan() {
                lines = append(lines, r.Text())
            }

            err = r.Err()
            if err != nil {
                panic(err)
            }

            if len(lines) > 0 {
                ch <- lines
            }

            _, err = fi.Seek(0, seek_cur)
            if err != nil {
                panic(err)
            }

            time.Sleep(time.Second)
        }
    }()

    return ch
}
