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
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"reflect"
)

func replacePointers(b string) string {
	var (
		m = make(map[string]struct{})
		r []string
		n int
	)
	for _, v := range regexp.MustCompile(`(?:[[:^alnum:]]|^)(0x[[:alnum:]]{1,16})(?:[[:^alnum:]]|$)`).FindAllStringSubmatch(b, -1) {
		if v := v[1]; v != `0x0` {
			if _, ok := m[v]; !ok {
				n++
				m[v] = struct{}{}
				r = append(r, v, fmt.Sprintf(`%#x`, n))
			}
		}
	}
	// Normalize anonymous functions to generic format to avoid fragility (func1.func5 vs func1.func6 etc)
	// Regex matches e.g. ".func1.func5" or ".func12"
	reFunc := regexp.MustCompile(`\.func\d+(?:\.func\d+)*`)
	b = reFunc.ReplaceAllString(b, `.funcN`)

	return strings.NewReplacer(r...).Replace(b)
}

func TestNode_String(t *testing.T) {

	for _, testCase := range []struct {
		Name  string
		Node  Node
		Value string
	}{
		{
			Name:  `nil node`,
			Node:  nil,
			Value: `<nil>`,
		},
		{
			Name:  `single sequence`,
			Node:  New(Sequence),
			Value: "[0x1 printer_test.go:68 0x2 sequence.go:21]  github.com/joeycumines/go-behaviortree.TestNode_String | github.com/joeycumines/go-behaviortree.Sequence",
		},
		{
			Name:  `single closure`,
			Node:  New(func(children []Node) (Status, error) { panic(`TestNode_String`) }),
			Value: "[0x1 printer_test.go:73 0x2 printer_test.go:73]  github.com/joeycumines/go-behaviortree.TestNode_String | github.com/joeycumines/go-behaviortree.TestNode_String.funcN",
		},
		{
			Name:  `nil tick`,
			Node:  New(nil),
			Value: "[0x1 printer_test.go:78 0x0 -]  github.com/joeycumines/go-behaviortree.TestNode_String | <nil>",
		},
		{
			Name:  `example counter`,
			Node:  newExampleCounter(),
			Value: "[0x1 example_test.go:47 0x2 selector.go:21    ]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.Selector\n├── [0x3 example_test.go:49 0x4 sequence.go:21    ]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.Sequence\n│   ├── [0x5 example_test.go:51 0x6 example_test.go:52]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.newExampleCounter.funcN\n│   ├── [0x7 example_test.go:40 0x8 example_test.go:41]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.newExampleCounter.funcN\n│   └── [0x9 example_test.go:32 0xa example_test.go:33]  github.com/joeycumines/go-behaviortree.newExampleCounter.funcN | github.com/joeycumines/go-behaviortree.newExampleCounter.newExampleCounter.funcN\n└── [0xb example_test.go:62 0x4 sequence.go:21    ]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.Sequence\n    ├── [0xc example_test.go:64 0xd example_test.go:65]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.newExampleCounter.funcN\n    ├── [0x7 example_test.go:40 0x8 example_test.go:41]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.newExampleCounter.funcN\n    └── [0xe example_test.go:32 0xf example_test.go:33]  github.com/joeycumines/go-behaviortree.newExampleCounter.funcN | github.com/joeycumines/go-behaviortree.newExampleCounter.newExampleCounter.funcN",
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			value := testCase.Node.String()

			value = replacePointers(value)
			if value != testCase.Value {
				t.Errorf("unexpected value: %q\n> %s", value, strings.ReplaceAll(value, "\n", "\n> "))
			}
		})
	}
}

type mockPrinter struct {
	fprint func(output io.Writer, node Node) error
}

func (m *mockPrinter) Fprint(output io.Writer, node Node) error { return m.fprint(output, node) }

func TestNode_String_error(t *testing.T) {
	defer func() func() {
		old := DefaultPrinter
		DefaultPrinter = &mockPrinter{fprint: func(output io.Writer, node Node) error {
			return errors.New(`some_error`)
		}}
		return func() {
			DefaultPrinter = old
		}
	}()()
	if v := Node(nil).String(); v != `behaviortree.DefaultPrinter error: some_error` {
		t.Error(v)
	}
}

func TestTreePrinter_Fprint_copyError(t *testing.T) {
	r, w := io.Pipe()
	_ = r.Close()
	if err := (TreePrinter{Formatter: DefaultPrinterFormatter, Inspector: DefaultPrinterInspector}).Fprint(w, Node(nil)); err != io.ErrClosedPipe {
		t.Error(err)
	}
}

func dummyNode() (Tick, []Node) {
	return nil, nil
}

func TestDefaultPrinterInspector_nil(t *testing.T) {
	var actual [2]interface{}
	actual[0], actual[1] = DefaultPrinterInspector(nil, nil)
	if !reflect.DeepEqual(
		actual,
		[2]interface{}{
			[]interface{}{
				`0x0`,
				`-`,
				`0x0`,
				`-`,
			},
			`<nil> | <nil>`,
		},
	) {
		t.Errorf("unexpected diff\nexpected: %v\nactual:   %v", [2]interface{}{
			[]interface{}{
				`0x0`,
				`-`,
				`0x0`,
				`-`,
			},
			`<nil> | <nil>`,
		}, actual)
	}
	var node Node = dummyNode
	actual[0], actual[1] = DefaultPrinterInspector(node, nil)
	if !reflect.DeepEqual(
		actual,
		[2]interface{}{
			[]interface{}{
				fmt.Sprintf(`%p`, node),
				`printer_test.go:128`,
				`0x0`,
				`-`,
			},
			`github.com/joeycumines/go-behaviortree.dummyNode | <nil>`,
		},
	) {
		t.Errorf("unexpected diff\nexpected: %v\nactual:   %v", [2]interface{}{
			[]interface{}{
				fmt.Sprintf(`%p`, node),
				`printer_test.go:128`,
				`0x0`,
				`-`,
			},
			`github.com/joeycumines/go-behaviortree.dummyNode | <nil>`,
		}, actual)
	}
	tick := Selector
	actual[0], actual[1] = DefaultPrinterInspector(nil, tick)
	if !reflect.DeepEqual(
		actual,
		[2]interface{}{
			[]interface{}{
				`0x0`,
				`-`,
				fmt.Sprintf(`%p`, tick),
				`selector.go:21`,
			},
			`<nil> | github.com/joeycumines/go-behaviortree.Selector`,
		},
	) {
		t.Errorf("unexpected diff\nexpected: %v\nactual:   %v", [2]interface{}{
			[]interface{}{
				`0x0`,
				`-`,
				fmt.Sprintf(`%p`, tick),
				`selector.go:21`,
			},
			`<nil> | github.com/joeycumines/go-behaviortree.Selector`,
		}, actual)
	}
}

func TestTreePrinter_Fprint_emptyMeta(t *testing.T) {
	p := TreePrinter{
		Inspector: func(node Node, tick Tick) (meta []interface{}, value interface{}) {
			return []interface{}{``, ``, ``}, ``
		},
		Formatter: DefaultPrinterFormatter,
	}
	b := new(bytes.Buffer)
	_ = p.Fprint(b, nn(nil))
	if v := b.String(); v != `[  ]` {
		t.Errorf("unexpected value: %q", v)
	}
}

func TestTreePrinter_Alignment(t *testing.T) {
	// this checks that a deep child with expanding metadata affects the root column size
	p := TreePrinter{
		Inspector: func(node Node, tick Tick) (meta []interface{}, value interface{}) {
			return []interface{}{
				fmt.Sprintf("col1-%s", node.String()),
				fmt.Sprintf("col2-%s", node.String()),
			}, node.String()
		},
		Formatter: DefaultPrinterFormatter,
	}

	// Root(A) -> Child(B) -> Grandchild(C)
	// We want C's metadata to be LONG, forcing A and B to pad their columns.
	// But Node.String() is recursive so let's just use dummy strings via context/value if we could?
	// Or simpler: just use manually constructed treePrinterNodes if possible?
	// The test uses TreePrinter publicly.

	// Let's use a custom Inspector that returns long strings based on depth?
	// Not easy to track depth in Inspector.
	// We'll use the functional node value.

	// root: "root"
	// child: "child"
	// grandchild: "grandchild-very-long"

	// Mock nodes returns their name when ticked?
	// The Inspector implementation above calls `node.String()` which loops forever.
	// Let's make a simple inspector.

	p.Inspector = func(node Node, tick Tick) (meta []interface{}, value interface{}) {
		node()
		// name is actually the tick, but we can return strings directly from this mock.
		// Wait, the node structure:
		// root -> gets tick returns values.
		// let's assume we can map node pointer to name.
		return []interface{}{"c1", "c2"}, "val"
	}

	// But we need per-node differences.
	// Let's use the Formatter directly since that's what we are testing (the treePrinterNode implementation).
	root := DefaultPrinterFormatter()
	root.Add([]interface{}{"r1", "r2"}, "root")

	child := root.Add([]interface{}{"c1", "c2"}, "child")

	// grandchild has LONG metadata
	child.Add([]interface{}{"grandchild-1-very-long", "gc2"}, "grandchild")

	// We expect the output to have the first column padded to the width of "grandchild-1-very-long"
	// "grandchild-1-very-long" length is 22.
	// "r1" length is 2. Padding should be 20 spaces.

	// TreePrinterNode.Bytes() is what we test.
	b := root.Bytes()
	got := string(b)

	// Verify alignment
	// r1 is followed by r2.
	// [r1                   r2]  root
	// The space between r1 and r2 matches the max width of column 0.
	// Max width is 22. r1 is 2. Need 20 spaces.
	// But `print` adds one space separator between columns in the loop:
	// b.WriteString(s)
	// pad := sizes[i] - len(s) ... b.Write(spaces)
	// (next iteration) if i > 0 b.WriteByte(' ')

	// So for "r1":
	// Write "r1"
	// Pad = 22 - 2 = 20 spaces.
	// Space ' '
	// "r2"
	// "grandchild-1-very-long" (22 chars).
	// c1 (2 chars). Pad 20.
	// r1 (2 chars). Pad 20.

	// Wait, my manual calculation check:
	// sizes[0] = 22.
	// r1: len 2. diff 20.
	// c1: len 2. diff 20.
	// grandchild...: len 22. diff 0.

	// Let's check if the output contains the specific padded line for root.
	// "r1" + 20 spaces + " " + "r2" -> "r1                     r2" (21 spaces total between them?)
	// Code:
	// b.WriteString(s)
	// pad := ...; b.Write(pad)
	// Loop next: Space.

	// So yes, 20 spaces padding, then 1 space separator.

	// Just checking for the presence of the padded root string is enough to prove propagation.
	// r1 (2) + pad (20) + space (1) + r2 (2) + pad (1) -> 21 spaces between r1 and r2, plus 1 after.
	if !strings.Contains(got, "[r1                     r2 ]") {
		t.Errorf("alignment failed, root header not padded correctly:\n%s", got)
	}
	if !strings.Contains(got, "[c1                     c2 ]") {
		t.Errorf("alignment failed, child header not padded correctly:\n%s", got)
	}
}
