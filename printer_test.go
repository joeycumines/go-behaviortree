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
	"github.com/go-test/deep"
	"io"
	"regexp"
	"runtime"
	"strings"
	"testing"
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
			Value: "[0x1 printer_test.go:62 0x2 sequence.go:21]  github.com/joeycumines/go-behaviortree.TestNode_String | github.com/joeycumines/go-behaviortree.Sequence",
		},
		{
			Name:  `single closure`,
			Node:  New(func(children []Node) (Status, error) { panic(`TestNode_String`) }),
			Value: "[0x1 printer_test.go:67 0x2 printer_test.go:67]  github.com/joeycumines/go-behaviortree.TestNode_String | github.com/joeycumines/go-behaviortree.TestNode_String.func1",
		},
		{
			Name:  `nil tick`,
			Node:  New(nil),
			Value: "[0x1 printer_test.go:72 0x0 -]  github.com/joeycumines/go-behaviortree.TestNode_String | <nil>",
		},
		{
			Name:  `example counter`,
			Node:  newExampleCounter(),
			Value: "[0x1 example_test.go:47 0x2 selector.go:21]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.Selector\n├── [0x3 example_test.go:49 0x4 sequence.go:21]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.Sequence\n│   ├── [0x5 example_test.go:51 0x6 example_test.go:52]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.newExampleCounter.func3\n│   ├── [0x7 example_test.go:40 0x8 example_test.go:41]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.newExampleCounter.func2\n│   └── [0x9 example_test.go:32 0xa example_test.go:33]  github.com/joeycumines/go-behaviortree.newExampleCounter.func1 | github.com/joeycumines/go-behaviortree.newExampleCounter.func1.1\n└── [0xb example_test.go:62 0x4 sequence.go:21]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.Sequence\n    ├── [0xc example_test.go:64 0xd example_test.go:65]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.newExampleCounter.func4\n    ├── [0x7 example_test.go:40 0x8 example_test.go:41]  github.com/joeycumines/go-behaviortree.newExampleCounter | github.com/joeycumines/go-behaviortree.newExampleCounter.func2\n    └── [0x9 example_test.go:32 0xa example_test.go:33]  github.com/joeycumines/go-behaviortree.newExampleCounter.func1 | github.com/joeycumines/go-behaviortree.newExampleCounter.func1.1",
		},
	} {
		t.Run(testCase.Name, func(t *testing.T) {
			value := testCase.Node.String()
			//t.Logf("\n---\n%s\n---", value)
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

func Test_treePrinterNodeXlabMeta_String_panicLen(t *testing.T) {
	defer func() {
		if r := fmt.Sprint(recover()); r != `m.sizes [4] mismatched m.strings [one two]` {
			t.Error(r)
		}
	}()
	m := &treePrinterNodeXlabMeta{
		treePrinterNodeXlab: &treePrinterNodeXlab{
			sizes: []int{4},
		},
		strings: []string{`one`, `two`},
	}
	_ = m.String()
	t.Error(`expected panic`)
}

func Test_treePrinterNodeXlabMeta_String_panicInterfaces(t *testing.T) {
	defer func() {
		if r := fmt.Sprint(recover()); r != `m.interfaces [] should be nil` {
			t.Error(r)
		}
	}()
	m := &treePrinterNodeXlabMeta{
		treePrinterNodeXlab: &treePrinterNodeXlab{
			sizes: []int{4},
		},
		strings:    []string{`one`, `two`},
		interfaces: make([]interface{}, 0),
	}
	_ = m.String()
	t.Error(`expected panic`)
}

func dummyNode() (Tick, []Node) {
	return nil, nil
}

func TestDefaultPrinterInspector_nil(t *testing.T) {
	var actual [2]interface{}
	actual[0], actual[1] = DefaultPrinterInspector(nil, nil)
	if diff := deep.Equal(
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
	); diff != nil {
		t.Errorf("unexpected diff:\n%s", strings.Join(diff, "\n"))
	}
	var node Node = dummyNode
	actual[0], actual[1] = DefaultPrinterInspector(node, nil)
	if diff := deep.Equal(
		actual,
		[2]interface{}{
			[]interface{}{
				fmt.Sprintf(`%p`, node),
				`printer_test.go:155`,
				`0x0`,
				`-`,
			},
			`github.com/joeycumines/go-behaviortree.dummyNode | <nil>`,
		},
	); diff != nil {
		t.Errorf("unexpected diff:\n%s", strings.Join(diff, "\n"))
	}
	tick := Selector
	actual[0], actual[1] = DefaultPrinterInspector(nil, tick)
	if diff := deep.Equal(
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
	); diff != nil {
		t.Errorf("unexpected diff:\n%s", strings.Join(diff, "\n"))
	}
}

func TestDefaultPrinterInspector_noName(t *testing.T) {
	defer func() func() {
		old := runtimeFuncForPC
		runtimeFuncForPC = func(pc uintptr) *runtime.Func { return nil }
		return func() {
			runtimeFuncForPC = old
		}
	}()()
	var actual [2]interface{}
	actual[0], actual[1] = DefaultPrinterInspector(dummyNode, Sequence)
	if diff := deep.Equal(
		actual,
		[2]interface{}{
			[]interface{}{
				`0x0`,
				`-`,
				`0x0`,
				`-`,
			},
			`- | -`,
		},
	); diff != nil {
		t.Errorf("unexpected diff:\n%s", strings.Join(diff, "\n"))
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
	if v := string(b.Bytes()); v != `[  ]  ` {
		t.Error(v)
	}
}
