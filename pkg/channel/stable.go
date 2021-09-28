package channel

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

func CreateStableDescriptor(ctx context.Context, size int, open func() (io.Reader, error), resume func() (io.Reader, error)) *StableDescriptor {
	desc := &StableDescriptor{
		total:    size,
		deadline: time.Now().Add(MinimumStabilityDuration),
		checks:   make([]func() error, 0),
		open:     open,
		resume:   resume,
		content:  &bufio.Reader{},
		err:      nil,
	}

	deadline, ok := ctx.Deadline()
	if ok {
		desc.deadline = deadline
	}

	return desc
}

type StableDescriptor struct {
	total    int
	deadline time.Time
	checks   []HealthCheck
	open     func() (io.Reader, error)
	resume   func() (io.Reader, error)
	content  *bufio.Reader
	err      error
	io.ReadCloser
	sync.Mutex
}

type HealthCheck = func() error

func (s *StableDescriptor) AddHealthChecks(check ...func() error) (*StableDescriptor, error) {
	if s == nil {
		return nil, errors.New("reference to descriptor is nil")
	}
	s.Lock()
	defer s.Unlock()

	s.checks = append(s.checks, check...)

	return s, nil
}

// Open is a function that once called will return exclusive access to the underlying reader
// If the this is called under the condition of a temporary outatge, then this function will signal
// call resume instead of open, so that a connection can be re-established
func (s *StableDescriptor) Open() (io.ReadCloser, error) {
	if s == nil {
		return nil, errors.New("reference to descriptor is nil")
	}

	if s.err != nil {
		if errors.Is(s.err, ErrTemporaryOutage) && s.resume != nil {
			resumed, err := s.resume()
			if err != nil {
				return nil, err
			}

			s.Lock()
			s.content.Reset(resumed)
			return s, nil
		}
		return nil, s.err
	}

	opened, err := s.open()
	if err != nil {
		return nil, err
	}

	s.Lock()
	s.content.Reset(opened)
	return s, nil
}

func (s *StableDescriptor) Read(p []byte) (int, error) {
	if s == nil {
		return 0, errors.New("pointer to stable descriptor is nil")
	}

	if errors.Is(s.err, ErrTemporaryOutage) {
		curr, ok := s.err.(*IncompleteTransition)
		if ok {
			discarded, err := s.content.Discard(int(curr.current))
			if err != nil {
				s.err = err
				return 0, err
			}

			if discarded != int(curr.current) {
				s.err = errors.New("could not recover position from outage")
				return 0, s.err
			}
			return s.read(p)
		}
	}

	err := s.IsStable()
	if err != nil {
		s.err = err
		return 0, err
	}

	return s.read(p)
}

func (s *StableDescriptor) read(p []byte) (int, error) {
	read, err := s.content.Read(p)
	if err != nil {
		s.err = err
		return read, err
	}

	return read, nil
}

func (s *StableDescriptor) Close() error {
	if s == nil {
		return errors.New("reference to descriptor is nil")
	}
	s.Unlock()

	return nil
}

func (s *StableDescriptor) Reset(ctx context.Context) error {
	if s == nil {
		return errors.New("reference to descriptor is nil")
	}

	s.Lock()
	defer s.Unlock()

	s.err = nil
	s.deadline = time.Now().Add(MinimumStabilityDuration)

	deadline, ok := ctx.Deadline()
	if ok {
		s.deadline = deadline
	}

	return nil
}

func (s *StableDescriptor) IsStable() error {
	if time.Now().After(s.deadline) {
		return errors.New("descriptor must be reset or recreated")
	}

	var err error
	for _, c := range s.checks {
		err = c()
		if err != nil {
			return err
		}
	}

	return nil
}

const MinimumStabilityDurationEnvironmentVariable = "MINIMUM_STABILITY"

var MinimumStabilityDuration time.Duration

func init() {
	MinimumStabilityDuration = time.Hour * 48

	if os.Getenv("MINIMUM_STABILITY") != "" {
		dur, err := time.ParseDuration(os.Getenv("MINIMUM_STABILITY"))
		if err != nil {
			return
		}

		if dur > 0 {
			MinimumStabilityDuration = dur
		}
	}
}
