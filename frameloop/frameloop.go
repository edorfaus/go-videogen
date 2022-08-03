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
	sent   chan *Frame

	cancel func()
	done   chan struct{}

	mu  sync.Mutex
	err error
}

// New creates a new frame loop.
//
// The frames channel is used to get the frames to be written to the
// target writer. The given frameRate will be maintained by writing the
// last received frame again if a new frame was not received in time.
//
// If the sent channel is not nil, then every written frame will be sent
// to it right after it was written. If the same frame was written again
// due to not receiving a new one, then that frame will be sent again.
// However, a full sent channel will lose frames (non-blocking sends).
//
// Every received frame will be written at least once before a new frame
// is accepted, unless the frame loop ends first.
//
// The loop will start when the first frame is received, and will end
// after ctx is done, or Stop() is called, or a write error occurs.
func New(
	ctx context.Context, target io.Writer, frameRate int,
	frames <-chan *Frame, sent chan *Frame,
) *FrameLoop {
	c, cancel := context.WithCancel(ctx)
	fl := &FrameLoop{
		target: target,
		frames: frames,
		sent:   sent,
		cancel: cancel,
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
		defer fl.mu.Unlock()
		close(fl.done)
	}()
	defer func() {
		if v := recover(); v != nil {
			fl.mu.Lock()
			if fl.err == nil {
				fl.err = PanicErr{"frame loop", v}
			}
			fl.mu.Unlock()
			panic(v)
		}
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

func (fl *FrameLoop) sendFrame(frame *Frame) {
	b := frame.Bounds()
	w, h := b.Dx()*4, b.Dy()
	i := frame.PixOffset(b.Min.X, b.Min.Y)

	// The frame might be a sub-image, in which case Pix contains extra
	// data we have to skip, thus looping one write per line. However,
	// if the frame is not a sub-image, then Pix contains exactly the
	// data we want to send, so we can send everything in one write.
	if i == 0 && frame.Stride == w && len(frame.Pix) == w*h {
		w *= h
		h = 1
	}

	for j := 0; j < h; j++ {
		if _, err := fl.target.Write(frame.Pix[i : i+w]); err != nil {
			fl.stop(err)
			return
		}
		i += frame.Stride
	}

	select {
	case fl.sent <- frame:
	default:
	}
}

// WaitFrame waits for a frame to be returned by the frame loop, and
// returns that frame. This returns nil if the frame loop exits before
// another frame is returned, or if the returned frame was nil.
func (fl *FrameLoop) WaitFrame() *Frame {
	select {
	case <-fl.done:
		return nil
	case frame := <-fl.sent:
		return frame
	}
}

// Done returns a channel that will be closed when the frame loop has
// ended and will not write any more frames.
func (fl *FrameLoop) Done() <-chan struct{} {
	return fl.done
}

// Stop will cause the frame loop to end, and not write out any more
// frames.
//
// Calling Stop may cause a received frame to not be written out, if the
// time to write it out has not yet arrived. However, a frame that is
// already being written will be completed before the loop is ended, to
// avoid writing partial frames.
func (fl *FrameLoop) Stop() {
	fl.stop(nil)
}

// Err will return any error that occurred in the frame loop. It will be
// valid once the frame loop has ended, and the Stop channel is closed.
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
