package errext

import "errors"

// vendored from https://github.com/sapcc/go-bits/blob/master/errext/errext.go (also licensed Apache 2.0) to prevent go.mod go bump to 1.21

// As is a variant of errors.As() that leverages generics to present a nicer interface.
//
//	//this code:
//	var perr os.PathError
//	if errors.As(err, &perr) {
//		handle(perr)
//	}
//	//can be rewritten as:
//	if perr, ok := errext.As[os.PathError](err); ok {
//		handle(perr)
//	}
//
// This is sometimes more verbose (like in this example), but allows to scope
// the specific error variable to the condition's then-branch, and also looks
// more idiomatic to developers already familiar with type casts.
func As[T error](err error) (T, bool) {
	var result T
	ok := errors.As(err, &result)
	return result, ok
}
