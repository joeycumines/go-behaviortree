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
	"errors"
	"reflect"
	"runtime"
	"sync"
)

var (
	runtimeCallers       = runtime.Callers
	runtimeCallersFrames = runtime.CallersFrames
	runtimeFuncForPC     = runtime.FuncForPC
)

var (
	valueCallMutex  sync.Mutex
	valueDataMutex  sync.RWMutex
	valueDataKey    interface{}
	valueDataChan   chan interface{}
	valueDataCaller [1]uintptr
)

// WithValue will return the receiver wrapped with a key-value pair, using similar semantics to the context package.
//
// Values should only be used to attach information to BTs in a way that transits API boundaries, not for passing
// optional parameters to functions. Some package-level synchronisation was necessary to facilitate this mechanism. As
// such, this and the Node.Value method should be used with caution, preferably only outside normal operation.
//
// The same restrictions on the key apply as for context.WithValue.
func (n Node) WithValue(key, value interface{}) Node {
	if n == nil {
		panic(errors.New(`behaviortree.Node.WithValue nil receiver`))
	}
	if key == nil {
		panic(errors.New(`behaviortree.Node.WithValue nil key`))
	}
	if !reflect.TypeOf(key).Comparable() {
		panic(errors.New(`behaviortree.Node.WithValue key is not comparable`))
	}
	return func() (Tick, []Node) {
		n.valueHandle(func(k interface{}) (interface{}, bool) {
			if k == key {
				return value, true
			}
			return nil, false
		})
		return n()
	}
}

// Value will return the value associated with this node for key, or nil if there is none.
//
// See also Node.WithValue, as well as the value mechanism provided by the context package.
func (n Node) Value(key interface{}) interface{} {
	if n != nil {
		valueCallMutex.Lock()
		defer valueCallMutex.Unlock()
		return n.valueSync(key)
	}
	return nil
}

// valueSync is split out into it's own method and the Node.valuePrep call is used as a discriminator for relevant
// values
func (n Node) valueSync(key interface{}) (value interface{}) {
	if n.valuePrep(key) {
		select {
		case value = <-valueDataChan:
		default:
		}
		valueDataMutex.Lock()
		valueDataKey = nil
		valueDataChan = nil
		valueDataMutex.Unlock()
	}
	return
}

func (n Node) valuePrep(key interface{}) bool {
	valueDataMutex.Lock()
	if runtimeCallers(2, valueDataCaller[:]) < 1 {
		valueDataMutex.Unlock()
		return false
	}
	valueDataKey = key
	valueDataChan = make(chan interface{}, 1)
	valueDataMutex.Unlock()
	n()
	return true
}

func (n Node) valueHandle(fn func(key interface{}) (interface{}, bool)) {
	valueDataMutex.RLock()
	dataKey, dataChan, dataCaller := valueDataKey, valueDataChan, valueDataCaller
	valueDataMutex.RUnlock()

	// fast exit case 1: there is no pending value operation
	if dataChan == nil {
		return
	}

	// fast exit case 2: pending value operation is not relevant
	value, ok := fn(dataKey)
	if !ok {
		return
	}
	dataKey = nil

	// slow case, may require walking the entire call stack
	const depth = 2 << 7
	callers := make([]uintptr, depth)
	for skip := 4; skip > 0; skip += depth {
		callers = callers[:runtimeCallers(skip, callers[:])]
		for _, caller := range callers {
			if caller == dataCaller[0] {
				select {
				case dataChan <- value:
				default:
				}
				return
			}
		}
		if len(callers) != depth {
			return
		}
	}
}
