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
	"context"
	"testing"
	"time"
)

func TestContext_Tick_nilFn(t *testing.T) {
	if v := new(Context).Tick(nil); v != nil {
		t.Error(`expected nil`)
	}
}

func TestContext_Cancel_noContext(t *testing.T) {
	if status, err := new(Context).Cancel(nil); err != nil || status != Success {
		t.Error(status, err)
	}
}

func TestContext_Done_noContext(t *testing.T) {
	if status, err := new(Context).Done(nil); err != nil || status != Success {
		t.Error(status, err)
	}
}

func TestContext_Init_default(t *testing.T) {
	c := new(Context)
	if status, err := c.Init(nil); err != nil || status != Success {
		t.Fatal(status, err)
	}
	ctx := c.ctx
	if err := ctx.Err(); err != nil {
		t.Fatal(err)
	}
	if status, err := c.Init(nil); err != nil || status != Success {
		t.Fatal(status, err)
	}
	if err := ctx.Err(); err == nil {
		t.Error(c)
	}
	if c.ctx == nil || c.cancel == nil || c.ctx.Err() != nil {
		t.Fatal(c)
	}
	c.cancel()
	if c.ctx.Err() == nil {
		t.Fatal(c)
	}
}

func TestContext_WithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c := new(Context)
	if v := c.WithCancel(ctx); v != c {
		t.Error(v)
	}
	if c.ctx != nil || c.cancel != nil || c.parent == nil {
		t.Fatal(c)
	}
	if status, err := c.Init(nil); err != nil || status != Success {
		t.Fatal(status, err)
	}
	if c.ctx == nil || c.cancel == nil || c.ctx.Err() != nil {
		t.Fatal(c)
	}
	cancel()
	if err := c.ctx.Err(); err != context.Canceled {
		t.Fatal(err)
	}
}

func TestContext_WithDeadline(t *testing.T) {
	c := new(Context)
	if v := c.WithDeadline(context.Background(), time.Now().Add(-time.Second)); v != c {
		t.Error(v)
	}
	if c.ctx != nil || c.cancel != nil || c.parent == nil {
		t.Fatal(c)
	}
	if status, err := c.Init(nil); err != nil || status != Success {
		t.Fatal(status, err)
	}
	if err := c.ctx.Err(); err != context.DeadlineExceeded {
		t.Fatal(err)
	}
}
