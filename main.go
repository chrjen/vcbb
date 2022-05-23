package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/nfnt/resize"
)

// Blueprint format.
type Blueprint struct {
	Pixels []byte
	Footer
}

type Footer struct {
	_      int32 // Always 2s, probably data type with int32=2.
	Height int32
	_      int32
	Width  int32
	_      int32
	Bytes  int32
	_      int32
	Layer  int32
}

const (
	LayerLogic = 65536
	LayerON    = 131072
	LayerOFF   = 262144
)

const defaultOutputName = "blueprint.png"

func main() {
	log.SetFlags(log.Lshortfile)

	fScale := flag.Uint("s", 1, "Scale of the output image. (integer)")
	fReverse := flag.Bool("r", false, "Takes in an image and outputs a blueprint.")
	fClipboard := flag.Bool("c", false, "Read directly from system's clipboard instead of stdin or file.")
	fOutputName := flag.String("o", "", "Name of the output file. Default to same name as input or 'blueprint.png' if from stdin or clipboard.")

	flag.Parse()

	var inputFile io.Reader = os.Stdin

	if flag.Arg(0) != "" && flag.Arg(0) != "-" {
		var err error
		inputFile, err = os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalf("Failed to read file (%s): %v\n", flag.Arg(0), err)
		}
	}

	if *fClipboard {
		s, err := clipboard.ReadAll()
		if err != nil {
			log.Fatalf("Failed to read clipboard: %v\n", err)
		}
		inputFile = strings.NewReader(s)
	}

	if *fOutputName == "" && flag.Arg(0) != "" {
		*fOutputName = strings.TrimSuffix(filepath.Base(flag.Arg(0)), filepath.Ext(flag.Arg(0))) + ".png"
	} else if *fOutputName == "" {
		*fOutputName = defaultOutputName
	}

	var blueprint Blueprint

	if !*fReverse {
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
			uint(blueprint.Width)**fScale,
			uint(blueprint.Height)**fScale,
			img,
			resize.NearestNeighbor,
		).(*image.RGBA)

		f, err := os.Create(*fOutputName)
		if err != nil {
			log.Fatalf("Failed to create image file: %v", err)
		}

		err = png.Encode(f, img)
		if err != nil {
			log.Fatalf("Failed to encode image file: %v", err)
		}

		fmt.Println("Created", *fOutputName)
	} else {
		// e := zstd.NewWriter(os.Stdout)
		// defer e.Close()

		// _, err := io.Copy(e, os.Stdin)
		// if err != nil {
		// 	log.Fatalf("Failed zstd compression: %v", err)
		// }
		fmt.Println("OOPS, not implemented yet")
	}
}
