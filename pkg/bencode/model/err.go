package model

import "errors"

var (
	ErrNum     = errors.New("expect num")
	ErrCol     = errors.New("expect colon")
	ErrEpI     = errors.New("expect char i")
	ErrEpE     = errors.New("expect char e")
	ErrTyp     = errors.New("wrong type")
	ErrIvd     = errors.New("invalid bencode")
	ErrEncode  = errors.New("unexpect error while encoding")
	ErrMarshal = errors.New("marshal dst must be struct or slice ptr")
)
