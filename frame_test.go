/*
   Copyright 2026 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package behaviortree

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNode_Frame(t *testing.T) {
	for _, test := range []struct {
		Name  string
		Node  Node
		Frame *Frame
	}{
		{
			Name: `nil`,
		},
		{
			Name: `nn with value`,
			Node: nn(nil, nil).WithValue(1, 2),
			Frame: &Frame{
				Function: `github.com/joeycumines/go-behaviortree.Node.WithValue.func1`,
				File:     `value.go`,
			},
		},
		{
			Name: `nn with value explicit frame`,
			Node: nn(nil, nil).WithFrame(&Frame{
				PC:       0x568dc0,
				Function: "github.com/joeycumines/go-behaviortree.glob..func1.1",
				File:     "C:/Users/under/go/src/github.com/joeycumines/go-behaviortree/behaviortree.go",
				Line:     53,
				Entry:    0x568dc0,
			}),
			Frame: &Frame{
				Function: "github.com/joeycumines/go-behaviortree.glob..func1.1",
				File:     "behaviortree.go",
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			for i := 0; i < 2; i++ {
				frame := test.Node.Frame()
				if frame != nil {
					if frame.Line == 0 {
						t.Error(frame.Line)
					}
					if frame.PC == 0 {
						t.Error(frame.PC)
					}
					if frame.Entry == 0 {
						t.Error(frame.Entry)
					}
					file := frame.File

					// Verify Immutability: Modifying the returned frame instance must not
					// affect the internal state of the Node or future calls.
					frame.Line = 0
					if f := test.Node.Frame(); f == nil || f.Line == 0 {
						t.Error("safety regression: internal implementation detail exposed (mutable)")
					}

					file = filepath.Base(file)

					if test.Frame != nil && file != test.Frame.File {
						t.Errorf("file mismatch: %s != %s", file, test.Frame.File)
					}
				}

				if (frame == nil) != (test.Frame == nil) {
					t.Errorf("frame nil presence mismatch: %v", frame)
				} else if frame != nil {
					// Manual comparison is required because the runtime Frame contains
					// the full absolute path, while the test expectation uses the base filename.
					f := frame.File
					if i := strings.LastIndex(f, "/"); i >= 0 {
						f = f[i+1:]
					} else if i := strings.LastIndex(f, "\\"); i >= 0 {
						f = f[i+1:]
					}

					if f != test.Frame.File {
						t.Errorf("File: %s != %s", f, test.Frame.File)
					}
					if frame.Function != test.Frame.Function {
						t.Errorf("Function: %s != %s", frame.Function, test.Frame.Function)
					}
					// PC, Line, and Entry are dynamic or checked for non-zero values above.
				}
			}
		})
	}
}

func TestTick_Frame(t *testing.T) {
	for _, test := range []struct {
		Name  string
		Tick  Tick
		Frame *Frame
	}{
		{
			Name: `nil`,
		},
		{
			Name: `sequence`,
			Tick: Sequence,
			Frame: &Frame{
				Function: `github.com/joeycumines/go-behaviortree.Sequence`,
				File:     `sequence.go`,
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			for i := 0; i < 2; i++ {
				frame := test.Tick.Frame()
				if frame != nil {
					if frame.Line == 0 {
						t.Error(frame.Line)
					}
					if frame.PC == 0 {
						t.Error(frame.PC)
					}
					if frame.Entry == 0 {
						t.Error(frame.Entry)
					}

					// Verify Immutability: Ensure subsequent calls return a correct, fresh frame.
					frame.Line = 0
					if f := test.Tick.Frame(); f == nil || f.Line == 0 {
						t.Error("tick frame integrity compromised")
					}
				}
				if (frame == nil) != (test.Frame == nil) {
					t.Errorf("frame nil mismatch")
				} else if frame != nil {
					f := frame.File
					if i := strings.LastIndex(f, "/"); i >= 0 {
						f = f[i+1:]
					} else if i := strings.LastIndex(f, "\\"); i >= 0 {
						f = f[i+1:]
					}
					if f != test.Frame.File {
						t.Errorf("File: %s != %s", f, test.Frame.File)
					}
					if frame.Function != test.Frame.Function {
						t.Errorf("Function: %s != %s", frame.Function, test.Frame.Function)
					}
				}
			}
		})
	}
}

func TestUseFrame(t *testing.T) {
	f := &Frame{Function: "test_frame_func", Line: 123}

	// 1. Direct Provider
	p := UseFrame(f)
	if v, ok := p.Value(vkFrame{}); !ok {
		t.Error("UseFrame returned false for vkFrame")
	} else {
		got := v.(*Frame)
		if got.Function != f.Function {
			t.Errorf("got function %q; want %q", got.Function, f.Function)
		}
	}

	// 2. Empty Frame
	fEmpty := &Frame{}
	pEmpty := UseFrame(fEmpty)
	if v, ok := pEmpty.Value(vkFrame{}); !ok {
		t.Error("UseFrame(empty) returned false")
	} else {
		got := v.(*Frame)
		if got != fEmpty {
			t.Errorf("UseFrame(empty) returned %p; want %p", got, fEmpty)
		}
	}

	// 3. Nil Frame
	pNil := UseFrame(nil)
	if v, ok := pNil.Value(vkFrame{}); !ok {
		t.Error("UseFrame(nil) should return true for vkFrame")
	} else if v != nil {
		t.Errorf("UseFrame(nil) should return nil value; got %v", v)
	}

	// 4. Integration
	var node Node = func() (Tick, []Node) {
		UseValueProvider(UseFrame(f))
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}

	// Use GetFrame to verify exact attached value
	gotF := GetFrame(node)
	if gotF == nil {
		t.Fatal("GetFrame(node) returned nil")
	}
	if gotF.Function != "test_frame_func" {
		t.Errorf("GetFrame(node).Function = %q", gotF.Function)
	}

	// 5. Integration with Nil
	var nodeNil Node = func() (Tick, []Node) {
		UseValueProvider(UseFrame(nil))
		return func(children []Node) (Status, error) { return Success, nil }, nil
	}

	if gf := GetFrame(nodeNil); gf != nil {
		t.Errorf("GetFrame(nodeNil) = %v; want nil", gf)
	}

	// 6. Unknown Key
	if _, ok := p.Value("unknown_key"); ok {
		t.Error("UseFrame returned true for unknown key")
	}
}
