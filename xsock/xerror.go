package xsock

import (
    v2 "vella/v2utils"
)

var (
    // context supplied via err.Ctx(context)
    // (see in contex doh..)
    errInvalidId        = v2.Err("invalid id")
    errExpiredId        = v2.Err("expired id")
    errBufferOverflow   = v2.Err("buffer overflow")
    errLength           = v2.Err("invalid length")
    errType             = v2.Err("invalid type")
    errHeader           = v2.Err("invalid header")
    errTransmitTimeout  = v2.Err("transmission timeout")
    errTTL              = v2.Err("TTL out of range")
    errMissingHandler   = v2.Err("handler not defined")
    errCorrupted        = v2.Err("corrupted")
)
