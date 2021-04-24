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
	"context"
	"time"
)

// Context provides support for tick(s) utilising context as a means of cancelation, with cancelation triggered by
// either BT-driven logic or the normal means (parent cancelation, deadline / timeout).
//
// Note that it must be initialised by means of it's Init method (implements a tick) prior to use (Context.Tick tick).
// Init may be ticked any number of times (each time triggering cancelation of any prior context).
type Context struct {
	parent func() (context.Context, context.CancelFunc)
	ctx    context.Context
	cancel context.CancelFunc
}

// WithCancel configures the receiver to initialise context like context.WithCancel(parent), returning the receiver
func (c *Context) WithCancel(parent context.Context) *Context {
	c.parent = func() (context.Context, context.CancelFunc) { return context.WithCancel(parent) }
	return c
}

// WithDeadline configures the receiver to initialise context like context.WithDeadline(parent, deadline), returning
// the receiver
func (c *Context) WithDeadline(parent context.Context, deadline time.Time) *Context {
	c.parent = func() (context.Context, context.CancelFunc) { return context.WithDeadline(parent, deadline) }
	return c
}

// WithTimeout configures the receiver to initialise context like context.WithTimeout(parent, timeout), returning
// the receiver
func (c *Context) WithTimeout(parent context.Context, timeout time.Duration) *Context {
	c.parent = func() (context.Context, context.CancelFunc) { return context.WithTimeout(parent, timeout) }
	return c
}

// Init implements a tick that will cancel existing context, (re)initialise the context, then succeed, note that it
// must not be called concurrently with any other method, and it must be ticked prior to any Context.Tick tick
func (c *Context) Init([]Node) (Status, error) {
	if c.cancel != nil {
		c.cancel()
	}
	if c.parent != nil {
		c.ctx, c.cancel = c.parent()
	} else {
		c.ctx, c.cancel = context.WithCancel(context.Background())
	}
	return Success, nil
}

// Tick returns a tick that will call fn with the receiver's context, returning nil if fn is nil (for consistency
// with other implementations in this package), note that a Init node must have already been ticked on all possible
// execution paths, or a panic may occur, due to fn being passed a nil context.Context
func (c *Context) Tick(fn func(ctx context.Context, children []Node) (Status, error)) Tick {
	if fn != nil {
		return func(children []Node) (Status, error) { return fn(c.ctx, children) }
	}
	return nil
}

// Cancel implements a tick that will cancel the receiver's context (noop if it has none) then succeed
func (c *Context) Cancel([]Node) (Status, error) {
	if c.cancel != nil {
		c.cancel()
	}
	return Success, nil
}

// Err implements a tick that will succeed if the receiver does not have a context or it has been canceled
func (c *Context) Err([]Node) (Status, error) {
	if c.ctx == nil || c.ctx.Err() != nil {
		return Success, nil
	}
	return Failure, nil
}

// Done implements a tick that will block on the receiver's context being canceled (noop if it has none) then succeed
func (c *Context) Done([]Node) (Status, error) {
	if c.ctx != nil {
		<-c.ctx.Done()
	}
	return Success, nil
}
