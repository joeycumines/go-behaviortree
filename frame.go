/*
   Copyright 2021 Joseph Cumines

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
	"reflect"
)

type (
	// Frame is a partial copy of runtime.Frame.
	//
	// This packages captures details about the caller of it's New and NewNode functions, embedding them into the
	// nodes themselves, for tree printing / tracing purposes.
	Frame struct {
		// PC is the program counter for the location in this frame.
		// For a frame that calls another frame, this will be the
		// program counter of a call instruction. Because of inlining,
		// multiple frames may have the same PC value, but different
		// symbolic information.
		PC uintptr
		// Function is the package path-qualified function name of
		// this call frame. If non-empty, this string uniquely
		// identifies a single function in the program.
		// This may be the empty string if not known.
		Function string
		// File and Line are the file name and line number of the
		// location in this frame. For non-leaf frames, this will be
		// the location of a call. These may be the empty string and
		// zero, respectively, if not known.
		File string
		Line int
		// Entry point program counter for the function; may be zero
		// if not known.
		Entry uintptr
	}

	vkFrame struct{}
)

// Frame will return the call frame for the caller of New/NewNode, an approximation based on the receiver, or nil.
//
// This method uses the Value mechanism and is subject to the same warnings / performance limitations.
func (n Node) Frame() *Frame {
	if v, _ := n.Value(vkFrame{}).(*Frame); v != nil {
		v := *v
		return &v
	}
	return newFrame(n)
}

// Frame will return an approximation of a call frame based on the receiver, or nil.
func (t Tick) Frame() *Frame { return newFrame(t) }

func newFrame(v interface{}) (f *Frame) {
	if v := reflect.ValueOf(v); v.IsValid() && v.Kind() == reflect.Func && !v.IsNil() {
		p := v.Pointer()
		if v := runtimeFuncForPC(p); v != nil {
			f = &Frame{
				PC:       p,
				Function: v.Name(),
				Entry:    v.Entry(),
			}
			f.File, f.Line = v.FileLine(f.Entry)
		}
	}
	return
}
