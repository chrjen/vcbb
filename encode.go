package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/atotto/clipboard"
	"github.com/mkideal/cli"
)

type encodeCmdT struct {
	cli.Helper
	Clipboard  bool   `cli:"c,clipboard" usage:"write directly to system's clipboard instead of stdout"`
	OutputName string `cli:"o,output" usage:"write to this file instead of stdout"`
	Layer      uint   `cli:"l,layer" usage:"set blueprint layer: 0=logic, 1=on, 2=off" dft:"0"`
}

var encodeCmd = &cli.Command{
	Name:    "encode",
	Aliases: []string{"e"},
	Desc:    "Takes in a PNG image containing the circuit and outputs a corresponding blueprint.",
	Text:    fmt.Sprintf("example: %s encode [options] <file>", filepath.Base(os.Args[0])),
	Argv:    func() interface{} { return new(encodeCmdT) },

	Fn: func(ctx *cli.Context) error {
		argv := ctx.Argv().(*encodeCmdT)
		args := ctx.Args()

		var outputFile io.WriteCloser = os.Stdout
		var outputReader io.Reader
		var outputString string

		// Open and decode image file.
		if len(ctx.Args()) < 1 {
			ctx.WriteUsage()
			log.Fatalf("Missing file argument\n")
		}

		file, err := os.Open(args[0])
		if err != nil {
			log.Fatalf("Failed to open file (%s): %v\n", args[0], err)
		}

		if ctx.IsSet("-o") {
			outputFile, err = os.Create(argv.OutputName)
			if err != nil {
				log.Fatalf("Failed to create output file (%s): %v\n", argv.OutputName, err)
			}
		}

		var wg sync.WaitGroup
		if argv.Clipboard {
			outputReader, outputFile = io.Pipe()
			wg.Add(1)
			go func() {
				defer wg.Done()
				buf, _ := io.ReadAll(outputReader)
				outputString = string(buf)
			}()
		}

		srcimg, err := png.Decode(file)
		if err != nil {
			log.Fatalf("Failed to decode file (%s): %v\n", args[0], err)
		}
		img := image.NewRGBA(srcimg.Bounds())
		draw.Draw(img, img.Bounds(), srcimg, img.Bounds().Min, draw.Src)

		// Create blueprint data.
		var blueprint Blueprint

		blueprint.Pixels, err = Compress(img.Pix)
		if err != nil {
			log.Fatalf("Failed to compress pixel data: %v\n", err)
		}

		blueprint.Footer.HeightType = 2
		blueprint.Footer.WidthType = 2
		blueprint.Footer.BytesType = 2
		blueprint.Footer.LayerType = 2
		blueprint.Height = int32(img.Rect.Dy())
		blueprint.Width = int32(img.Rect.Dx())
		blueprint.Bytes = int32(len(img.Pix))
		blueprint.Layer = ToLayer(argv.Layer)

		// fmt.Println("Blueprint layer: ", blueprint.Layer)

		if blueprint.Layer == -1 {
			log.Fatalf("Got unknown layer type: %d\n", argv.Layer)
		}

		w := base64.NewEncoder(base64.StdEncoding, outputFile)
		_, err = w.Write(blueprint.Pixels)
		if err != nil {
			log.Fatalf("Failed to write blueprint data: %v\n", err)
		}
		err = binary.Write(w, binary.LittleEndian, blueprint.Footer)
		if err != nil {
			log.Fatalf("Failed to write blueprint data: %v\n", err)
		}
		w.Close()

		fmt.Fprintln(outputFile)
		outputFile.Close()

		if argv.Clipboard {
			wg.Wait()
			err = clipboard.WriteAll(outputString)
			if err != nil {
				log.Fatalf("Failed to write to clipboard: %v\n", err)
			}
		}

		return nil
	},
}
