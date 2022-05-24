//go:build !linux

package main

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/klauspost/compress/zstd"
)

func Decompress(buf []byte) ([]byte, error) {
	d, err := zstd.NewReader(nil)
	if err != nil {
		return nil, err
	}

	return d.DecodeAll(buf, nil)
}

func Compress(buf []byte) ([]byte, error) {
	e, err := zstd.NewWriter(nil)
	if err != nil {
		return nil, err
	}

	return e.EncodeAll(buf, nil), nil
}
