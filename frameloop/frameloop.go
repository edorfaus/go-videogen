package frameloop

import (
	"context"
	"errors"
	"image"
	"io"
	"sync"
	"time"
)

type Frame = image.NRGBA

type FrameLoop struct {
	target io.Writer
	frames <-chan *Frame

	cancel func()
	done   chan struct{}

	sync chan struct{}

	mu  sync.Mutex
	err error
}

func New(
	ctx context.Context, target io.Writer, frames <-chan *Frame,
	frameRate int,
) *FrameLoop {
	c, cancel := context.WithCancel(ctx)
	fl := &FrameLoop{
		target: target,
		frames: frames,
		cancel: cancel,
		sync:   make(chan struct{}, 1),
		done:   make(chan struct{}),
	}
	if frameRate < 1 {
		fl.stop(errors.New("invalid frame rate, must be above 0"))
		close(fl.done)
		return fl
	}
	go fl.frameLoop(c, frameRate)
	return fl
}

func (fl *FrameLoop) frameLoop(ctx context.Context, frameRate int) {
	defer func() {
		fl.mu.Lock()
		close(fl.done)
		fl.mu.Unlock()
	}()

	var frame *Frame
	select {
	case <-ctx.Done():
		return
	case f, ok := <-fl.frames:
		if !ok {
			fl.stop(errors.New("channel closed before first frame"))
			return
		}
		frame = f
	}

	ticker := time.NewTicker(time.Second / time.Duration(frameRate))
	defer ticker.Stop()

	fl.sendFrame(frame)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fl.sendFrame(frame)
		case f, ok := <-fl.frames:
			if !ok {
				// Stop trying to receive from the closed channel
				fl.frames = nil
				break
			}
			frame = f
			// Send the new frame at least once before taking another
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fl.sendFrame(frame)
			}
		}
	}
}

func (fl *FrameLoop) sendFrame(frame *image.NRGBA) {
	b := frame.Bounds()
	w, h := b.Dx()*4, b.Dy()
	i := frame.PixOffset(b.Min.X, b.Min.Y)

	if i == 0 && frame.Stride == w && len(frame.Pix) == w*h {
		// The frame is not a sub-image, Pix holds exactly what we want
		if _, err := fl.target.Write(frame.Pix); err != nil {
			fl.stop(err)
			return
		}
		fl.sendSync()
		return
	}

	// The frame is a sub-image, so Pix contains extra data we'll skip
	for j := 0; j < h; j++ {
		if _, err := fl.target.Write(frame.Pix[i : i+w]); err != nil {
			fl.stop(err)
			return
		}
		i += frame.Stride
	}
	fl.sendSync()
}

func (fl *FrameLoop) sendSync() {
	select {
	case fl.sync <- struct{}{}:
	default:
	}
}

func (fl *FrameLoop) Sync() <-chan struct{} {
	return fl.sync
}

func (fl *FrameLoop) Done() <-chan struct{} {
	return fl.done
}

func (fl *FrameLoop) Stop() {
	fl.stop(nil)
}

func (fl *FrameLoop) Err() error {
	fl.mu.Lock()
	err := fl.err
	fl.mu.Unlock()
	return err
}

func (fl *FrameLoop) stop(err error) {
	fl.mu.Lock()
	if fl.err == nil {
		fl.err = err
	}
	fl.cancel()
	fl.mu.Unlock()
}
