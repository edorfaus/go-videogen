package encoder

import (
	"context"
	"fmt"
	"io"
)

// AVI with raw video - for max speed
func FfmpegAviCopy(
	ctx context.Context, target io.Writer, w, h, fr int,
) (io.WriteCloser, error) {

	return FfmpegOutArgs(
		ctx, target, w, h, fr, "-f", "avi", "-c:v", "copy",
	)
}

// WebM with VP9 - for smaller data, and maybe compatibility
func FfmpegWebmVP9(
	ctx context.Context, target io.Writer, w, h, fr int,
) (io.WriteCloser, error) {

	return FfmpegOutArgs(
		ctx, target, w, h, fr,
		"-f", "webm", "-c:v", "libvpx-vp9", "-pix_fmt", "yuva420p",
		// try to run quickly enough for realtime use
		"-deadline", "realtime", "-quality", "realtime",
	)
}

func FfmpegOutArgs(
	ctx context.Context, target io.Writer, w, h, fr int,
	outArgs ...string,
) (io.WriteCloser, error) {

	if ctx == nil {
		return nil, fmt.Errorf("the context cannot be nil")
	}
	if w < 1 || h < 1 {
		return nil, fmt.Errorf("invalid frame size: %vx%v", w, h)
	}
	if fr < 1 {
		return nil, fmt.Errorf("invalid frame rate: %v", fr)
	}

	args := []string{
		"-hide_banner", "-nostdin",

		// input args

		"-an", // no audio
		"-f", "rawvideo", "-pixel_format", "rgba",
		"-framerate", fmt.Sprintf("%v", fr),
		"-video_size", fmt.Sprintf("%vx%v", w, h),

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

	// format and encoding specifications, and anything else provided
	args = append(args, outArgs...)

	// output file: stdout
	args = append(args, "-")

	return PipeCommand(ctx, target, "ffmpeg", args...)
}
