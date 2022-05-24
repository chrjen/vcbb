package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mkideal/cli"
)

// Blueprint format.
type Blueprint struct {
	Pixels []byte
	Footer
}

type Footer struct {
	HeightType int32 // Always 2s, probably data type with int32=2.
	Height     int32
	WidthType  int32
	Width      int32
	BytesType  int32
	Bytes      int32
	LayerType  int32
	Layer      int32
}

const (
	LayerLogic = int32(65536)
	LayerON    = int32(131072)
	LayerOFF   = int32(262144)
)

const defaultOutputName = "blueprint.png"

var rootCmd = &cli.Command{
	Desc: "Virtual Circuit Board blueprint tool by chrjen",
	Text: fmt.Sprintf(`example: %s <command>`, filepath.Base(os.Args[0])),
	Fn: func(ctx *cli.Context) error {
		ctx.WriteUsage()
		return nil
	},
}

func ToLayer(l uint) int32 {
	switch l {
	case 0:
		return LayerLogic
	case 1:
		return LayerON
	case 2:
		return LayerOFF
	default:
		return -1
	}
}

func main() {
	log.SetFlags(log.Lshortfile)

	if err := cli.Root(rootCmd,
		cli.Tree(decodeCmd),
		cli.Tree(encodeCmd),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
