package channel

import (
	"context"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"
)

func testSource(ctx context.Context) *StableDescriptor {
	testReader := strings.NewReader("content")
	testResumedReader := strings.NewReader("content")

	sourceDesc := CreateStableDescriptor(ctx, len("content"),
		func() (io.Reader, error) {
			return testReader, nil
		}, func() (io.Reader, error) {
			return testResumedReader, nil
		})

	return sourceDesc
}

// Tests that a transition can be interrupted and resumed
func TestTransientTransitionInterruption(t *testing.T) {
	desc := &TransientDescriptor{}

	desc.SetTransition(func(ctx context.Context, url *url.URL, reader io.ReadCloser) (*StableDescriptor, error) {
		interrupt, err := InterruptedTransition(500, time.Millisecond*100)
		if err != nil {
			t.Error(err)
			t.Fail()
		}

		return nil, interrupt
	})
	desc.SetExpected(1000)
	desc.SetLocation(&url.URL{})

	ctx := context.Background()

	desc.SetSource(testSource(ctx))

	desc, err := desc.ShouldTransition(ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	current, total, progress, err := desc.Position()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if current != 500 {
		t.Errorf("unexpected current offset %d", current)
		t.Fail()
	}

	if total != 1000 {
		t.Error("unexpected total value")
		t.Fail()
	}

	if progress != 0.5 {
		t.Error("unexpected progress value")
		t.Fail()
	}

	desc, _ = desc.SetTransition(func(ctx context.Context, url *url.URL, reader io.ReadCloser) (*StableDescriptor, error) {
		return &StableDescriptor{}, nil
	})

	desc, err = desc.ShouldTransition(ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	current, total, progress, err = desc.Position()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if current != 1000 {
		t.Error("unexpected current offset")
		t.Fail()
	}

	if total != 1000 {
		t.Error("unexpected total value")
		t.Fail()
	}

	if progress != 1.0 {
		t.Error("unexpected progress value")
		t.Fail()
	}
}

func TestTransientChannel(t *testing.T) {
	ctx := context.Background()
	desc := &TransientDescriptor{}

	modified, err := desc.SetExpected(1000)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if desc != modified {
		t.Error("Expected the same pointer")
		t.Fail()
	}

	modified, err = desc.SetLocation(&url.URL{})
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if desc != modified {
		t.Error("Expected the same pointer")
		t.Fail()
	}

	modified, err = desc.SetSource(testSource(ctx))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if desc != modified {
		t.Error("Expected the same pointer")
		t.Fail()
	}

	modified, err = desc.SetTransition(func(ctx context.Context, url *url.URL, reader io.ReadCloser) (*StableDescriptor, error) {
		return testSource(ctx), nil
	})
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if desc != modified {
		t.Error("Expected the same pointer")
		t.Fail()
	}

	desc, err = desc.ShouldTransition(ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = desc.Wait()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	modified, err = desc.SetOffset(1000)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if desc != modified {
		t.Error("Expected the same pointer")
		t.Fail()
	}
}
