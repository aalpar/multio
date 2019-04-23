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

package multio

import (
	"io"
	"sync"
)

const (
	BLOCK_SIZE     = 1 << 15
	CHANNEL_LENGTH = 1 << 10
)

type MultiReader struct {
	Mutex     sync.Mutex
	BlockSize int
	Chans     []chan []byte
	Buf       []byte
}

type Reader struct {
	Chan chan []byte
	Buf  []byte
}

func NewMultiReader(os ...MultiReaderOption) *MultiReader {
	q := &MultiReader{
		Mutex:     sync.Mutex{},
		BlockSize: BLOCK_SIZE,
		Chans:     []chan []byte{},
		Buf:       nil,
	}
	for _, o := range os {
		o(q)
	}
	return q
}

func (mr *MultiReader) NewReader(os ...ReaderOption) (q *Reader) {
	q = &Reader{Chan: make(chan []byte, CHANNEL_LENGTH)}
	for _, o := range os {
		o(q)
	}
	mr.Chans = append(mr.Chans, q.Chan)
	return q
}

func (mr *MultiReader) Write(bs []byte) (int, error) {
	bslen := len(bs)
	bsi := 0
	for {
		writelen := mr.BlockSize - len(mr.Buf)
		if bslen-bsi < writelen {
			writelen = bslen - bsi
		}
		mr.Buf = append(mr.Buf, bs[bsi:bsi+writelen]...)
		bsi += writelen
		if len(mr.Buf) != mr.BlockSize {
			return bsi, nil
		}
		for _, c := range mr.Chans {
			c <- mr.Buf
		}
		mr.Buf = make([]byte, 0, mr.BlockSize)
	}
}

func (mr *MultiReader) Flush() error {
	if len(mr.Buf) == 0 {
		return nil
	}
	for _, c := range mr.Chans {
		c <- mr.Buf
	}
	mr.Buf = nil
	return nil
}

func (mr *MultiReader) Close() error {
	if len(mr.Buf) == 0 {
		return nil
	}
	for _, c := range mr.Chans {
		c <- mr.Buf
		close(c)
	}
	mr.Buf = nil
	return nil
}

func (r *Reader) Read(bs []byte) (n int, err error) {
	ok := true
	if len(r.Buf) == 0 {
		r.Buf, ok = <-r.Chan
	}
	if !ok {
		return 0, io.EOF
	}
	n = copy(bs, r.Buf)
	r.Buf = r.Buf[n:]
	if len(r.Buf) == 0 && len(r.Chan) == 0 {
		return n, io.EOF
	}
	return n, nil
}

type MultiReaderOption func(*MultiReader)

type ReaderOption func(*Reader)

func MultiReaderBlockSize(n int) MultiReaderOption {
	return func(y *MultiReader) {
		y.BlockSize = n
	}
}
func ReaderChannelLength(n int) ReaderOption {
	return func(y *Reader) {
		y.Chan = make(chan []byte, n)
	}
}
