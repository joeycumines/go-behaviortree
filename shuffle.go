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

import "math/rand"

// Shuffle implements randomised child execution order via encapsulation, using the provided source to shuffle the
// children prior to passing through to the provided tick (a nil source will use global math/rand), note that this
// function will return nil if a nil tick is provided
func Shuffle(tick Tick, source rand.Source) Tick {
	if tick == nil {
		return nil
	}
	if source == nil {
		source = defaultSource{}
	}
	return func(children []Node) (Status, error) {
		children = copyNodes(children)
		rand.New(source).Shuffle(len(children), func(i, j int) { children[i], children[j] = children[j], children[i] })
		return tick(children)
	}
}

type defaultSource struct{ rand.Source }

func (d defaultSource) Int63() int64 {
	return rand.Int63()
}
