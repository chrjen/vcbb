package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mkideal/cli"
	"github.com/nfnt/resize"
)

type decodeCmdT struct {
	cli.Helper
	Scale        uint   `cli:"s,scale" usage:"scale of the output image" dft:"1"`
	Clipboard    bool   `cli:"c,clipboard" usage:"read directly from system's clipboard instead of stdin or file"`
	OutputName   string `cli:"o,output" usage:"name of the output file" dft:"blueprint.png"`
	OutputStdout bool   `cli:"stdout" usage:"write to stdout instead of to file"`
}

var decodeCmd = &cli.Command{
	Name:    "decode",
	Aliases: []string{"d"},
	Desc:    "Takes in a blueprint and outputs a PNG image with the content of the blueprint.",
	Text:    fmt.Sprintf(`example: %s decode [options] [file]`, filepath.Base(os.Args[0])),
	Argv:    func() interface{} { return new(decodeCmdT) },

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*decodeCmdT)
		args := ctx.Args()

		var blueprint Blueprint
		var inputFile io.Reader = os.Stdin

		if len(args) > 0 && args[0] != "-" {
			var err error
			inputFile, err = os.Open(args[0])
			if err != nil {
				log.Fatalf("Failed to read file (%s): %v\n", args[0], err)
			}
		}

		if argv.Clipboard {
			s, err := clipboard.ReadAll()
			if err != nil {
				log.Fatalf("Failed to read clipboard: %v\n", err)
			}
			inputFile = strings.NewReader(s)
		}

		if ctx.IsSet("--clipboard") && len(args) > 0 {
			argv.OutputName = strings.TrimSuffix(filepath.Base(args[0]), filepath.Ext(args[0])) + ".png"
		} else if argv.OutputName == "" {
			argv.OutputName = defaultOutputName
		}

		// Reads base64 encoded string from stdin.
		buf, err := io.ReadAll(base64.NewDecoder(base64.StdEncoding, inputFile))
		if err != nil {
			log.Fatalf("Failed to read input: %v\n", err)
		}

		// Read the footer.
		FooterSize := binary.Size(blueprint.Footer)
		binary.Read(bytes.NewReader(buf[len(buf)-FooterSize:]), binary.LittleEndian, &blueprint.Footer)

		// Decompress zstd-compressed pixel data.
		d, err := Decompress(buf[:len(buf)-FooterSize])
		if err != nil {
			log.Fatalf("Failed zstd decompression: %v", err)
		}
		blueprint.Pixels = d

		// Create new image with pixel data and write to file.
		img := image.NewRGBA(image.Rectangle{
			image.Point{0, 0},
			image.Point{int(blueprint.Width), int(blueprint.Height)},
		})
		img.Pix = blueprint.Pixels

		img = resize.Resize(
			uint(blueprint.Width)*argv.Scale,
			uint(blueprint.Height)*argv.Scale,
			img,
			resize.NearestNeighbor,
		).(*image.RGBA)

		f, err := os.Create(argv.OutputName)
		if err != nil {
			log.Fatalf("Failed to create image file: %v", err)
		}

		if argv.OutputStdout {
			err = png.Encode(os.Stdout, img)
			if err != nil {
				log.Fatalf("Failed to encode image file: %v", err)
			}
		} else {
			err = png.Encode(f, img)
			if err != nil {
				log.Fatalf("Failed to encode image file: %v", err)
			}
			fmt.Println("Wrote to file", argv.OutputName)
		}

		return nil
	},
}
