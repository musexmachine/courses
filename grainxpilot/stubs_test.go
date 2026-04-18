package grainxpilot

import (
	"context"
	"errors"
	"testing"
)

func TestStubImplementationsReturnNotImplemented(t *testing.T) {
	ctx := context.Background()
	ga := StubGrainAdapter{}
	if _, err := ga.ListRecordings(ctx, ListRecordingsRequest{}); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented from ListRecordings, got %v", err)
	}

	n := StubNormalizer{}
	if _, err := n.Normalize(ctx, NormalizeRequest{}); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented from Normalize, got %v", err)
	}

	bw := StubBrowserWorker{}
	if err := bw.Attach(ctx, BrowserAttachRequest{}); !errors.Is(err, ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented from Attach, got %v", err)
	}
}
