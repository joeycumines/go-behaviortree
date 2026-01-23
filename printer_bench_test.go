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
	"io"
	"testing"
)

func BenchmarkNode_String(b *testing.B) {
	node := New(Sequence)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = node.String()
	}
}

func BenchmarkTreePrinter_Fprint_simple(b *testing.B) {
	node := New(Sequence)
	p := TreePrinter{
		Inspector: DefaultPrinterInspector,
		Formatter: DefaultPrinterFormatter,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = p.Fprint(io.Discard, node)
	}
}

func BenchmarkTreePrinter_Fprint_nested(b *testing.B) {
	// create a tree with some depth and children
	// root -> (seq, sel)
	// seq -> (seq, seq)
	// sel -> (seq, seq)
	node := New(
		Sequence,
		New(
			Selector,
			New(Sequence),
			New(Sequence),
		),
		New(
			Selector,
			New(Sequence),
			New(Sequence),
		),
	)
	p := TreePrinter{
		Inspector: DefaultPrinterInspector,
		Formatter: DefaultPrinterFormatter,
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = p.Fprint(io.Discard, node)
	}
}

func BenchmarkDefaultPrinterInspector(b *testing.B) {
	node := New(Sequence)
	tick := Sequence
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = DefaultPrinterInspector(node, tick)
	}
}
