package channel

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

func TestStableChannel(t *testing.T) {
	ctx := context.Background()

	var testerr error

	testReader := strings.NewReader("normal open")
	testResumedReader := strings.NewReader("normal open")

	desc := CreateStableDescriptor(ctx, len("normal open"),
		func() (io.Reader, error) {
			return testReader, nil
		}, func() (io.Reader, error) {
			return testResumedReader, nil
		})

	desc.AddHealthChecks(func() error {
		return testerr
	})

	reader, err := desc.Open()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	buffer := make([]byte, len("normal open"))
	read, err := reader.Read(buffer)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if read != len(buffer) {
		t.Errorf("unexpected number of bytes read %d", read)
		t.Fail()
	}

	err = desc.Close()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	err = desc.Reset(ctx)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	// Test renentrant with temporary outage
	reader, err = desc.Open()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	testerr, err = TemporaryOutage(0, time.Millisecond*100)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	read, err = reader.Read(buffer)
	if err == nil {
		t.Error("expected an error on read")
		t.Fail()
	}

	if read > 0 {
		t.Error("expected nothing to be read")
		t.Fail()
	}
	desc.Close()

	reader, err = desc.Open()
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	read, err = reader.Read(buffer)
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	if read != len(buffer) {
		t.Error("unexpected number of bytes read")
		t.Fail()
	}

}
