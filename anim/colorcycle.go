package anim

import (
	"image/color"
)

func ColorCycle(c color.NRGBA) color.NRGBA {
	switch {
	case c.R == 255 && c.G < 255 && c.B == 0:
		c.G++
	case c.R > 0 && c.G == 255 && c.B == 0:
		c.R--
	case c.R == 0 && c.G == 255 && c.B < 255:
		c.B++
	case c.R == 0 && c.G > 0 && c.B == 255:
		c.G--
	case c.R < 255 && c.G == 0 && c.B == 255:
		c.R++
	case c.R == 255 && c.G == 0 && c.B > 0:
		c.B--
	// end of regular cycle, below is to start it from other colors
	default:
		if c.G < 255 {
			c.G++
		}
		if c.R < 128 && c.R > 0 {
			c.R--
		}
		if c.B < 128 && c.B > 0 {
			c.B--
		}
		if c.R >= 128 {
			if c.B >= 128 {
				c.R--
				if c.B < 255 {
					c.B++
				}
			} else {
				if c.R < 255 {
					c.R++
				}
				if c.B > 0 {
					c.B--
				}
			}
		} else if c.B >= 128 && c.B < 255 {
			c.B++
		}
	}
	return c
}
