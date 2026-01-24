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
	"fmt"
	"io"
	"strconv"
	"strings"
)

type (
	// Printer models something providing behavior tree printing capabilities
	Printer interface {
		// Fprint writes a representation node to output
		Fprint(output io.Writer, node Node) error
	}

	// TreePrinter provides a generalised implementation of Printer used as the DefaultPrinter
	TreePrinter struct {
		// Inspector configures the meta and value for a node with a given tick
		Inspector func(node Node, tick Tick) (meta []interface{}, value interface{})
		// Formatter initialises a new printer tree and returns it as a TreePrinterNode
		Formatter func() TreePrinterNode
	}

	// TreePrinterNode models a BT node for printing and is used by the TreePrinter implementation in this package
	TreePrinterNode interface {
		// Add should wire up a new node to the receiver then return it
		Add(meta []interface{}, value interface{}) TreePrinterNode
		// Bytes should encode the node and all children in preparation for use within TreePrinter
		Bytes() []byte
	}
)

var (
	// DefaultPrinter is used to implement Node.String
	DefaultPrinter Printer = TreePrinter{
		Inspector: DefaultPrinterInspector,
		Formatter: DefaultPrinterFormatter,
	}
)

// String implements fmt.Stringer using DefaultPrinter
func (n Node) String() string {
	var b bytes.Buffer
	if err := DefaultPrinter.Fprint(&b, n); err != nil {
		return fmt.Sprintf("behaviortree.DefaultPrinter error: %s", err)
	}
	return b.String()
}

// DefaultPrinterInspector is used by DefaultPrinter
func DefaultPrinterInspector(node Node, tick Tick) ([]interface{}, interface{}) {
	var (
		nodeName string
		tickName string
	)
	var nodeStrings, tickStrings frameStrings
	if v := node.Frame(); v != nil {
		nodeStrings = getFrameStrings(v)
		nodeName = v.Function
	} else if node == nil {
		nodeStrings.ptr = "0x0"
		nodeStrings.file = "-"
		nodeName = "<nil>"
	}
	if nodeStrings.file == "" {
		nodeStrings.file = "-"
	}
	if nodeName == "" {
		nodeName = "-"
	}
	if name := node.Name(); name != "" {
		nodeName = name
	}

	if v := tick.Frame(); v != nil {
		tickStrings = getFrameStrings(v)
		tickName = v.Function
	} else if tick == nil {
		tickStrings.ptr = "0x0"
		tickStrings.file = "-"
		tickName = "<nil>"
	}
	if tickStrings.file == "" {
		tickStrings.file = "-"
	}
	if tickName == "" {
		tickName = "-"
	}

	// Defaults for empty strings (e.g. if Mock prevented Frame lookup)
	if nodeStrings.ptr == "" {
		nodeStrings.ptr = "0x0"
	}
	if tickStrings.ptr == "" {
		tickStrings.ptr = "0x0"
	}

	return []interface{}{
		nodeStrings.ptr,
		nodeStrings.file,
		tickStrings.ptr,
		tickStrings.file,
	}, nodeName + " | " + tickName
}

type frameStrings struct {
	ptr  string
	file string
}

func getFrameStrings(f *Frame) (s frameStrings) {
	s.ptr = formatPtr(f.PC)
	s.file = shortFileLine(f.File, f.Line)
	return
}

// Fprint implements Printer.Fprint
func (p TreePrinter) Fprint(output io.Writer, node Node) error {
	tree := p.Formatter()
	p.build(tree, node)
	if _, err := io.Copy(output, bytes.NewReader(tree.Bytes())); err != nil {
		return err
	}
	return nil
}

func (p TreePrinter) build(tree TreePrinterNode, node Node) {
	if node != nil {
		tick, children := node()
		tree = tree.Add(p.Inspector(node, tick))
		for _, child := range children {
			p.build(tree, child)
		}
	}
}

func formatPtr(p uintptr) string {
	if p == 0 {
		return "0x0"
	}
	return "0x" + strconv.FormatUint(uint64(p), 16)
}

func shortFileLine(f string, l int) string {
	if i := strings.LastIndex(f, "/"); i >= 0 {
		f = f[i+1:]
	} else if i := strings.LastIndex(f, "\\"); i >= 0 {
		f = f[i+1:]
	}
	return fmt.Sprintf("%s:%d", f, l)
}

type treePrinterNode struct {
	meta     []string
	value    string
	children []*treePrinterNode
}

func (n *treePrinterNode) Add(meta []interface{}, value interface{}) TreePrinterNode {
	strs := make([]string, len(meta))
	for i, v := range meta {
		if s, ok := v.(string); ok {
			strs[i] = s
		} else {
			strs[i] = fmt.Sprint(v)
		}
	}
	strVal := fmt.Sprint(value)

	// first call initializes the root
	if n.meta == nil && n.value == "" && n.children == nil {
		n.meta = strs
		n.value = strVal
		return n
	}

	child := &treePrinterNode{
		meta:  strs,
		value: strVal,
	}
	n.children = append(n.children, child)

	return child
}

func (n *treePrinterNode) Bytes() []byte {
	if n.meta == nil && n.value == "" && n.children == nil {
		return []byte("<nil>")
	}
	var sizes []int
	n.measure(&sizes)
	var b bytes.Buffer
	n.print(&b, "", sizes)
	out := b.Bytes()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out
}

func (n *treePrinterNode) measure(sizes *[]int) {
	if len(n.meta) > len(*sizes) {
		newSizes := make([]int, len(n.meta))
		copy(newSizes, *sizes)
		*sizes = newSizes
	}
	for i, s := range n.meta {
		if l := len(s); l > (*sizes)[i] {
			(*sizes)[i] = l
		}
	}
	for _, child := range n.children {
		child.measure(sizes)
	}
}

func (n *treePrinterNode) print(b *bytes.Buffer, prefix string, sizes []int) {
	// print meta
	// treeprint style: [meta1 meta2]  Value
	b.WriteByte('[')
	for i, s := range n.meta {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(s)
		if i < len(sizes) {
			pad := sizes[i] - len(s)
			if pad > 0 {
				b.Write(bytes.Repeat([]byte{' '}, pad))
			}
		}
	}
	b.WriteByte(']')

	// print value
	if n.value != "" {
		b.WriteString("  ") // two spaces
		lines := strings.Split(n.value, "\n")
		b.WriteString(lines[0])
		for i := 1; i < len(lines); i++ {
			b.WriteByte('\n')
			b.WriteString(prefix)
			// Indent subsequent lines to align with the rest of the node's content block
			// The visuals are: prefix + connector (4 chars: "├── " or "└── ")
			// So we need 4 spaces to align under the connector.
			b.WriteString("    ")
			b.WriteString(lines[i])
		}
	}

	b.WriteByte('\n')

	// print children
	for i, child := range n.children {
		last := i == len(n.children)-1
		b.WriteString(prefix)
		var newPrefix string
		if last {
			b.WriteString("└── ")
			newPrefix = prefix + "    "
		} else {
			b.WriteString("├── ")
			newPrefix = prefix + "│   "
		}
		child.print(b, newPrefix, sizes)
	}
}

// DefaultPrinterFormatter is used by DefaultPrinter
func DefaultPrinterFormatter() TreePrinterNode { return new(treePrinterNode) }
