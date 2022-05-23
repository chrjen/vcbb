//go:build !linux

package main

import (
	"bytes"
	"io/ioutil"

	"github.com/klauspost/compress/zstd"
)

func Decompress(buf []byte) ([]byte, error) {

	d, err := zstd.NewReader(bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(d)
}
