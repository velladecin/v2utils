package snmp

import (
    v2 "vella/v2utils"
)

var (
    errInvalidTarget        = v2.Err("invalid target")
    errRetriesOutOfRange    = v2.Err("max retries out of range")
    errTimeoutOutOfRange    = v2.Err("timeout out of range")
    errRepetitionsOutOfRange = v2.Err("repetitions out of range")
)
