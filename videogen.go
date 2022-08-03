package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"time"

	"github.com/edorfaus/go-videogen/anim"
	"github.com/edorfaus/go-videogen/encoder"
	"github.com/edorfaus/go-videogen/frameloop"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		if errors.Is(err, frameloop.ErrPanic) {
			// If the panic was rethrown, the runtime might be about to
			// print a stack trace and exit for us; we don't want to
			// stop that by exiting too early, so sleep for a bit.
			time.Sleep(time.Second)
		}
		os.Exit(1)
	}
}

func run() error {
	if err := handleArgs(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Fprintf(
		os.Stderr, "WxH: %vx%v | frameRate: %v | duration: %v\n",
		Width, Height, Rate, Duration,
	)

	frame := image.NewNRGBA(image.Rect(0, 0, Width+4, Height))

	frames := make(chan *frameloop.Frame, 1)
	frameSent := make(chan *frameloop.Frame, 1)

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
	fl := frameloop.New(ctx, rawVideo, Rate, frames, frameSent)
	defer fl.Stop()

	gb.Draw(frame, pos, draw.Src)
	subFrame := frame.SubImage(image.Rect(0, 0, Width, Height))
	frames <- subFrame.(*frameloop.Frame)

	// Run the animation
	frameCount := int64(time.Duration(Rate) * Duration / time.Second)
	for i := int64(1); i < frameCount; i++ {
		if fl.WaitFrame() == nil {
			break
		}

		gb.CycleColor(256 / Rate)
		gb.CycleGradient(gb.Bounds().Dy() / (Rate * 2))

		gb.Draw(frame, pos, draw.Src)

		x := int(i % 8)
		if x > 3 {
			x = 7 - x
		}
		f := frame.SubImage(image.Rect(x, 0, Width+x, Height))
		frames <- f.(*frameloop.Frame)
	}

	// Wait for the last frame to have been sent
	fl.WaitFrame()

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
