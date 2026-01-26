package behaviortree

import (
	"slices"
	"testing"
)

// SimpleValuer is a minimal implementation of the Valuer interface.
type SimpleValuer struct {
	values map[any]any
}

func (v *SimpleValuer) Value(key any) any {
	if v.values == nil {
		return nil
	}
	return v.values[key]
}

// SimpleValueAttachable is a minimal implementation of the ValueAttachable interface.
// It returns a *SimpleValuer (pointer) to simulate an immutable-like flow where the new object
// is returned, but for simplicity here we just use the same underlying map or copy it.
// To properly test the "With*" semantics which return a new T, we'll make it return a new *SimpleValuer.
type SimpleValueAttachable struct {
	*SimpleValuer
}

// WithValue implements ValueAttachable[*SimpleValuer].
func (s SimpleValueAttachable) WithValue(key, value any) *SimpleValuer {
	newValues := make(map[any]any)
	if s.SimpleValuer != nil && s.SimpleValuer.values != nil {
		for k, v := range s.SimpleValuer.values {
			newValues[k] = v
		}
	}
	newValues[key] = value
	return &SimpleValuer{values: newValues}
}

func TestInterop_Frame(t *testing.T) {
	// 1. Test GetFrame with a Valuer that has no frame
	v := &SimpleValuer{}
	if f := GetFrame(v); f != nil {
		t.Errorf("expected nil frame, got %v", f)
	}

	// 2. Test WithFrame -> GetFrame
	frame := &Frame{Function: "test_func", File: "test_file.go", Line: 123}
	v2 := WithFrame[*SimpleValuer](SimpleValueAttachable{v}, frame)

	f2 := GetFrame(v2)
	if f2 == nil {
		t.Fatal("expected frame, got nil")
	}
	if *f2 != *frame {
		t.Errorf("expected frame %v, got %v", frame, f2)
	}

	// 3. Test WithFrame(nil) -> GetFrame (should be nil)
	v3 := WithFrame[*SimpleValuer](SimpleValueAttachable{v2}, nil)
	if f3 := GetFrame(v3); f3 != nil {
		t.Errorf("expected nil frame after clearing, got %v", f3)
	}
}

func TestInterop_Name(t *testing.T) {
	// 1. Test GetName with empty Valuer
	v := &SimpleValuer{}
	if n := GetName(v); n != "" {
		t.Errorf("expected empty name, got %q", n)
	}

	// 2. Test WithName -> GetName
	name := "test_node"
	v2 := WithName[*SimpleValuer](SimpleValueAttachable{v}, name)
	if n := GetName(v2); n != name {
		t.Errorf("expected name %q, got %q", name, n)
	}
}

func TestInterop_Structure(t *testing.T) {
	// 1. Test GetStructure with empty Valuer
	v := &SimpleValuer{}
	if s := GetStructure(v); s != nil {
		t.Errorf("expected nil structure, got %v", s)
	}

	// 2. Test WithStructure -> GetStructure
	// Create a dummy metadata child

	// Note: SimpleValuer needs to implement Metadata to be used in structure?
	// The Metadata interface requires Value and Children.
	// Our SimpleValuer only implements Valuer.
	// Let's verify: type Metadata interface { Value(key any) any; Children(yield func(Metadata) bool) }
	// So we need a SimpleMetadata struct or make SimpleValuer implement it.

	mdStart := &SimpleMetadata{SimpleValuer: &SimpleValuer{}}

	children := []Metadata{mdStart}
	seq := slices.Values(children)

	// We need a helper to wrap SimpleValuer as ValueAttachable that returns *SimpleMetadata?
	// actually WithStructure[T] returns T.
	// If we use SimpleValueAttachable it returns *SimpleValuer.
	// But GetStructure returns iter.Seq[Metadata].

	v2 := WithStructure[*SimpleValuer](SimpleValueAttachable{v}, seq)

	s2 := GetStructure(v2)
	if s2 == nil {
		t.Fatal("expected structure, got nil")
	}

	got := slices.Collect(s2)
	if len(got) != 1 || got[0] != mdStart {
		t.Errorf("structure mismatch")
	}

	// 3. Test WithStructure(nil) -> GetStructure (should be nil)
	v3 := WithStructure[*SimpleValuer](SimpleValueAttachable{v2}, nil)
	if s3 := GetStructure(v3); s3 != nil {
		t.Errorf("expected nil structure after clearing, got %v", s3)
	}
}

// SimpleMetadata implements Metadata for structure testing
type SimpleMetadata struct {
	*SimpleValuer
}

func (m *SimpleMetadata) Children(yield func(Metadata) bool) {
	// no-op for this test
}
