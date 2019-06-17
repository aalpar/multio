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
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

func Example() {

	r := rand.New(rand.NewSource(time.Now().Unix()))
	rdr := NewDuplexMultiReader(io.LimitReader(r, 1<<30))

	f0, err := os.Create("example1.dat")
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	f1, err := os.Create("example2.dat")
	if err != nil {
		log.Fatalf("err: %v", err)
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
			log.Fatalf("err: %v", err)
		}
		wg.Done()
	}()

	go func() {
		n, err := io.Copy(f1, rdr1)
		fmt.Printf("copied %d\n", n)
		if err != nil {
			log.Fatalf("err: %v", err)
		}
		wg.Done()
	}()

	wg.Wait()

	//Output:
	//copied 1073741824
	//copied 1073741824

}
