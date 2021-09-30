package channel

import (
	"errors"
	"testing"
)

func TestInterruptedTransition(t *testing.T) {
	_, err := InterruptedTransition(404, MinimumCooldown-10)
	if err == nil {
		t.Error("expected an error when the duration is less than the minimum cooldown")
	}

	interruption, err := InterruptedTransition(404, MinimumCooldown)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if interruption.cooldownTimer.Stop() {
		t.Error("cooldown timer should've been stopped on creation")
		t.Fail()
	}

	if interruption.cooldown != MinimumCooldown {
		t.Error("unexpected cooldown value")
		t.Fail()
	}

	if !errors.Is(interruption, ErrInterruptedTransition) {
		t.Error("expected the interruption to be an ErrIncompleteTransition")
		t.Fail()
	}

	if interruption.Current() != 404 {
		t.Error("expected interruption value to be 404")
		t.Fail()
	}

	if interruption.cooldownTimer.Stop() {
		t.Error("expected cooldown timer to be stopped after Is was called")
		t.Fail()
	}
}
