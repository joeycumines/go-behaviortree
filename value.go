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
	"errors"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

var (
	runtimeCallers       = runtime.Callers
	runtimeCallersFrames = runtime.CallersFrames
	runtimeFuncForPC     = runtime.FuncForPC
)

var (
	valueCallMutex  sync.Mutex
	valueDataMutex  sync.RWMutex
	valueDataKey    any
	valueDataChan   chan any
	valueDataCaller [1]uintptr
	valueActive     uint32
)

// WithValue will return the receiver wrapped with a key-value pair, using similar semantics to the context package.
//
// Values should only be used to attach information to BTs in a way that transits API boundaries, not for passing
// optional parameters to functions. Some package-level synchronisation was necessary to facilitate this mechanism. As
// such, this and the Node.Value method should be used with caution, preferably only outside normal operation.
//
// The same restrictions on the key apply as for context.WithValue.
func (n Node) WithValue(key, value any) Node {
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
		UseValueHandler(func(k any) (any, bool) {
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
func (n Node) Value(key any) any {
	if n != nil {
		valueCallMutex.Lock()
		defer valueCallMutex.Unlock()
		return n.valueSync(key)
	}
	return nil
}

// valueSync is split out into it's own method and the Node.valuePrep call is used as a discriminator for relevant
// values
func (n Node) valueSync(key any) (value any) {
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

func (n Node) valuePrep(key any) bool {
	valueDataMutex.Lock()
	if runtimeCallers(2, valueDataCaller[:]) < 1 {
		valueDataMutex.Unlock()
		return false
	}
	valueDataKey = key
	valueDataChan = make(chan any, 1)
	atomic.StoreUint32(&valueActive, 1)
	valueDataMutex.Unlock()
	defer atomic.StoreUint32(&valueActive, 0)
	n()
	return true
}

// ValueProvider defines a mechanism to provide values for specific keys.
type ValueProvider interface {
	// Value returns the value associated with the key, or (nil, false) if not found.
	Value(key any) (any, bool)
}

// ValueProviderFunc is a function type that implements [ValueProvider].
type ValueProviderFunc func(key any) (any, bool)

// Value implements the [ValueProvider] interface.
func (f ValueProviderFunc) Value(key any) (any, bool) {
	return f(key)
}

// UseValueHandler is a convenience wrapper for [UseValueProvider] that accepts a function.
func UseValueHandler(fn func(key any) (any, bool)) {
	UseValueProvider(ValueProviderFunc(fn))
}

// ValueProviders is a slice of [ValueProvider] that implements [ValueProvider] itself.
// It iterates over the providers in order and returns the first found value.
type ValueProviders []ValueProvider

// Value implements the ValueProvider interface.
func (v ValueProviders) Value(key any) (any, bool) {
	for _, p := range v {
		if val, ok := p.Value(key); ok {
			return val, true
		}
	}
	return nil, false
}

// UseValueProviders is a convenience wrapper for [UseValueProvider] that accepts multiple providers.
// They will be combined into a single [ValueProviders] slice.
func UseValueProviders(providers ...ValueProvider) {
	UseValueProvider(ValueProviders(providers))
}

// UseValueProvider may be called when expanding _any_ [Node], and allows the provided
// provider to respond to [Node.Value] queries for that node. If there are multiple,
// called during node expansion, the outermost one that responds with true will take precedence.
// In the absence of an in-progress [Node.Value] call, the cost of this function is trivial.
// See also [Node.WithValue], which is implemented using this mechanism.
//
// Use cases:
//   - Attaching metadata to custom nodes (e.g., frame/caller information)
//   - Efficiently handling large numbers of key-value pairs without wrapping nodes repeatedly
//   - Debugging or inspection of custom node implementations during development
//
// The provider receives a key and should return (value, true) if it handles that key,
// or (nil, false) otherwise. Only providers registered on nodes appearing in the call stack
// from the [Node.Value] call will be considered.
//
// Example: Register a custom metadata handler in a custom node factory
//
//	func customNode(logic bt.Tick) bt.Node {
//		handler := func(key any) (any, bool) {
//			if metadataKey, ok := key.(metadataKeyType); ok {
//				return nodeMetadata{info: "custom"}, true
//			}
//			return nil, false
//		}
//		return func() (bt.Tick, []bt.Node) {
//			bt.UseValueHandler(handler) // This "registers" the handler
//			return logic, nil
//		}
//	}
func UseValueProvider(provider ValueProvider) {
	if atomic.LoadUint32(&valueActive) == 0 {
		return
	}
	valueDataMutex.RLock()
	dataKey, dataChan, dataCaller := valueDataKey, valueDataChan, valueDataCaller
	valueDataMutex.RUnlock()

	// fast exit case 1: there is no pending value operation
	if dataChan == nil {
		return
	}

	// fast exit case 2: pending value operation is not relevant
	value, ok := provider.Value(dataKey)
	if !ok {
		return
	}
	dataKey = nil

	// slow case, may require walking the entire call stack
	const depth = 2 << 7
	callers := make([]uintptr, depth)
	skip := 2
	for {
		n := runtimeCallers(skip, callers)
		if n == 0 {
			break
		}
		for _, pc := range callers[:n] {
			if pc == dataCaller[0] {
				select {
				case dataChan <- value:
				default:
				}
				return
			}
		}
		if n < len(callers) {
			break
		}
		skip += n
	}
}
