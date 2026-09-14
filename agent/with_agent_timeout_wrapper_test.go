// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package agent

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestWithAgentTimeoutOnToolContext verifies NewToolContext's wrapper returns a
// usable Context and CancelFunc (the previous implementation returned nil, nil
// and panicking on defer cancel()), and that the timeout fires.
func TestWithAgentTimeoutOnToolContext(t *testing.T) {
	toolCtx := NewToolContext(&ContextMock{}, "call-1", nil, nil)

	ctx, cancel := toolCtx.WithAgentTimeout(20 * time.Millisecond)
	if ctx == nil || cancel == nil {
		t.Fatalf("WithAgentTimeout() = (%v, %v), want a usable Context and a non-nil CancelFunc", ctx, cancel)
	}
	defer cancel()

	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Errorf("ctx.Err() = %v, want context.DeadlineExceeded", ctx.Err())
		}
	case <-time.After(time.Second):
		t.Fatal("WithAgentTimeout context was not canceled within 1s")
	}
}

// TestWithAgentTimeoutOnCallbackContext verifies NewCallbackContext's wrapper
// returns a usable Context and CancelFunc, and that the timeout fires.
func TestWithAgentTimeoutOnCallbackContext(t *testing.T) {
	cbCtx := NewCallbackContext(&ContextMock{}, nil)

	ctx, cancel := cbCtx.WithAgentTimeout(20 * time.Millisecond)
	if ctx == nil || cancel == nil {
		t.Fatalf("WithAgentTimeout() = (%v, %v), want a usable Context and a non-nil CancelFunc", ctx, cancel)
	}
	defer cancel()

	select {
	case <-ctx.Done():
		if !errors.Is(ctx.Err(), context.DeadlineExceeded) {
			t.Errorf("ctx.Err() = %v, want context.DeadlineExceeded", ctx.Err())
		}
	case <-time.After(time.Second):
		t.Fatal("WithAgentTimeout context was not canceled within 1s")
	}
}

// TestWithAgentTimeoutOnWrappers_CancelDoesNotPanic covers the idiomatic
// defer cancel() path that previously nil-pointer panicked when the wrappers
// returned a nil CancelFunc.
func TestWithAgentTimeoutOnWrappers_CancelDoesNotPanic(t *testing.T) {
	for _, tc := range []struct {
		name string
		ctx  Context
	}{
		{"tool", NewToolContext(&ContextMock{}, "call-1", nil, nil)},
		{"callback", NewCallbackContext(&ContextMock{}, nil)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx.WithAgentTimeout(time.Hour)
			if ctx == nil || cancel == nil {
				t.Fatalf("WithAgentTimeout() = (%v, %v), want non-nil", ctx, cancel)
			}
			cancel() // must not panic
			select {
			case <-ctx.Done():
			default:
				t.Fatal("cancel() did not close Done()")
			}
		})
	}
}
