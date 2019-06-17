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
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	NADA        = ``
	SHORT_GREEK = `Lorem ipsum`
	LONG_GREEK  = `Lorem ipsum dolor sit amet, nulla gravida litora nulla sed, ullamcorper sapien molestie consectetuer et amet orci, amet nec non aliquam eu nascetur sem, etiam aliquam pellentesque adipiscing parturient arcu. Auctor lacinia dui, eget nibh pharetra lectus adipiscing leo nulla, ipsa nascetur convallis sit gravida tincidunt. Ante cras vitae massa libero quis vivamus, risus lacus nec nibh eget, neque posuere eros, volutpat id, laoreet diam curabitur. Quam ridiculus dictum cursus ante in ac, dapibus ut vestibulum a commodo pede duis, pede dignissim in tristique nec, nostra urna, adipiscing lobortis. Metus donec vulputate molestie vitae consequat odio, eros vel urna, magna fusce amet elit in. Massa sapien in ultricies ullamco, mauris pellentesque sit. Tortor sed ut vel nulla arcu, arcu pellentesque at diam eget. Dolores et adipiscing ornare ipsum, sapien quam. Sed lacinia turpis integer odio nulla vitae, montes porttitor mollis ante in at porta. Quisque sapien ornare, cras voluptatibus tortor eget sit ac, nulla facilisi ligula. Rhoncus viverra sapien ligula, aliquid ultrices cursus justo tempor velit fringilla, fringilla elit, aenean dolor enim porttitor accumsan, accumsan aenean nulla quam.
Et in donec sed elit, donec arcu mauris libero sodales voluptates arcu, rutrum magna non. Perspiciatis nulla, ligula qui magna mattis phasellus primis, feugiat aliquip ante amet proin, adipiscing quam a tellus in pede, dui ad eros nullam tristique duis. Feugiat pellentesque odio adipiscing. Magna ipsum tincidunt ipsa nec, eros metus odio sit diam. Fusce adipiscing eget mauris, ante vulputate ultricies scelerisque, amet vestibulum vestibulum praesent sollicitudin, quidem massa vitae magna ut vivamus. Non adipiscing fusce diam nullam, unde mauris in, orci lacus sodales a sit proin dictum, quam ac consectetuer dictum condimentum libero, nibh phasellus integer.
Penatibus vulputate, eget purus massa nonummy cras ante, in dignissim mauris natoque tortor orci suspendisse. Ipsum sagittis aliquam, orci sed lectus mauris ac faucibus ut, wisi elit vestibulum amet elit vel. Dignissim amet lacus ullamcorper, orci at sed vestibulum porttitor. Sapien metus commodo ante pulvinar pede, sed congue odio lacus arcu rutrum, vitae dui sed mi in non nisl, eget sociis nunc vitae. Nec morbi nulla duis sem, congue etiam, nisl a metus viverra in, tellus hendrerit felis id suscipit leo. Velit nunc orci sed arcu nec venenatis, parturient nascetur malesuada, aliquet tempus elit, amet vestibulum libero sodales laoreet in. Dictum sem lectus assumenda fringilla sit nunc, est dolor blandit pretium hendrerit mollis mus, vehicula nunc nulla, ligula turpis quis gravida et, facilisis lorem nibh.
Sed mus placerat sagittis ac, pellentesque tellus vitae elementum, non non nisl magna. Volutpat luctus aliquet nisl tortor, etiam libero, id et posuere ut congue dignissim suspendisse. Vel dui vel mattis praesent, in morbi accumsan nascetur ipsum, euismod ac duis semper vel dolor non, possimus viverra mauris wisi nec nec. Maecenas eleifend tortor mollis commodo, felis praesent doloribus. Est cum. Modi cras morbi, suspendisse pellentesque eget nullam ut nam. Parturient proin ornare ante nec lacus, magna vestibulum lorem condimentum, id aenean lectus. Tortor ante mauris est vehicula, ante pede rutrum orci malesuada, nunc vehicula rhoncus aliquam aliquam hac luctus. Ultricies augue id morbi convallis dolor.`
)

func TestNewDuplexMultiReader(t *testing.T) {
	mr := NewDuplexMultiReader(strings.NewReader(SHORT_GREEK))
	r0 := mr.NewReader()
	if mr.cs == nil {
		t.Fatalf("unexpected value")
	}
	if len(mr.cs) != 1 {
		t.Fatalf("unexpected value")
	}
	if r0.buf == nil {
		t.Fatalf("unexpected value")
	}
	if len(r0.buf) > 0 {
		t.Fatalf("unexpected value")
	}
	if r0.mr != mr {
		t.Fatalf("unexpected value")
	}
	r1 := mr.NewReader()
	if mr.cs == nil {
		t.Fatalf("unexpected value")
	}
	if len(mr.cs) != 2 {
		t.Fatalf("unexpected value")
	}
	if r1.mr != mr {
		t.Fatalf("unexpected value")
	}
}

func TestReaderBasicShort(t *testing.T) {
	mr := NewDuplexMultiReaderWithSize(strings.NewReader(SHORT_GREEK), 5)
	r0 := mr.NewReader()
	bs := make([]byte, 10)

	n, err := r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	n, err = r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r0.Read(bs)
	if err != io.EOF {
		t.Fatalf("err: %v", err)
	}
	if n != 1 {
		t.Fatalf("unexpected value: %d", n)
	}
	if string(bs[:n]) != "m" {
		t.Fatalf("unexpected value")
	}

}

func TestReaderMiultiShort(t *testing.T) {
	mr := NewDuplexMultiReaderWithSize(strings.NewReader(SHORT_GREEK), 5)
	r0 := mr.NewReader()
	r1 := mr.NewReader()
	bs := make([]byte, 10)

	n, err := r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != io.EOF {
		t.Fatalf("err: %v", err)
	}
	if n != 1 {
		t.Fatalf("unexpected value: %d", n)
	}
	if string(bs[:n]) != "m" {
		t.Fatalf("unexpected value")
	}

	n, err = r0.Read(bs)
	if err != io.EOF {
		t.Fatalf("err: %v", err)
	}
	if n != 1 {
		t.Fatalf("unexpected value: %d", n)
	}
	if string(bs[:n]) != "m" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err == nil {
		t.Fatalf("unexpected value: %v", err)
	}
	if n != 0 {
		t.Fatalf("unexpected value: %d", n)
	}

	n, err = r0.Read(bs)
	if err == nil {
		t.Fatalf("unexpected value: %v", err)
	}
	if n != 0 {
		t.Fatalf("unexpected value: %d", n)
	}

}

func TestReaderMiultiShort2(t *testing.T) {
	mr := NewDuplexMultiReaderWithSize(strings.NewReader(SHORT_GREEK), 2<<8)
	r0 := mr.NewReader()
	r1 := mr.NewReader()
	bs := make([]byte, 5)

	n, err := r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != io.EOF {
		t.Fatalf("err: %v", err)
	}
	if n != 1 {
		t.Fatalf("unexpected value: %d", n)
	}
	if string(bs[:n]) != "m" {
		t.Fatalf("unexpected value")
	}

	n, err = r0.Read(bs)
	if err != io.EOF {
		t.Fatalf("err: %v", err)
	}
	if n != 1 {
		t.Fatalf("unexpected value: %d", n)
	}
	if string(bs[:n]) != "m" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err == nil {
		t.Fatalf("unexpected value: %v", err)
	}
	if n != 0 {
		t.Fatalf("unexpected value: %d", n)
	}

	n, err = r0.Read(bs)
	if err == nil {
		t.Fatalf("unexpected value: %v", err)
	}
	if n != 0 {
		t.Fatalf("unexpected value: %d", n)
	}

}

func TestReaderMiultiShortFile(t *testing.T) {
	f, _ := os.Open("testdata/dat.txt")
	mr := NewDuplexMultiReaderWithSize(f, 5)
	r0 := mr.NewReader()
	r1 := mr.NewReader()
	bs := make([]byte, 10)

	n, err := r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if n != 5 {
		t.Fatalf("unexpected value")
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != io.EOF {
		t.Fatalf("err: %v", err)
	}
	if n != 2 {
		t.Fatalf("unexpected value: %d", n)
	}
	if string(bs[:n]) != "m\n" {
		t.Fatalf("unexpected value: %q", bs[:n])
	}

	n, err = r0.Read(bs)
	if err != io.EOF {
		t.Fatalf("err: %v", err)
	}
	if n != 2 {
		t.Fatalf("unexpected value: %d", n)
	}
	if string(bs[:n]) != "m\n" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err == nil {
		t.Fatalf("unexpected value: %v", err)
	}
	if n != 0 {
		t.Fatalf("unexpected value: %d", n)
	}

	n, err = r0.Read(bs)
	if err == nil {
		t.Fatalf("unexpected value: %v", err)
	}
	if n != 0 {
		t.Fatalf("unexpected value: %d", n)
	}

}

func TestReaderClose(t *testing.T) {
	mr := NewDuplexMultiReaderWithSize(strings.NewReader(LONG_GREEK), 10)
	r0 := mr.NewReader()
	r1 := mr.NewReader()
	bs := make([]byte, 5)

	n, err := r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	n, err = r0.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(bs[:n]) != "Lorem" {
		t.Fatalf("unexpected value")
	}

	r0.Close()

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(bs[:n]) != " ipsu" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(bs[:n]) != "m dol" {
		t.Fatalf("unexpected value")
	}

	n, err = r1.Read(bs)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if string(bs[:n]) != "or si" {
		t.Fatalf("unexpected value: %q", string(bs))
	}

	n, err = r0.Read(bs)
	if err != ErrClosedMultiReader {
		t.Fatalf("err: %v", err)
	}

}

func TestReaderMiultiLopsided(t *testing.T) {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	mr := NewDuplexMultiReaderWithSize(r, 10)

	r0 := mr.NewReaderWithLength(3)
	r1 := mr.NewReaderWithLength(3)

	bs0 := make([]byte, 1<<10)
	bs1 := make([]byte, 1<<10)

	cn := sync.NewCond(&sync.Mutex{})

	go func() {
		n := 0
		for i := 0; i < 4; i++ {
			cn.L.Lock()
			if i == 3 {
				cn.Broadcast()
			}
			cn.L.Unlock()
			p, err := r0.Read(bs0[n:])
			n += p
			if err != nil {
				t.Fatalf("err: %v", err)
			}
		}
	}()

	cn.L.Lock()
	cn.Wait()
	nn, err := r1.Read(bs1)
	cn.L.Unlock()
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if nn != 10 {
		t.Fatalf("unexpected value: %d", nn)
	}
	if !reflect.DeepEqual(bs0[:nn], bs1[:nn]) {
		t.Fatalf("unexpected value")
	}

}

func TestReaderParallel(t *testing.T) {

	BN := 2 << 27 // 128mb

	r := rand.New(rand.NewSource(time.Now().Unix()))
	mr := NewDuplexMultiReader(r)

	N := 80

	rss := make([]*DuplexReader, N)
	bss := make([]*bytes.Buffer, N)

	for i := 0; i < N; i++ {
		rss[i] = mr.NewReader()
		bss[i] = &bytes.Buffer{}
	}

	wg := sync.WaitGroup{}

	for i := 0; i < N; i++ {

		wg.Add(1)

		go func(j int) {
			defer func() { rss[j].Close(); wg.Done() }()
			q := r.Int63n(int64(BN))
			var nn int
			var err error
			var n int64
			bs := make([]byte, q)
			for err == nil && n < q {
				nn, err = rss[j].Read(bs[n:])
				if nn == 0 && err == nil {
					t.Fatalf("unexpected value")
				}
				n += int64(nn)
			}
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if int64(n) != q {
				t.Fatalf("unexpected value")
			}
		}(i)

	}

	wg.Wait()

	li := 0
	for i := 1; i < N; i++ {
		l := len(bss[i].Bytes())
		if l > len(bss[li].Bytes()) {
			li = i
		}
	}

	bs0 := bss[li].Bytes()
	for i := 0; i < N; i++ {
		bs1 := bss[i].Bytes()
		l := len(bs1)
		for j := 0; j < l; j++ {
			if bs0[j] != bs1[j] {
				t.Logf("len(bs0) = %d", len(bs0))
				t.Logf("len(bs1) = %d", len(bs1))
				t.Logf("bs0 = %v", bs0[:j+1])
				t.Logf("bs1 = %v", bs1[:j+1])
				t.Fatalf("unexpected value: %d", j)
			}
		}
	}

}

func BenchmarkRead(b *testing.B) {

	GiB := 1 << 24

	r := rand.New(rand.NewSource(time.Now().Unix()))
	rs := make([]io.Reader, b.N)
	bs := make([]byte, GiB)

	b.ReportAllocs()
	b.ResetTimer()

	mr := NewDuplexMultiReader(r)
	for i := 0; i < b.N; i++ {
		rs[i] = mr.NewReader()
	}

	for i := 0; i < b.N; i++ {
		n := 0
		for n < GiB {
			p, err := rs[i].Read(bs[n:])
			n += p
			if err != nil {
				b.Fatal(err)
			}
		}
	}

}

func BenchmarkThreadedRead(b *testing.B) {

	GiB := 1 << 30

	r := rand.New(rand.NewSource(time.Now().Unix()))
	rs := make([]io.Reader, b.N)

	mr := NewDuplexMultiReader(r)
	for i := 0; i < b.N; i++ {
		rs[i] = mr.NewReader()
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		go func(j int) {
			n, _ := io.CopyN(ioutil.Discard, rs[j], int64(GiB))
			if int(n) != GiB {
				b.Fatalf("unexpected value")
			}
		}(i)
	}

}
