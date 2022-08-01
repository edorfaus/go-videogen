package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/edorfaus/go-videogen/anim"
	"github.com/edorfaus/go-videogen/frameloop"
)

var Width, Height, Rate int

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func args() {
	flag.IntVar(&Width, "w", 8, "width of video")
	flag.IntVar(&Height, "h", 8, "height of video")
	flag.IntVar(&Rate, "r", 10, "frame rate of video")
	flag.Parse()
}

func run() error {
	args()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Fprintf(os.Stderr, "WxH: %vx%v | frameRate: %v\n", Width, Height, Rate)

	frame := image.NewNRGBA(image.Rect(0, 0, Width, Height))

	frames := make(chan *frameloop.Frame, 1)

	fl := frameloop.New(ctx, os.Stdout, frames, Rate)
	defer fl.Stop()

	frames <- frame

	c := color.NRGBA{0, 0, 0, 255}
	for i := 0; i < Rate*10; i++ {
		select {
		case <-fl.Done():
			break
		case <-fl.Sync():
			frame.Set(1, 1, c)
			//frames <- frame
			c = anim.ColorCycle(c)
		}
	}

	fl.Stop()
	<-fl.Done()

	return fl.Err()
}
