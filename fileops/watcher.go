package fileops
import (
    "time"
)

type fileChange struct {
    // path to file
    path string

    // File_std(0) - no
    // File_chg(4) - yes
    change uint8

    // status of file or type of change
    status uint8
}
func (fc fileChange) Path() string { return fc.path }
func (fc fileChange) Status() uint8 { return fc.status }
func (fc fileChange) Data() any { return fc.change }

type fileWatch struct {
    path string
    ino uint64 
    ctime int64
    comms chan Notify
}

func (fw *fileWatch) Path() string { return fw.path }
func (fw *fileWatch) Ino() uint64 { return fw.ino }
func (fw *fileWatch) Comms() chan Notify { return fw.comms }
func (fw *fileWatch) Exists() bool {
    if fw.ino == 0 && fw.ctime == 0 {
        return false
    }
    return true
}
func (fw *fileWatch) updateInode() {
    stat := stat(fw.path)
    fw.ino = stat.Ino
    fw.ctime = stat.Ctim.Sec
}
func NewWatcher(path string) FileObj {
    fw := &fileWatch{path, 0, 0, make(chan Notify)}
    fw.updateInode()

    go func(fw *fileWatch) {
        var ino uint64
        var ctime int64
        for {
            ino = fw.ino
            ctime = fw.ctime
            fw.updateInode()

            // default - no change
            var c uint8 = File_std
            var s uint8 = File_std
            if ino == fw.ino && ino == 0 {
                // no change but (still) missing
                s = File_mis
            } else if ino != fw.ino {
                // change - ino mismatch
                if fw.ino == 0 {
                    // disappeared
                    c = File_chg
                    s = File_mis
                } else {
                    // re-appeared or was replaced/rotated
                    c = File_chg
                    s = File_new
                }
            } else if ctime != fw.ctime {
                // change - ctime mismatch
                c = File_chg
            }

            fw.comms <- fileChange{fw.path, c, s}
            time.Sleep(1 * time.Second)
        }
    }(fw)

    return fw
}
