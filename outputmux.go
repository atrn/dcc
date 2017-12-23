// dcc - dependency-driven C/C++ compiler front end
//
// Copyright © A.Newman 2015.
//
// This source code is released under version 2 of the  GNU Public License.
// See the file LICENSE for details.
//
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

// OutputMux is an io.Writer multiplexor used to ensure each Writer's
// output is NOT interleaved. The OutputMux NewWriter method returns
// an io.WriteCloser that writes via the OutputMux. Write calls cache
// the output until the writer is closed at which time the OutputMux
// flushes any accumulated output. There are presently no limit to
// the amount od data buffered in the output mux. This should only
// be a problem for C++ (that's a joke...C++ template errors).
//
type OutputMux struct {
	w io.Writer         // output
	c chan outputMuxMsg // input
	s chan struct{}     // stop
}

// The io.Writer that talks to an OutputMux is the write-side of an os.Pipe.
// the read-side of the pipe is read by a separate goroutine that transfers
// lines of text read from the pipe to the central multiplexor. Lines are
// transferred as outputMuxMsg structures.
//
type outputMuxMsg struct {
	r   io.Reader
	s   string
	eof bool
}

// NewOutputMux returns a new OutputMux that will write
// its output to the given io.Writer.
//
func NewOutputMux(w io.Writer) *OutputMux {
	return &OutputMux{
		w: w,
		c: make(chan outputMuxMsg, 100),
		s: make(chan struct{}),
	}
}

// NewWriter returns an io.Writer used to send data to the receiver
// for eventual output.
//
func (om *OutputMux) NewWriter() *os.File {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	go func(r io.Reader) {
		in := bufio.NewScanner(r)
		for in.Scan() {
			om.c <- outputMuxMsg{r, in.Text(), false}
		}
		om.c <- outputMuxMsg{r, "", true}
	}(r)
	return w
}

// Run runs the receiver's input processing loop.
//
func (om *OutputMux) Run() {
	// Each Reader goroutine (and its corresponding Writer) has an
	// slice of strings that acts as a buffer for its output. We
	// don't expect too much output (although the C++ template
	// driven errors can be quite long).
	//
	buffers := make(map[io.Reader][]string)

	flush := func(r io.Reader) {
		for _, s := range buffers[r] {
			fmt.Fprintln(om.w, s)
		}
		buffers[r] = nil
	}

	flushall := func() {
		for r := range buffers {
			flush(r)
		}
	}

	for {
		select {
		case <-om.s:
			flushall()
			om.s <- struct{}{}
			return

		case msg := <-om.c:
			if msg.eof {
				flush(msg.r)
			} else {
				if _, ok := buffers[msg.r]; !ok {
					buffers[msg.r] = make([]string, 0)
				}
				buffers[msg.r] = append(buffers[msg.r], msg.s)
			}
		}
	}
}

// Stop sends a stop signal to the cause the receiver's Run method
// to stop processing.
//
func (om *OutputMux) Stop() {
	om.s <- struct{}{}
	<-om.s
}
