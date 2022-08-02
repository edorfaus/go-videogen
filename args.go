package main

import (
	"flag"
	"fmt"
)

var Width, Height, Rate int

var OutFormat OutputFormat

func handleArgs() {
	flag.IntVar(&Width, "w", 8, "width of video")
	flag.IntVar(&Height, "h", 8, "height of video")
	flag.IntVar(&Rate, "r", 10, "frame rate of video")
	flag.Var(&OutFormat, "f", "`format` to output: raw, avi, webm")
	flag.Parse()
}

type OutputFormat int

const (
	OutputAvi OutputFormat = iota
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
