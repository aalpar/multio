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
	default_BLOCK_SIZE_B   = 1 << 15
	default_CHANNEL_LENGTH = 1 << 10
)

var ErrClosedReader = errors.New("closed multireader")

type entry struct {
	i   int64
	err error
	bs  []byte
}

// MultiplexReader is a structure that allows replication of a source reader to many sink readers
type MultiplexReader struct {
	blocksizeB int
	rdr        io.Reader
	mtx        mutex
	baseBi     int64
	cs         map[chan entry]chan entry
}

// NewMultiplexReader creates a new source reader
func NewMultiplexReader(r io.Reader) *MultiplexReader {
	return NewMultiplexReaderWithSize(r, default_BLOCK_SIZE_B)
}

// NewMultiplexReaderWithSize creates a new source reader that buffers in blocks of `size` bytes.
func NewMultiplexReaderWithSize(r io.Reader, sizeB int) *MultiplexReader {
	q := &MultiplexReader{
		blocksizeB: sizeB,
		rdr:        r,
		mtx:        newMutex(),
		cs:         map[chan entry]chan entry{},
	}
	return q
}

// Reader is a sink reader.  NewReader creates new reader sinks from a MultiplexReeader source.
type Reader struct {
	mr     *MultiplexReader
	baseBi int64
	nextBi int64
	c      chan entry
	buf    []byte
	closed bool
	err    error
}

// NewReader creates a new sink Reader from a MultiplexReader source
func (mr *MultiplexReader) NewReader() *Reader {
	return mr.NewReaderWithLength(default_CHANNEL_LENGTH)
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
	if mr.baseBi > 0 {
		panic("late start")
	}
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
	// close outside of lock.  this allows readers to break stalls by calling Close()
	close(r.c)
	r.mr.mtx.Lock()
	defer r.mr.mtx.Unlock()
	r.remove()
	r.err = err
	if err == nil {
		r.err = ErrClosedReader
		return nil
	}
	return err
}

func (r *Reader) WriteTo(w io.Writer) (nn int64, err error) {
	for err == nil {
		n := 0
		n, err = r.read(func() (int, error) {
			wn, werr := w.Write(r.buf)
			if wn == len(r.buf) {
				return wn, r.err
			}
			if werr != nil {
				return wn, werr
			}
			return wn, io.ErrShortWrite
		})
		nn += int64(n)
	}
	if err == io.EOF {
		return nn, nil
	}
	return nn, err
}

// fill bs from multiplex reader source or error.
func (mr *MultiplexReader) fill(bs []byte) (nn int, err error) {
	nn = 0
	for nn < len(bs) && err == nil {
		n := 0
		n, err = mr.rdr.Read(bs[nn:])
		nn += n
	}
	return nn, err
}

func (mr *MultiplexReader) distribute(bs []byte, err error) {
	for c := range mr.cs {
		func() {
			defer func() {
				// ignore panics on channel send
				recover()
			}()
			c <- entry{
				i:   mr.baseBi,
				bs:  bs,
				err: err,
			}
		}()
	}
}

// Read fulfills the io.Reader interface
func (r *Reader) Read(bs []byte) (nn int, err error) {
	return r.read(func() (int, error) {
		// copy channel buffer to the read destination and move the channel buffer forward
		// return error associated with the buffer only after the last read
		nn = copy(bs, r.buf)
		return nn, nil
	})
}

// Read fulfills the io.Reader interface
func (r *Reader) read(coutfn func() (int, error)) (nn int, err error) {
	// nothing left in the buffer, go to the channel
	// and get the next buffer if available.
	l := len(r.buf)
	if l == 0 && r.err != nil {
		return 0, r.err
	}
	if l != 0 {
		nn, err = coutfn()
		r.buf = r.buf[nn:]
		r.baseBi += int64(nn)
		r.nextBi = r.baseBi
		return nn, err
	}
	var ent entry
	select {
	case ent = <-r.c:
		// channel is non-nil and has a buffer on it - read it
	case r.mr.mtx <- struct{}{}:
		// lock and read from the source, distribute to the sinks
		defer r.mr.mtx.Unlock()
		select {
		default:
			if r.closed {
				return 0, r.err
			}
			// nothing available on channel so copy new bytes in from reader
			// and redistribute to all other readers.  channel is empty here
			brs := make([]byte, r.mr.blocksizeB)
			// fill the buffer until error or blocksize
			nn, err = r.mr.fill(brs)
			brs = brs[:nn]
			// brs now has the new bytes and err is any error resulting from the last read
			// distribute the new buffer and error to all the other readers
			r.mr.distribute(brs, err)
			r.mr.baseBi += int64(nn)
			ent, _ = <-r.c
		case ent = <-r.c:
			// protect against having items in the channel.  this should be a rare visit
		}
	}
	r.baseBi = ent.i
	r.buf = ent.bs
	r.err = ent.err
	nn, err = coutfn()
	r.buf = r.buf[nn:]
	r.baseBi += int64(nn)
	r.nextBi = r.baseBi
	return nn, err
}
