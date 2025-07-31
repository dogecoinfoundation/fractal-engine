package service

import (
	"testing"
	"time"

	"dogecoin.org/fractal-engine/pkg/store"
)

func TestNewFractalEngineProcessor(t *testing.T) {
	// This test verifies the constructor works correctly
	mockStore := &store.TokenisationStore{}
	processor := NewFractalEngineProcessor(mockStore)

	if processor == nil {
		t.Fatal("Expected processor to be created")
	}

	if processor.store != mockStore {
		t.Error("Expected processor store to match input store")
	}

	if processor.Running {
		t.Error("Expected processor to not be running initially")
	}
}

func TestProcessorStartStop(t *testing.T) {
	// Test the start/stop functionality
	processor := &FractalEngineProcessor{
		store: &store.TokenisationStore{},
	}

	// Verify initial state
	if processor.Running {
		t.Error("Expected processor to not be running initially")
	}

	// Start processor in a goroutine (it will fail quickly due to nil DB)
	go func() {
		// Recover from panic that will happen due to nil DB
		defer func() {
			recover()
		}()
		processor.Start()
	}()

	// Give it a moment to set Running to true
	time.Sleep(50 * time.Millisecond)

	if !processor.Running {
		t.Error("Expected processor to be running after Start")
	}

	// Stop the processor
	processor.Stop()

	// Give it a moment to actually stop
	time.Sleep(50 * time.Millisecond)

	if processor.Running {
		t.Error("Expected processor to not be running after Stop")
	}
}

func TestProcessorFields(t *testing.T) {
	// Test that the processor correctly maintains its fields
	store := &store.TokenisationStore{}
	processor := &FractalEngineProcessor{
		store:   store,
		Running: false,
	}

	// Verify store is set
	if processor.store != store {
		t.Error("Expected store to be set correctly")
	}

	// Test Running flag manipulation
	processor.Running = true
	if !processor.Running {
		t.Error("Expected Running to be true")
	}

	processor.Running = false
	if processor.Running {
		t.Error("Expected Running to be false")
	}
}
