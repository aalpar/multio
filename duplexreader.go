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

type entry struct {
	err error
	bs  []byte
}

type DuplexMultiReader struct {
	blocksize int
	rdr       io.Reader
	mtx       sync.Mutex
	cs        []chan entry
}

type DuplexReader struct {
	mr  *DuplexMultiReader
	c   chan entry
	buf []byte
	err error
}

// NewDuplexReader create a reader than fans read results out between many readers
func NewDuplexMultiReader(r io.Reader) *DuplexMultiReader {
	return NewDuplexMultiReaderWithSize(r, BLOCK_SIZE)
}

// NewDuplexReaderWithSize create a reader than fans read results out between many readers.  size specifies the size of the read buffer.
func NewDuplexMultiReaderWithSize(r io.Reader, size int) *DuplexMultiReader {
	q := &DuplexMultiReader{
		blocksize: size,
		rdr:       r,
		cs:        []chan entry{},
	}
	return q
}

// NewReader create a new reader from the MultiDuplexReader
func (mr *DuplexMultiReader) NewReader() *DuplexReader {
	return mr.NewReaderWithLength(CHANNEL_LENGTH)
}

// NewReaderWithLength create a new reader from the MultiDuplexReader with the specified channel length.
func (mr *DuplexMultiReader) NewReaderWithLength(length int) *DuplexReader {
	q := &DuplexReader{
		mr:  mr,
		c:   make(chan entry, length),
		buf: []byte{},
	}
	mr.mtx.Lock()
	defer mr.mtx.Unlock()
	mr.cs = append(mr.cs, q.c)
	return q
}

func (r *DuplexReader) HasMultiReader() bool {
	return r.mr != nil
}

// Read fulfills the io.Reader interface
func (r *DuplexReader) Read(bs []byte) (nn int, err error) {
	l := len(r.buf)
	mr := r.mr
	if l == 0 {
		select {
		case ent := <-r.c:
			// simple ... buffer len == 0 so copy a buffer and error from
			// the channel
			r.buf = ent.bs
			r.err = ent.err
		default:
			// nothing available on channel so copy new bytes in from reader
			// and redistribute to all other reader
			mr.mtx.Lock()
			brs := make([]byte, mr.blocksize)
			nn = 0
			for nn < mr.blocksize && err == nil {
				n := 0
				n, err = mr.rdr.Read(brs[nn:])
				nn += n
			}
			brs = brs[:nn]
			// brs now has the new bytes and err is any error resulting from the last read
			// distribute the new buffer and error to all the other readers
			for _, c := range mr.cs {
				if c != r.c {
					c <- entry{
						bs:  brs,
						err: err,
					}
				}
			}
			mr.mtx.Unlock()
			// set the local buffer
			r.buf = brs
			r.err = err
		}
	}
	nn = copy(bs, r.buf)
	r.buf = r.buf[nn:]
	return nn, r.err
}
