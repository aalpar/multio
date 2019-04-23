// MIT License
//
// Copyright (c) 2019 Aaron H. Alpar
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
package multio

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestWrite1(t *testing.T) {
	mr := NewMultiReader(MultiReaderBlockSize(5))
	r := mr.NewReader()
	n, err := mr.Write([]byte("testing now."))
	if err != nil {
		t.Fatalf("unexpected value")
	}
	if n != len("testing now.") {
		t.Fatalf("unexpected value")
	}
	if len(r.Chan) != 2 {
		t.Fatalf("unexpected value")
	}
	var bs []byte = <-r.Chan
	if len(bs) != 5 {
		t.Fatalf("unexpected value")
	}
	if !reflect.DeepEqual(bs, []byte("testi")) {
		t.Fatalf("unexpected value")
	}
	bs = <-r.Chan
	if len(bs) != 5 {
		t.Fatalf("unexpected value")
	}
	if !reflect.DeepEqual(bs, []byte("ng no")) {
		t.Fatalf("unexpected value")
	}
	n, err = mr.Write([]byte(" testing again."))
	if err != nil {
		t.Fatalf("unexpected value")
	}
	if n != len(" testing again.") {
		t.Fatalf("unexpected value")
	}
}

func TestWrite2(t *testing.T) {
	mr := NewMultiReader()
	r := mr.NewReader()
	n, err := mr.Write([]byte("testing now."))
	if err != nil {
		t.Fatalf("unexpected value")
	}
	if n != len("testing now.") {
		t.Fatalf("unexpected value")
	}
	if len(r.Chan) != 0 {
		t.Fatalf("unexpected value")
	}
	if len(mr.Buf) != len("testing now.") {
		t.Fatalf("unexpected value")
	}
	n, err = mr.Write([]byte(" testing again."))
	if err != nil {
		t.Fatalf("unexpected value")
	}
	if n != len(" testing again.") {
		t.Fatalf("unexpected value")
	}
	if len(r.Chan) != 0 {
		t.Fatalf("unexpected value")
	}
	if len(mr.Buf) != len("testing now. testing again.") {
		t.Fatalf("unexpected value")
	}
}

func TestRead(t *testing.T) {
	mr := NewMultiReader()
	r := mr.NewReader()
	mr.Write([]byte("testing now."))
	mr.Write([]byte(" testing again."))
	mr.Close()
	bs := make([]byte, 10)
	n, err := r.Read(bs)
	if err != nil {
		t.Fatalf("unexpected value: %v", err)
	}
	if n != 10 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "testing no" {
		t.Fatalf("unexpected value")
	}
	n, err = r.Read(bs)
	if err != nil {
		t.Fatalf("unexpected value")
	}
	if n != 10 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "w. testing" {
		t.Fatalf("unexpected value")
	}
	n, err = r.Read(bs)
	if err != io.EOF {
		t.Fatalf("unexpected value")
	}
	if n != 7 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != " again." {
		t.Fatalf("unexpected value")
	}

	mr = NewMultiReader(MultiReaderBlockSize(5))
	r = mr.NewReader()
	mr.Write([]byte("testing now."))
	mr.Write([]byte(" testing again."))
	mr.Close()
	buf := &bytes.Buffer{}
	n0, err0 := io.Copy(buf, r)
	if err0 != nil {
		t.Fatalf("unexpected value")
	}
	if n0 != 27 {
		t.Fatalf("unexpected value: %v %v", len(mr.Chans[0]), mr)
	}
	if buf.String() != "testing now. testing again." {
		t.Fatalf("unexpected value: %q", buf.String())
	}

}

func TestReadWrite(t *testing.T) {
	mr := NewMultiReader()
	r0 := mr.NewReader()
	r1 := mr.NewReader()
	go func() {
		mr.Write([]byte("testing now."))
		mr.Write([]byte(" testing again."))
		mr.Flush()
		mr.Close()
	}()
	buf0 := &bytes.Buffer{}
	n0, err0 := io.Copy(buf0, r0)
	if err0 != nil {
		t.Fatalf("unexpected value")
	}
	if n0 != 27 {
		t.Fatalf("unexpected value")
	}
	buf1 := &bytes.Buffer{}
	n1, err1 := io.Copy(buf1, r1)
	if err1 != nil {
		t.Fatalf("unexpected value")
	}
	if n1 != 27 {
		t.Fatalf("unexpected value")
	}

}

func BenchmarkSetup(b *testing.B) {
	mr := NewMultiReader()
	q := []*Reader{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q = append(q, mr.NewReader())
	}
	go func() {
		mr.Write([]byte("testing now."))
		mr.Write([]byte(" testing again."))
		mr.Flush()
		mr.Close()
	}()

}

func BenchmarkReadWrite(b *testing.B) {
	mr := NewMultiReader()
	q := []*Reader{}
	for i := 0; i < b.N; i++ {
		q = append(q, mr.NewReader())
	}
	b.ResetTimer()
	go func() {
		mr.Write([]byte("testing now."))
		mr.Write([]byte(" testing again."))
		mr.Flush()
		mr.Close()
	}()
	for i := 0; i < b.N; i++ {
		n, err := io.Copy(&bytes.Buffer{}, q[i])
		if err != nil {
			b.Fatalf("unexpected value")
		}
		if n != 27 {
			b.Fatalf("unexpected value")
		}
	}

}
