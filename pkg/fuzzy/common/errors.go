package common

import "errors"

var (
	ErrNilMessages         error = errors.New("nil messages")
	ErrNoMessages          error = errors.New("no messages")
	ErrTargetLangUndefined error = errors.New("target language undefined")
)
