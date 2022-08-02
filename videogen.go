package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"
	"os/exec"

	"github.com/edorfaus/go-videogen/anim"
	"github.com/edorfaus/go-videogen/frameloop"
)

var Width, Height, Rate int

var RawOutput bool

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
	flag.BoolVar(&RawOutput, "raw", false, "write raw RGBA output")
	flag.Parse()
}

func run() error {
	args()

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
	if RawOutput {
		return target, nil, nil
	}

	cmd := exec.CommandContext(
		ctx, "ffmpeg",
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
		"-f", "webm", "-pix_fmt", "yuva420p",
		// attempt to avoid buffering latency
		"-avioflags", "direct",
		"-fflags", "flush_packets",
		"-flush_packets", "1",
		// output file: stdout
		"-",
	)

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
