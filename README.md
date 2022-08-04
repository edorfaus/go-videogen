VideoGen
========

This package is meant as a base for programs that generate video streams
in real-time, with transparency; e.g. for stream overlays.

The main things it provides are a frameloop that outputs frames at a
consistent framerate, and some ffmpeg-based encoders to turn the raw
video stream into something that more players will accept as input.

This package uses the Go stdlib's image package to handle the frames,
with each frame being an _*image.NRGBA_ instance.

Thus, you can use whatever image drawing library you want to generate
those frames, including none at all, as long as you get such images out.

---

This package also includes a main program to serve as a full example,
which generates a simple animation that can be used for testing.

That program has arguments for the video size, framerate and duration,
and for what kind of output you want: raw RGBA, AVI/RGBA, or WebM/VP9.

The AVI and WebM options require that the _ffmpeg_ program is installed
and available on your PATH; the raw output does not.

**WARNING:** the program outputs the resulting video stream to _stdout_,
so you will probably want to redirect that.
