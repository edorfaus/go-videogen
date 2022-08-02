package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
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

	// Set up output video stream encoding
	rawVideo, ffmpeg, err := encoder(ctx, os.Stdout)
	if err != nil {
		return err
	}

	// Start the frame output loop
	fl := frameloop.New(ctx, rawVideo, frames, Rate)
	defer fl.Stop()

	frames <- frame

	// Run the animation
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
		"-avioflags", "direct",
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
