package main

import (
	"github.com/DataDog/zstd"
)

func Decompress(buf []byte) ([]byte, error) {
	d, err := zstd.Decompress(nil, buf)
	return d, err
}

func Compress(buf []byte) ([]byte, error) {
	d, err := zstd.Compress(nil, buf)
	return d, err
}
