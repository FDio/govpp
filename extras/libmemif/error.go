// Copyright (c) 2017 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !windows && !darwin

package libmemif

/*
#cgo LDFLAGS: -lmemif

#include <unistd.h>
#include <libmemif.h>
*/
import "C"

// List of errors thrown by go-libmemif.
// Error handling code should compare returned error by value against these variables.
var (
	ErrSyscall       = newMemifError(1)
	ErrConnRefused   = newMemifError(2)
	ErrAccess        = newMemifError(3)
	ErrNoFile        = newMemifError(4)
	ErrFileLimit     = newMemifError(5)
	ErrProcFileLimit = newMemifError(6)
	ErrAlready       = newMemifError(7)
	ErrAgain         = newMemifError(8)
	ErrBadFd         = newMemifError(9)
	ErrNoMem         = newMemifError(10)
	ErrInvalArgs     = newMemifError(11)
	ErrNoConn        = newMemifError(12)
	ErrConn          = newMemifError(13)
	ErrClbFDUpdate   = newMemifError(14)
	ErrFileNotSock   = newMemifError(15)
	ErrNoShmFD       = newMemifError(16)
	ErrCookie        = newMemifError(17)

	// Not thrown, instead properly handled inside the golang adapter:
	ErrNoBufRing    = newMemifError(18)
	ErrNoBuf        = newMemifError(19)
	ErrNoBufDetails = newMemifError(20)

	ErrIntWrite     = newMemifError(21)
	ErrMalformedMsg = newMemifError(22)
	ErrQueueID      = newMemifError(23)
	ErrProto        = newMemifError(24)
	ErrIfID         = newMemifError(25)
	ErrAcceptSlave  = newMemifError(26)
	ErrAlreadyConn  = newMemifError(27)
	ErrMode         = newMemifError(28)
	ErrSecret       = newMemifError(29)
	ErrNoSecret     = newMemifError(30)
	ErrMaxRegion    = newMemifError(31)
	ErrMaxRing      = newMemifError(32)
	ErrNotIntFD     = newMemifError(33)
	ErrDisconnect   = newMemifError(34)
	ErrDisconnected = newMemifError(35)
	ErrUnknownMsg   = newMemifError(36)
	ErrPollCanceled = newMemifError(37)

	// Errors added by the adapter:
	ErrNotInit     = newMemifError(100, "libmemif is not initialized")
	ErrAlreadyInit = newMemifError(101, "libmemif is already initialized")
	ErrUnsupported = newMemifError(102, "the feature is not supported by C-libmemif")

	// Received unrecognized error code from C-libmemif.
	ErrUnknown = newMemifError(-1, "unknown error")
)

// MemifError implements and extends the error interface with the method Code(),
// which returns the integer error code as returned by C-libmemif.
type MemifError struct {
	code        int
	description string
}

// Error prints error description.
func (e *MemifError) Error() string {
	return e.description
}

// Code returns the integer error code as returned by C-libmemif.
func (e *MemifError) Code() int {
	return e.code
}

// A registry of libmemif errors. Used to convert C-libmemif error code into
// the associated MemifError.
var errorRegistry = map[int]*MemifError{}

// newMemifError builds and registers a new MemifError.
func newMemifError(code int, desc ...string) *MemifError {
	var err *MemifError
	if len(desc) > 0 {
		err = &MemifError{code: code, description: "libmemif: " + desc[0]}
	} else {
		err = &MemifError{code: code, description: "libmemif: " + C.GoString(C.memif_strerror(C.int(code)))}
	}
	errorRegistry[code] = err
	return err
}

// getMemifError returns the MemifError associated with the given C-libmemif
// error code.
func getMemifError(code int) error {
	if code == 0 {
		return nil /* success */
	}
	err, known := errorRegistry[code]
	if !known {
		return ErrUnknown
	}
	return err
}
