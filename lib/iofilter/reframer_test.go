package iofilter

import (
	"bytes"
	"io"
	"testing"

	. "github.com/warpfork/go-wish"
)

func TestReframer(t *testing.T) {
	t.Run("Indenting writers should DTRT", func(t *testing.T) {
		t.Run("Given a single line", func(t *testing.T) {
			buf := bytes.NewBufferString("msg\n")
			var out bytes.Buffer
			n, err := io.Copy(LineIndentingWriter(&out), buf)
			Wish(t, err, ShouldEqual, nil)
			Wish(t, n, ShouldEqual, int64(4))
			Wish(t, out.String(), ShouldEqual, "\tmsg\n")
		})

		t.Run("Given a couple lines", func(t *testing.T) {
			buf := bytes.NewBufferString("\nwow\ndang\nmsg\n")
			var out bytes.Buffer
			n, err := io.Copy(LineIndentingWriter(&out), buf)
			Wish(t, err, ShouldEqual, nil)
			Wish(t, n, ShouldEqual, int64(1+3+1+4+1+3+1))
			Wish(t, out.String(), ShouldEqual, "\t\n\twow\n\tdang\n\tmsg\n")
		})

		t.Run("Unterminated lines don't get flushed, sorry", func(t *testing.T) {
			buf := bytes.NewBufferString("msg1\nmsg2")
			var out bytes.Buffer
			n, err := io.Copy(LineIndentingWriter(&out), buf)
			Wish(t, err, ShouldEqual, nil)
			Wish(t, n, ShouldEqual, int64(4+1+4))
			Wish(t, out.String(), ShouldEqual, "\tmsg1\n")
		})

		t.Run("Nested indenters work", func(t *testing.T) {
			var out bytes.Buffer
			wr := LineIndentingWriter(&out)
			wr.Write([]byte("msg1\n"))
			LineIndentingWriter(wr).Write([]byte("msg2\n"))
			wr.Write([]byte("msg3\n"))
			Wish(t, out.String(), ShouldEqual, "\tmsg1\n\t\tmsg2\n\tmsg3\n")
		})
	})
}
