// SPDX-FileCopyrightText: 2018 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

// Package respondwith contains some helper functions for generating responses
// in HTTP handlers. Its name is like that because it pairs up with the function
// names in this package, e.g. "respondwith.ErrorText" or "respondwith.JSON".
package respondwith

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gofrs/uuid/v5"

	"github.com/sapcc/go-bits/logg"
	"github.com/sapcc/go-bits/must"
)

// JSON serializes the given data into an HTTP response body
// The `code` argument specifies the HTTP response code, usually 200.
func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	// NOTE 1: We used to do json.Marshal() here, but switched to json.Encoder.Encode() to avoid
	// allocating an additional []byte buffer with the entire JSON response before writing it out.
	//
	// NOTE 2: Intuition would suggest to wrap `w` in a bufio.Writer, but json.Encoder already writes
	// into an internal buffer first and then sends that entire buffer into w.Write() all at once, so
	// we do not need to add buffering ourselves.

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		logg.Error("could not respondwith.JSON(): " + err.Error())
	}
}

// ErrorText produces an error response with HTTP status code 500 and
// Content-Type text/plain if the given error is non-nil. Otherwise, nothing is
// done and false is returned. Idiomatic usage looks like this:
//
//	value, err := thisMayFail()
//	if respondwith.ErrorText(w, err) {
//		return
//	}
//
//	useValue(value)
func ErrorText(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
	return true
}

// ObfuscatedErrorText produces an obfuscated error response with HTTP status code 500 and
// Content-Type text/plain containing an error id if the given error is non-nil.
// The real error with the respective id is printed into the program log.
// If the error is nil, nothing is done and false is returned.
// Idiomatic usage looks like this:
//
//	value, err := thisMayFail()
//	if respondwith.ObfuscatedErrorText(w, err) {
//		return
//	}
//
//	useValue(value)
func ObfuscatedErrorText(w http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}

	logUUID := must.Return(uuid.NewV4()).String()

	logg.Error("%s is: %s", logUUID, err.Error())
	http.Error(w, fmt.Sprintf("Internal Server Error (ID = %s)", logUUID), http.StatusInternalServerError)
	return true
}
