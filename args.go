package main

import (
	"flag"
	"fmt"
	"time"
)

var Width, Height, Rate int

var OutFormat = OutputAvi

var Duration time.Duration

func handleArgs() (err error) {
	flag.IntVar(&Width, "w", 256, "width of video")
	flag.IntVar(&Height, "h", 256, "height of video")
	flag.IntVar(&Rate, "r", 10, "frame rate of video")
	flag.Var(&OutFormat, "f", "`format` to output: raw, avi, webm")
	flag.DurationVar(
		&Duration, "d", 10*time.Second, "duration of output video",
	)

	flag.Parse()

	pos := func(v int, n string) {
		if v <= 0 && err == nil {
			err = fmt.Errorf("%s must be positive: %v", n, v)
		}
	}
	pos(Width, "width")
	pos(Height, "height")
	pos(Rate, "frame rate")
	if Duration <= 0 && err == nil {
		err = fmt.Errorf("duration must be positive: %v", Duration)
	}

	return err
}

type OutputFormat int

const (
	OutputAvi OutputFormat = iota + 1
	OutputRaw
	OutputWebM
)

func (f OutputFormat) String() string {
	switch f {
	case OutputAvi:
		return "avi"
	case OutputRaw:
		return "raw"
	case OutputWebM:
		return "webm"
	default:
		return fmt.Sprintf("unknown (%d)", int(f))
	}
}

func (f *OutputFormat) Set(v string) error {
	switch v {
	case "r", "raw":
		*f = OutputRaw
	case "a", "avi":
		*f = OutputAvi
	case "w", "webm":
		*f = OutputWebM
	default:
		return fmt.Errorf("unknown output format %q", v)
	}
	return nil
}

func (f *OutputFormat) Get() interface{} {
	return *f
}
