// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2016-2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package utils

// #cgo LDFLAGS: -lcrypt
// #define _GNU_SOURCE
// #include <crypt.h>
// #include <stdlib.h>
import "C"
import "unsafe"

// CryptUser wraps C library crypt_r
func CryptUser(key, salt string) string {
	data := C.struct_crypt_data{}
	ckey := C.CString(key)
	csalt := C.CString(salt)
	out := C.GoString(C.crypt_r(ckey, csalt, &data))
	C.free(unsafe.Pointer(ckey))
	C.free(unsafe.Pointer(csalt))
	return out
}
