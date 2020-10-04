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
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

// ExampleFile basic example that shows replication of a random stream to two
// files
func ExampleFile() {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	rdr := NewMultiplexReader(io.LimitReader(r, 1<<24))

	f0, err := os.Create("example1.dat")
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	f1, err := os.Create("example2.dat")
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	// all readers must be created before first read
	rdr0 := rdr.NewReader()
	rdr1 := rdr.NewReader()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		n, err := io.Copy(f0, rdr0)
		fmt.Printf("copied %d\n", n)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		wg.Done()
	}()

	go func() {
		n, err := io.Copy(f1, rdr1)
		fmt.Printf("copied %d\n", n)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
		wg.Done()
	}()

	wg.Wait()

	//Output:
	//copied 16777216
	//copied 16777216

}


// ExampleFileWithJam example shows freeing readers blocked by fill buffer channels.
func ExampleFileWithJam() {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	// 1<<26 is beyond the number of bytes that the default channel sizes can hold
	rdr := NewMultiplexReader(io.LimitReader(r, 1<<26))

	f0, err := os.Create("example1.dat")
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}
	f1, err := os.Create("example2.dat")
	if err != nil {
		log.Fatalf("err: %v\n", err)
	}

	// all readers must be created before first read
	rdr0 := rdr.NewReader()
	rdr1 := rdr.NewReader()

	wg := sync.WaitGroup{}
	wg.Add(2)

	// wait for 2 seconds before closing the blocked reader.
	ctx, canfn := context.WithTimeout(context.Background(), time.Second*2)
	defer canfn()

	go func() {
		defer wg.Done()
		// this will one buffer block from the channel then allow the rest of the
		// channel to fill waiting for a read
		n, err := io.Copy(f0, io.LimitReader(rdr0, default_BLOCK_SIZE_B))
		fmt.Printf("copied %d\n", n)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}()

	go func() {
		defer wg.Done()
		n, err := io.Copy(f1, rdr1)
		fmt.Printf("copied %d\n", n)
		if err != nil {
			fmt.Printf("err: %v\n", err)
		}
	}()

	// wait for timeout
	select {
	case <-ctx.Done():
		// close the reader causing the block.  if a block occurs, the reader
		// with the longest channel length will be causing the block.
		// for this example, always reader 0
		if rdr0.Len() != 0 {
			rdr0.Close()
		} else if rdr1.Len() != 0 {
			rdr1.Close()
		}
	}

	// can only get here if both wg.Done()s have been visited
	wg.Wait()

	//Output:
	//copied 32768
	//copied 67108864

}
