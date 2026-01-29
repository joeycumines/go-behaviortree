package behaviortree

import (
	"testing"
)

func TestValueProviders(t *testing.T) {
	type k1 struct{}
	type k2 struct{}
	type k3 struct{}

	p1 := ValueProviderFunc(func(key any) (any, bool) {
		if _, ok := key.(k1); ok {
			return "v1", true
		}
		return nil, false
	})
	p2 := ValueProviderFunc(func(key any) (any, bool) {
		if _, ok := key.(k2); ok {
			return "v2", true
		}
		return nil, false
	})
	p3 := ValueProviderFunc(func(key any) (any, bool) {
		if _, ok := key.(k1); ok { // Masked by p1
			return "v1-shadowed", true
		}
		if _, ok := key.(k3); ok {
			return "v3", true
		}
		return nil, false
	})

	providers := ValueProviders{p1, p2, p3}

	if v, ok := providers.Value(k1{}); !ok || v != "v1" {
		t.Errorf("p1 failed: %v, %v", v, ok)
	}
	if v, ok := providers.Value(k2{}); !ok || v != "v2" {
		t.Errorf("p2 failed: %v, %v", v, ok)
	}
	if v, ok := providers.Value(k3{}); !ok || v != "v3" {
		t.Errorf("p3 failed: %v, %v", v, ok)
	}
	if _, ok := providers.Value("unknown"); ok {
		t.Error("unknown key failed")
	}
}

func TestUseValueProviders(t *testing.T) {
	type key struct{}

	var node Node = func() (Tick, []Node) {
		UseValueProviders(
			ValueProviderFunc(func(k any) (any, bool) {
				if k == (key{}) {
					return "found", true
				}
				return nil, false
			}),
			ValueProviderFunc(func(k any) (any, bool) {
				return "shadowed", true
			}),
		)
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}

	if v := node.Value(key{}); v != "found" {
		t.Errorf("expected found, got %v", v)
	}
}

func TestUseValueHandler_shadowing(t *testing.T) {
	// Ensure that UseValueHandler wraps correctly and works
	type key struct{}
	var node Node = func() (Tick, []Node) {
		UseValueHandler(func(k any) (any, bool) {
			if k == (key{}) {
				return "val", true
			}
			return nil, false
		})
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}
	if v := node.Value(key{}); v != "val" {
		t.Errorf("expected val, got %v", v)
	}
}

func TestValueProviders_Empty(t *testing.T) {
	var providers ValueProviders
	if _, ok := providers.Value("anything"); ok {
		t.Error("expected false for empty providers")
	}
}

func TestValueProviderFunc(t *testing.T) {
	called := false
	pf := ValueProviderFunc(func(key any) (any, bool) {
		called = true
		return "val", true
	})
	if v, ok := pf.Value("key"); !ok || v != "val" {
		t.Errorf("unexpected result: %v, %v", v, ok)
	}
	if !called {
		t.Error("expected func to be called")
	}
}
