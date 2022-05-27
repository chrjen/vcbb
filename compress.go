//go:build !(linux && amd64)

package main

import "github.com/klauspost/compress/zstd"

func Decompress(buf []byte) ([]byte, error) {
	d, err := zstd.NewReader(nil, zstd.WithDecoderMaxMemory(1<<30))
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
