package main

import "github.com/DataDog/zstd"

func Decompress(buf []byte) ([]byte, error) {
	return zstd.Decompress(nil, buf)
}

func Compress(buf []byte) ([]byte, error) {
	return zstd.Compress(nil, buf)
}
