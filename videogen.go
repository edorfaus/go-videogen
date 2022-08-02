package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"os"
	"os/exec"

	"github.com/edorfaus/go-videogen/anim"
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
	rawVideo, ffmpeg, err := encoder(ctx, os.Stdout)
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

	// Shut down the encoder, if necessary
	if ffmpeg != nil {
		if err := rawVideo.Close(); err != nil {
			return err
		}
		if err := ffmpeg.Wait(); err != nil {
			return err
		}
	}

	return nil
}

func encoder(
	ctx context.Context, target io.WriteCloser,
) (io.WriteCloser, *exec.Cmd, error) {
	if OutFormat == OutputRaw {
		return target, nil, nil
	}

	args := []string{
		"-hide_banner", "-nostdin",

		// input args

		"-an", // no audio
		"-f", "rawvideo", "-pixel_format", "rgba",
		"-framerate", fmt.Sprintf("%v", Rate),
		"-video_size", fmt.Sprintf("%vx%v", Width, Height),

		// attempt to avoid buffering latency
		//"-avioflags", "direct", // this seems to cause missing data
		"-fflags", "nobuffer",
		"-probesize", "32",
		"-analyzeduration", "0",
		"-fpsprobesize", "0",

		// input file: stdin
		"-i", "-",

		// output args

		// attempt to avoid buffering latency
		"-avioflags", "direct",
		"-fflags", "flush_packets",
		"-flush_packets", "1",
	}

	switch OutFormat {
	case OutputAvi:
		// AVI with raw video - for max speed
		args = append(args, "-f", "avi", "-c:v", "copy")
	case OutputWebM:
		// WebM with VP9 - for smaller data, and maybe compatibility
		args = append(args,
			"-f", "webm", "-c:v", "libvpx-vp9", "-pix_fmt", "yuva420p",
			// try to run quickly enough for realtime use
			"-deadline", "realtime", "-quality", "realtime",
		)
	case OutputRaw:
		panic("reached unreachable code (this was checked earlier)")
	default:
		panic(fmt.Sprintf("unexpected OutFormat: %v", OutFormat))
	}

	// output file: stdout
	args = append(args, "-")

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	cmd.Stdout = target
	cmd.Stderr = os.Stderr

	input, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	return input, cmd, nil
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
