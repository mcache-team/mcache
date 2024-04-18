package item

import "errors"

var (
	NoDataError      = errors.New("no data")
	PrefixExisted    = errors.New("prefix already existed")
	PrefixNotExisted = errors.New("prefix not existed")
)
