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
	"errors"
	"io"
	"sync"
)

const (
	BLOCK_SIZE     = 1 << 15
	CHANNEL_LENGTH = 1 << 10
)

var ErrClosedMultiReader = errors.New("closed multireader")

type entry struct {
	err error
	bs  []byte
}

// DuplexMultiReader is a structure that allows multiple readers to all read the same data from a single reader.
type DuplexMultiReader struct {
	blocksize int
	rdr       io.Reader
	mtx       sync.Mutex
	cs        map[chan entry]chan entry
}

// NewDuplexMultiReader create a reader than fans reads to many readers.
func NewDuplexMultiReader(r io.Reader) *DuplexMultiReader {
	return NewDuplexMultiReaderWithSize(r, BLOCK_SIZE)
}

// NewDuplexMultiReaderWithSize create a reader than fans read data to many readers.  Size specifies the size of the read buffer.
func NewDuplexMultiReaderWithSize(r io.Reader, size int) *DuplexMultiReader {
	q := &DuplexMultiReader{
		blocksize: size,
		rdr:       r,
		cs:        map[chan entry]chan entry{},
	}
	return q
}

type DuplexReader struct {
	mr     *DuplexMultiReader
	c      chan entry
	buf    []byte
	closed bool
	err    error
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
	mr.cs[q.c] = q.c
	return q
}

func (r *DuplexReader) Close() error {
	return r.CloseWithError(nil)
}

func (r *DuplexReader) CloseWithError(err error) error {
	mr := r.mr
	close(r.c)
	mr.mtx.Lock()
	defer mr.mtx.Unlock()
	delete(mr.cs, r.c)
	r.closed = true
	r.buf = nil
	r.err = err
	if err == nil {
		r.err = ErrClosedMultiReader
		return nil
	}
	return err
}

// Read fulfills the io.Reader interface
func (r *DuplexReader) Read(bs []byte) (nn int, err error) {
	l := len(r.buf)
	if l == 0 {
		if r.err != nil {
			return l, r.err
		}
		var ent entry
		select {
		// channel is non-nil and has a buffer on it - read it
		case ent = <-r.c:
		default:
			// nothing available on channel so copy new bytes in from reader
			// and redistribute to all other readers
			mr := r.mr
			brs := make([]byte, mr.blocksize)
			nn = 0
			mr.mtx.Lock()
			if r.closed {
				mr.mtx.Unlock()
				return 0, r.err
			}
			// fill the buffer until error or blocksize
			for nn < mr.blocksize && err == nil {
				n := 0
				n, err = mr.rdr.Read(brs[nn:])
				nn += n
			}
			brs = brs[:nn]
			// brs now has the new bytes and err is any error resulting from the last read
			// distribute the new buffer and error to all the other readers
			for _, c := range mr.cs {
				func() {
					defer func() { recover() }()
					c <- entry{
						bs:  brs,
						err: err,
					}
				}()
			}
			mr.mtx.Unlock()
			// empty entry here when channel is closed
			ent, _ = <-r.c
		}
		r.buf = ent.bs
		r.err = ent.err
	}
	// copy channel buffer to the read destination and move the channel buffer forward
	// return error associated with the buffer only after the last read
	nn = copy(bs, r.buf)
	r.buf = r.buf[nn:]
	return nn, nil
}
