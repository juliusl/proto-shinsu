package channel

import (
	"context"
	"errors"
	"io"
	"net/url"
	"sync"
	"time"
)

// TransientDescriptor is a struct that represents data that is in a transient state. This means that state is changing.
// When the location, source, and expected values are set, the transition will automatically proceed, until offset == expected
type TransientDescriptor struct {
	offset     int64
	expected   int64
	location   *url.URL
	source     *StableDescriptor
	transition TransitionFunction
	context    context.Context
	err        error
	sync.RWMutex
}

// SetTransition is a function that sets the current transition function. A transition function should only return a StableDescriptor when the transition has fully completed
func (t *TransientDescriptor) SetTransition(transition TransitionFunction) (*TransientDescriptor, error) {
	if t == nil {
		return nil, errors.New("reference to transient descriptor is nil")
	}

	if transition == nil {
		return nil, errors.New("nil pointer value for `transition` parameter")
	}

	t.Lock()
	defer t.Unlock()

	t.transition = transition
	return t, nil
}

// SetLocation is a function that sets the current location. Setting the location implies that the location is available to this descriptor
func (t *TransientDescriptor) SetLocation(location *url.URL) (*TransientDescriptor, error) {
	if t == nil {
		return nil, errors.New("reference to transient descriptor is nil")
	}

	if location == nil {
		return nil, errors.New("nil pointer value for `location` parameter")
	}

	t.Lock()
	defer t.Unlock()

	t.location = location
	return t, nil
}

// SetSource is a function that sets the current source for this transition. Setting the source implies that the source is stable and available to read from
// If a stable descriptor is no longer stable, this will generate an error
func (t *TransientDescriptor) SetSource(source *StableDescriptor) (*TransientDescriptor, error) {
	if t == nil {
		return nil, errors.New("reference to transient descriptor is nil")
	}

	if source == nil {
		return nil, errors.New("nil pointer value for `source` parameter")
	}

	t.Lock()
	defer t.Unlock()

	err := source.IsStable()
	if err != nil {
		return nil, err
	}

	t.source = source
	return t, nil
}

func (t *TransientDescriptor) SetExpected(expected int64) (*TransientDescriptor, error) {
	if t == nil {
		return nil, errors.New("reference to transient descriptor is nil")
	}

	if expected < 0 {
		return nil, errors.New("negative value for `expected` parameter")
	}

	t.Lock()
	defer t.Unlock()

	t.expected = expected
	return t, nil
}

func (t *TransientDescriptor) SetOffset(offset int64) (*TransientDescriptor, error) {
	if t == nil {
		return nil, errors.New("reference to transient descriptor is nil")
	}

	if offset < 0 {
		return nil, errors.New("negative value for `offset` parameter")
	}

	if offset > t.expected {
		return nil, errors.New("offset parameter cannot be greater than `expected`")
	}

	t.Lock()
	defer t.Unlock()

	t.offset = offset
	return t, nil
}

func (t *TransientDescriptor) ShouldTransition(ctx context.Context) (*TransientDescriptor, error) {
	if t == nil {
		return nil, errors.New("reference to transient descriptor is nil")
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	deadline, ok := ctx.Deadline()
	if ok {
		if time.Now().After(deadline) {
			Shutdown()
		}

		return nil, errors.New("past deadline")
	}

	// TODO -- if we had statistics, we could not try to not begin this transition
	t.RLock()

	if t.transition == nil {
		return nil, errors.New("transition function is not set")
	}

	if t.err != nil && !errors.Is(t.err, ErrInterruptedTransition) {
		return nil, t.err
	}

	if t.context != nil {
		return nil, errors.New("previous transition did not complete")
	}

	if t.expected < 0 {
		return nil, errors.New("expected bytes transitioned is not set")
	}

	if t.offset < 0 {
		return nil, errors.New("offset is negative")
	}

	if t.offset == t.expected {
		return nil, errors.New("transition has already completed")
	}

	if t.source == nil {
		return nil, errors.New("source is not set")
	}

	if t.location == nil {
		return nil, errors.New("location is not set")
	}

	err := t.source.IsStable()
	if err != nil {
		return nil, err
	}

	// This is more of an assertion
	if t.location != nil && t.source != nil && t.expected > 0 && t.offset < t.expected {
		t.RUnlock()
		return t.Transition(ctx)
	}

	t.RUnlock()
	return nil, errors.New("descriptor is in an unknown state")
}

type TransitionFunction func(ctx context.Context, url *url.URL, reader io.Reader) (*StableDescriptor, error)

func (t *TransientDescriptor) Transition(ctx context.Context) (*TransientDescriptor, error) {
	if t == nil {
		return nil, errors.New("reference to transient descriptor is nil")
	}

	if t == nil {
		return nil, errors.New("Wait was called on a nil TransientDescriptor")
	}

	if t.transition == nil {
		return nil, errors.New("transition function is not set")
	}

	t.Lock()
	ctx, cancelFunc := context.WithCancel(ctx)

	go func(ctx context.Context, cancelFunc context.CancelFunc) {
		defer t.Unlock()
		defer cancelFunc()

		content, err := t.source.Open()
		if err != nil {
			t.err = err
			return
		}
		defer content.Close()

		// If transition should only return a stable descriptor if all of the data was transitioned
		// Otherwise it should return ErrIncompleteTransition
		stable, err := t.transition(ctx, t.location, io.NopCloser(content))
		if err != nil {
			t.err = err

			// The writer of transition has two choices
			// Either they return ErrIncompleteTransition directly, which implies that offset should remain at 0
			// Or they return an IncompleTransition structure that records the current offset. If the structure is
			// returned instead, that indicates that some progress was made however the transition is incomplete. This allows for resuming.
			if errors.Is(err, ErrInterruptedTransition) {
				inc, ok := err.(*IncompleteTransition)
				if !ok {
					return
				}
				// This can be resumed, so record the current offset and clear the current context so that
				// ShouldTransition can be called again
				t.offset = inc.Current()
				t.context = nil
			}

			return
		}

		t.source = stable
		t.offset = t.expected
		t.location = nil
		t.context = nil
		t.err = nil

	}(ctx, cancelFunc)

	t.context = ctx

	return t, nil
}

func (t *TransientDescriptor) Wait(timeout time.Duration) error {
	if t == nil {
		return errors.New("reference to transient descriptor is nil")
	}

	if t.context == nil {
		return errors.New("this descriptor hasn't started to transition yet")
	}

	select {
	case <-t.context.Done():
	case <-time.After(timeout):
		return errors.New("waiting for the transition has timed out")
	}

	t.RLock()
	defer t.RUnlock()

	return t.err
}

func (t *TransientDescriptor) Position() (current int64, expected int64, progress float32, err error) {
	if t == nil {
		return 0, 0, 0.0, errors.New("reference to transient descriptor is nil")
	}

	t.RLock()
	defer t.RUnlock()

	return t.offset, t.expected, float32(t.offset) / float32(t.expected), t.err
}

func (t *TransientDescriptor) Source() (*StableDescriptor, error) {
	if t == nil {
		return nil, errors.New("reference to transient descriptor is nil")
	}
	if t.source == nil {
		return nil, errors.New("source is not set")
	}

	return t.source, nil
}
