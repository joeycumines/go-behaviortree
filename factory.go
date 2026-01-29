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

var factory = defaultFactory

func defaultFactory(tick Tick, children []Node) Node {
	// N.B. we pick the leaf variant only if the children are _nil_ ONLY to
	// avoid a behavioral change, vs the old implementation
	if v := make([]uintptr, 1); runtimeCallers(3, v[:]) >= 1 {
		if v, _ := runtimeCallersFrames(v).Next(); v.PC != 0 {
			if children == nil {
				return (&leafNodeFrame{tick: tick, frame: NewFrame(v)}).node
			}
			return (&compositeNodeFrame{tick: tick, children: children, frame: NewFrame(v)}).node
		}
	}
	if children == nil {
		return leafNode(tick).node
	}
	return (&compositeNode{tick: tick, children: children}).node
}

type leafNode Tick

func (x leafNode) node() (Tick, []Node) { return Tick(x), nil }

type leafNodeFrame struct {
	tick  Tick
	frame Frame
}

func (x *leafNodeFrame) Value(key any) (any, bool) {
	if key == (vkFrame{}) {
		frame := x.frame
		return &frame, true
	}
	return nil, false
}

func (x *leafNodeFrame) node() (Tick, []Node) {
	UseValueProvider(x)
	return x.tick, nil
}

type compositeNode struct {
	tick     Tick
	children []Node
}

func (x *compositeNode) node() (Tick, []Node) {
	return x.tick, x.children
}

type compositeNodeFrame struct {
	tick     Tick
	children []Node
	frame    Frame
}

func (x *compositeNodeFrame) Value(key any) (any, bool) {
	if key == (vkFrame{}) {
		frame := x.frame
		return &frame, true
	}
	return nil, false
}

func (x *compositeNodeFrame) node() (Tick, []Node) {
	UseValueProvider(x)
	return x.tick, x.children
}
