package anim_test

import (
	"image/color"
	"testing"

	"github.com/edorfaus/go-videogen/anim"
)

func TestColorCycle(t *testing.T) {
	// slices instead of maps for performance reasons
	inLoop := make([]bool, 256*256*256)
	seen := make([]bool, 256*256*256)

	// Check that the main color cycle works as expected
	start := color.NRGBA{255, 0, 0, 255}
	c := start
	count := 0
	for {
		k := int(c.R)*256*256 + int(c.G)*256 + int(c.B)
		if inLoop[k] {
			break
		}
		inLoop[k] = true
		seen[k] = true
		count++

		c2 := anim.ColorCycle(c)
		if c2 == c {
			t.Errorf("got same color back for %v", c)
			return
		}
		c = c2
	}
	if c != start {
		t.Errorf("cycle did not end at the starting color")
		return
	}
	t.Logf("Base cycle is %v colors long", count)

	// Check that all the other colors also end up inside the cycle
	var find func(color.NRGBA) bool
	find = func(c color.NRGBA) bool {
		k := int(c.R)*256*256 + int(c.G)*256 + int(c.B)
		if inLoop[k] {
			return true
		}
		if seen[k] {
			t.Errorf("found secondary cycle at %v", c)
			return false
		}
		seen[k] = true

		c2 := anim.ColorCycle(c)
		if c2 == c {
			t.Errorf("got same color back for %v", c)
			return false
		}

		if find(c2) {
			inLoop[k] = true
			return true
		}
		return false
	}

	for r := 0; r <= 255; r++ {
		c.R = uint8(r)
		for g := 0; g <= 255; g++ {
			c.G = uint8(g)
			for b := 0; b <= 255; b++ {
				c.B = uint8(b)
				if !find(c) {
					t.Errorf("color does not go into cycle: %v", c)
					return
				}
			}
		}
	}
}
