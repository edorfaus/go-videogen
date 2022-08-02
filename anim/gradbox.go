package anim

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

type GradBox struct {
	Src  *image.Uniform
	Mask *image.Alpha

	Offset int

	gradient *image.Alpha
}

func NewGradBox(w, h int, c color.Color) (*GradBox, error) {
	if w < 1 || h < 1 {
		return nil, fmt.Errorf("NewGradBox: invalid size: %vx%v", w, h)
	}

	b := &GradBox{
		Src:  image.NewUniform(c),
		Mask: image.NewAlpha(image.Rect(0, 0, w, h)),
	}
	b.initMask()

	return b, nil
}

func (b *GradBox) SetColor(c color.Color) {
	b.Src.C = c
}

func (b *GradBox) CycleColor(steps int) {
	c := color.NRGBAModel.Convert(b.Src.C).(color.NRGBA)
	for i := 0; i < steps; i++ {
		c = ColorCycle(c)
	}
	b.Src.C = c
}

func (b *GradBox) CycleGradient(steps int) {
	b.Offset = (b.Offset + steps) % b.gradient.Bounds().Dy()
	// TODO: do this more efficiently, by copying an image rect first
	b.drawGradient(b.gradient.Rect.Min.Y, b.gradient.Rect.Max.Y)
}

func (b *GradBox) Draw(dst draw.Image, r image.Rectangle, op draw.Op) {
	draw.DrawMask(dst, r, b.Src, image.Point{}, b.Mask, image.Point{}, op)
}

func (b *GradBox) Bounds() image.Rectangle {
	return b.Mask.Bounds()
}

func (b *GradBox) initMask() {
	r := b.Mask.Bounds()

	if r.Dx() < 10 || r.Dy() < 10 {
		b.gradient = b.Mask
		b.drawGradient(0, r.Dy())
		return
	}

	bw := r.Dx()
	if bw > r.Dy() {
		bw = r.Dy()
	}
	bw /= 10
	if bw > 5 {
		bw = 5
	}

	b.gradient = b.Mask.SubImage(image.Rect(
		r.Min.X+bw, r.Min.Y+bw,
		r.Max.X-bw, r.Max.Y-bw,
	)).(*image.Alpha)

	zp := image.Point{}

	halfOpaque := image.NewUniform(color.Alpha{128})

	// top edge
	edge := r
	edge.Max.Y = r.Min.Y + bw
	draw.Draw(b.Mask, edge, image.Opaque, zp, draw.Src)
	if bw > 2 {
		edge.Max.Y = r.Min.Y + 1
		draw.Draw(b.Mask, edge, halfOpaque, zp, draw.Src)
	}

	// bottom edge
	edge.Min.Y = r.Max.Y - bw
	edge.Max.Y = r.Max.Y
	draw.Draw(b.Mask, edge, image.Opaque, zp, draw.Src)
	if bw > 2 {
		edge.Min.Y = r.Max.Y - 1
		draw.Draw(b.Mask, edge, halfOpaque, zp, draw.Src)
	}

	// left edge
	edge.Min.Y = r.Min.Y + bw
	edge.Max.Y = r.Max.Y - bw

	edge.Max.X = r.Min.X + bw
	draw.Draw(b.Mask, edge, image.Opaque, zp, draw.Src)

	// right edge
	edge.Min.X = r.Max.X - bw
	edge.Max.X = r.Max.X
	draw.Draw(b.Mask, edge, image.Opaque, zp, draw.Src)

	if bw > 2 {
		edge.Min.Y = r.Min.Y + 1
		edge.Max.Y = r.Max.Y - 1

		edge.Max.X = r.Min.X + 1
		draw.Draw(b.Mask, edge, halfOpaque, zp, draw.Src)

		edge.Min.X = r.Max.X - 1
		draw.Draw(b.Mask, edge, halfOpaque, zp, draw.Src)
	}

	b.drawGradient(b.gradient.Rect.Min.Y, b.gradient.Rect.Max.Y)
}

func (b *GradBox) drawGradient(from, to int) {
	r := b.gradient.Bounds()
	src := image.NewUniform(color.Alpha{128})

	height := r.Dy()
	yi := from - r.Min.Y
	if b.Offset < 0 {
		yi += (-b.Offset) % height
	} else {
		yi += height - (b.Offset % height)
	}
	pos := r
	for y := from; y < to; y++ {
		yi = yi % height
		src.C = color.Alpha{uint8(256 * yi / height)}
		yi++
		pos.Min.Y = y
		pos.Max.Y = y + 1
		draw.Draw(b.Mask, pos, src, image.Point{}, draw.Src)
	}
}
