package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestFromPassesThroughAppErrors(t *testing.T) {
	original := New(CodeMediaCorrupt, nil)
	if got := From(original); got != original {
		t.Fatalf("From returned a different error: %v", got)
	}
}

func TestFromFindsWrappedAppErrors(t *testing.T) {
	original := New(CodeModelInUse, nil)
	wrapped := fmt.Errorf("while deleting: %w", original)
	if got := From(wrapped); got != original {
		t.Fatalf("From did not unwrap to the original, got %v", got)
	}
}

func TestFromHidesUnknownErrors(t *testing.T) {
	cause := errors.New("exit status 137")
	got := From(cause)
	if got.Code != CodeInternal {
		t.Errorf("code = %q, want %q", got.Code, CodeInternal)
	}
	if got.Message != Message(CodeInternal) {
		t.Errorf("message leaked the cause: %q", got.Message)
	}
	if !errors.Is(got, cause) {
		t.Error("cause was dropped, so it cannot be logged")
	}
}

func TestFromNil(t *testing.T) {
	if From(nil) != nil {
		t.Error("From(nil) should be nil")
	}
}

func TestEveryCodeHasAMessage(t *testing.T) {
	codes := []string{
		CodeInternal, CodeMediaNoAudioStream, CodeMediaCorrupt, CodeMediaDRMProtected,
		CodeMediaUnsupported, CodeDiskInsufficientSpace, CodeModelNotInstalled,
		CodeModelChecksumMismatch, CodeModelInUse, CodeTranscribeOutOfMemory,
		CodeTranscribeFailed, CodeNetworkUnavailable,
	}
	for _, c := range codes {
		if _, ok := messages[c]; !ok {
			t.Errorf("code %s has no user message", c)
		}
	}
	if len(messages) != len(codes) {
		t.Errorf("messages has %d entries, codes has %d", len(messages), len(codes))
	}
}

func TestWithfKeepsCodeAndFormatsMessage(t *testing.T) {
	e := Withf(CodeDiskInsufficientSpace, nil, "Needs about %dMB free.", 350)
	if e.Code != CodeDiskInsufficientSpace {
		t.Errorf("code = %q", e.Code)
	}
	if e.Message != "Needs about 350MB free." {
		t.Errorf("message = %q", e.Message)
	}
}

func TestIs(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", New(CodeMediaUnsupported, nil))
	if !Is(err, CodeMediaUnsupported) {
		t.Error("Is should see through wrapping")
	}
	if Is(err, CodeMediaCorrupt) {
		t.Error("Is matched the wrong code")
	}
}
