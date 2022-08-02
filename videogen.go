package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"

	"github.com/edorfaus/go-videogen/anim"
	"github.com/edorfaus/go-videogen/encoder"
	"github.com/edorfaus/go-videogen/frameloop"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run() error {
	handleArgs()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Fprintf(os.Stderr, "WxH: %vx%v | frameRate: %v\n", Width, Height, Rate)

	frame := image.NewNRGBA(image.Rect(0, 0, Width, Height))

	frames := make(chan *frameloop.Frame, 1)

	// Set up the animation
	gb, pos, err := setupAnimation()
	if err != nil {
		return err
	}

	// Set up output video stream encoding
	rawVideo, err := setupEncoder(ctx, os.Stdout)
	if err != nil {
		return err
	}

	// Start the frame output loop
	fl := frameloop.New(ctx, rawVideo, frames, Rate)
	defer fl.Stop()

	gb.Draw(frame, pos, draw.Src)
	frames <- frame

	// Run the animation
	for i := 0; i < Rate*10; i++ {
		select {
		case <-fl.Done():
			break
		case <-fl.Sync():
			gb.CycleColor(256 / Rate)
			gb.CycleGradient(gb.Bounds().Dy() / (Rate * 2))

			gb.Draw(frame, pos, draw.Src)
			//frames <- frame
		}
	}

	// Stop the frame loop, wait for it to be done
	fl.Stop()
	<-fl.Done()
	if err := fl.Err(); err != nil {
		return err
	}

	// Shut down the encoder
	if err := rawVideo.Close(); err != nil {
		return err
	}

	return nil
}

func setupEncoder(
	ctx context.Context, target io.WriteCloser,
) (io.WriteCloser, error) {

	switch OutFormat {
	case OutputAvi:
		return encoder.FfmpegAviCopy(ctx, target, Width, Height, Rate)
	case OutputWebM:
		return encoder.FfmpegWebmVP9(ctx, target, Width, Height, Rate)
	case OutputRaw:
		return target, nil
	default:
		return nil, fmt.Errorf("unexpected OutFormat: %v", OutFormat)
	}
}

func setupAnimation() (*anim.GradBox, image.Rectangle, error) {
	pos := image.Rect(Width/4, Height/4, Width, Height)

	col := color.NRGBA{0, 255, 0, 255}
	sz := pos.Min
	switch {
	case sz.X < 4:
		sz.X = Width
		pos.Min.X = 0
	case sz.X < 8:
		sz.X = Width / 2
	}
	switch {
	case sz.Y < 4:
		sz.Y = Height
		pos.Min.Y = 0
	case sz.Y < 8:
		sz.Y = Height / 2
	}

	gb, err := anim.NewGradBox(sz.X, sz.Y, col)

	return gb, pos, err
}
