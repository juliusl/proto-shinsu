package channel

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

var ErrIncompleteTransition = errors.New("incomplete transition")
var ErrTemporaryOutage = errors.New("temporary source outage")

func TemporaryOutage(current int64, cooldown time.Duration) (*IncompleteTransition, error) {
	timer := time.NewTimer(cooldown)
	timer.Stop()

	return &IncompleteTransition{
		current:       current,
		cooldown:      cooldown,
		cooldownTimer: timer,
		error:         ErrTemporaryOutage,
	}, nil
}

func InterruptedTransition(current int64, cooldown time.Duration) (*IncompleteTransition, error) {
	if cooldown < MinimumCooldown {
		return nil, fmt.Errorf("cooldown must be greater than %d", MinimumCooldown)
	}

	timer := time.NewTimer(cooldown)
	timer.Stop()

	return &IncompleteTransition{
		current:       current,
		cooldown:      cooldown,
		cooldownTimer: timer,
		error:         ErrIncompleteTransition,
	}, nil
}

type IncompleteTransition struct {
	current       int64
	cooldown      time.Duration
	cooldownTimer *time.Timer
	error
}

func (i *IncompleteTransition) Unwrap() error {
	if i.cooldownTimer != nil {
		i.cooldownTimer.Reset(i.cooldown)
		select {
		case <-i.cooldownTimer.C:
		case <-session.Done():
		}
	}

	return i.error
}

func (i *IncompleteTransition) Current() int64 {
	return i.current
}

var (
	session    context.Context
	cancelFunc context.CancelFunc
)

const MinimumCooldownEnvironmentVariable = "MINIMUM_COOLDOWN"

var MinimumCooldown time.Duration

func init() {
	Reset()
}

func Shutdown() {
	cancelFunc()
}

func Reset() {
	session = context.Background()
	session, cancelFunc = context.WithCancel(session)
	MinimumCooldown = time.Millisecond * 100

	if os.Getenv("MINIMUM_COOLDOWN") != "" {
		dur, err := time.ParseDuration(os.Getenv("MINIMUM_COOLDOWN"))
		if err != nil {
			return
		}

		if dur > 0 {
			MinimumCooldown = dur
		}
	}
}
