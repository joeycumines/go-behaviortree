/*
   Copyright 2020 Joseph Cumines

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
			Node: nn(nil, nil).WithValue(vkFrame{}, &Frame{
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
					} else {
						frame.Line = 0
					}
					if frame.PC == 0 {
						t.Error(frame.PC)
					} else {
						frame.PC = 0
					}
					if frame.Entry == 0 {
						t.Error(frame.Entry)
					} else {
						frame.Entry = 0
					}
					if i := strings.LastIndex(frame.File, "/"); i >= 0 {
						frame.File = frame.File[i+1:]
					} else {
						t.Error(frame.File)
					}
				}
				if (frame == nil) != (test.Frame == nil) || (frame != nil && *frame != *test.Frame) {
					t.Errorf("%+v", frame)
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
					} else {
						frame.Line = 0
					}
					if frame.PC == 0 {
						t.Error(frame.PC)
					} else {
						frame.PC = 0
					}
					if frame.Entry == 0 {
						t.Error(frame.Entry)
					} else {
						frame.Entry = 0
					}
					if i := strings.LastIndex(frame.File, "/"); i >= 0 {
						frame.File = frame.File[i+1:]
					} else {
						t.Error(frame.File)
					}
				}
				if (frame == nil) != (test.Frame == nil) || (frame != nil && *frame != *test.Frame) {
					t.Errorf("%+v", frame)
				}
			}
		})
	}
}
