#!/bin/bash

W=8
H=8
RATE=10
IN=video.raw
OUT=-

INARG=()
OUTARG=()

for arg ; do
	if [[ "$arg" =~ ^([0-9]+)x([0-9]+)$ ]]; then
		W=${BASH_REMATCH[1]}
		H=${BASH_REMATCH[2]}
	elif [[ "$arg" =~ ^[0-9]+$ ]]; then
		RATE=$arg
	elif [[ "$arg" =~ ^(W|H|RATE|IN|OUT)=(.*)$ ]]; then
		val=${BASH_REMATCH[2]}
		eval "${BASH_REMATCH[1]}=\$val"
	elif [[ "$arg" =~ ^(INARG|OUTARG)"+="(.*)$ ]]; then
		val=${BASH_REMATCH[2]}
		eval "${BASH_REMATCH[1]}+=(\"\$val\")"
	else
		printf >&2 "Error: Unknown argument: %s\n" "$arg"
		exit 1
	fi
done

#ffmpeg -hide_banner -nostdin -an \
#	-f rawvideo -pix_fmt rgba -s ${W}x$H -r $RATE -i "$IN" \
#	-f webm -pix_fmt yuva420p "$OUT"

ffmpeg -hide_banner -nostdin -an \
	-f rawvideo -framerate $RATE -pixel_format rgba \
		-video_size ${W}x$H "${INARG[@]}" -i "$IN" \
	-f webm -pix_fmt yuva420p "${OUTARG[@]}" "$OUT"
