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
)

// need a channel based mutex to control access to source
type mutex chan struct{}

func newMutex() mutex {
	cc := make(chan struct{}, 1)
	return cc
}

func (m mutex) Lock() {
	m <- struct{}{}
}

func (m mutex) Unlock() {
	_, ok := <-m
	if !ok {
		panic("mutex already unlocked")
	}
}

const (
	BLOCK_SIZE     = 1 << 15
	CHANNEL_LENGTH = 1 << 10
)

var ErrClosedReader = errors.New("closed multireader")

type entry struct {
	err error
	bs  []byte
}

// MultiplexReader is a structure that allows replication of a source reader to many sink readers
type MultiplexReader struct {
	blocksize int
	rdr       io.Reader
	mtx       mutex
	cs        map[chan entry]chan entry
}

// NewMultiplexReader creates a new source reader
func NewMultiplexReader(r io.Reader) *MultiplexReader {
	return NewMultiplexReaderWithSize(r, BLOCK_SIZE)
}

// NewMultiplexReaderWithSize creates a new source reader that buffers in blocks of `size` bytes.
func NewMultiplexReaderWithSize(r io.Reader, size int) *MultiplexReader {
	q := &MultiplexReader{
		blocksize: size,
		rdr:       r,
		mtx:       newMutex(),
		cs:        map[chan entry]chan entry{},
	}
	return q
}

// Reader is a sink reader.  NewReader creates new reader sinks from a MultiplexReeader source.
type Reader struct {
	mr     *MultiplexReader
	c      chan entry
	buf    []byte
	closed bool
	err    error
}

// NewReader creates a new sink Reader from a MultiplexReader source
func (mr *MultiplexReader) NewReader() *Reader {
	return mr.NewReaderWithLength(CHANNEL_LENGTH)
}

// NewReaderWithLength creates a new sink Reader with the specified channel length.
// channel lenght must be greater than zero or the reader will deadlock on read.
func (mr *MultiplexReader) NewReaderWithLength(length int) *Reader {
	q := &Reader{
		mr:  mr,
		c:   make(chan entry, length),
		buf: []byte{},
	}
	mr.mtx.Lock()
	defer mr.mtx.Unlock()
	mr.cs[q.c] = q.c
	return q
}

// Close the sink.  Readers must be closed when read operations are completed.
// Calling Close releases the memory consumed by the channel of buffers and
// unblocks any calls blocked on this sink.
func (r *Reader) Close() error {
	return r.CloseWithError(nil)
}

func (r *Reader) remove() {
	delete(r.mr.cs, r.c)
	r.closed = true
	r.buf = nil
}

// CloseWithError closes the reader with the supplied error.  See Close.
func (r *Reader) CloseWithError(err error) error {
	mr := r.mr
	close(r.c)
	mr.mtx.Lock()
	defer mr.mtx.Unlock()
	r.remove()
	r.err = err
	if err == nil {
		r.err = ErrClosedReader
		return nil
	}
	return err
}

// Read fulfills the io.Reader interface
func (r *Reader) Read(bs []byte) (nn int, err error) {
	l := len(r.buf)
	if l == 0 {
		if r.err != nil {
			return 0, r.err
		}
		mr := r.mr
		var ent entry
		select {
		case ent = <-r.c:
			// channel is non-nil and has a buffer on it - read it
		case mr.mtx <- struct{}{}:
			// lock and read from the source, distribute to the sinks
			defer mr.mtx.Unlock()
			select {
			default:
				// nothing available on channel so copy new bytes in from reader
				// and redistribute to all other readers.  channel is empty here
				brs := make([]byte, mr.blocksize)
				nn = 0
				if r.closed {
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
				for c := range mr.cs {
					func() {
						defer func() { recover() }()
						c <- entry{
							bs:  brs,
							err: err,
						}
					}()
				}
				ent, _ = <-r.c
			case ent = <-r.c:
				// protect against having items in the channel.  this should be a rare visit
			}
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
