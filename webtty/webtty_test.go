package webtty

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"sync"
	"testing"
)

func makeInputMessage(message []byte) []byte {
	return append([]byte{byte(Input)}, message...)
}

type biDirectionalPipeEnd struct {
	send *io.PipeWriter
	recv *io.PipeReader
}

func (pe *biDirectionalPipeEnd) Read(p []byte) (n int, err error) {
	return pe.recv.Read(p)
}

func (pe *biDirectionalPipeEnd) Write(p []byte) (n int, err error) {
	return pe.send.Write(p)
}

type biDirectionalPipe struct {
	left  io.ReadWriter
	right io.ReadWriter
}

func newBiDirectionalPipe() *biDirectionalPipe {
	leftRecv, leftSend := io.Pipe()
	rightRecv, rightSend := io.Pipe()
	return &biDirectionalPipe{
		left:  &biDirectionalPipeEnd{leftSend, rightRecv},
		right: &biDirectionalPipeEnd{rightSend, leftRecv},
	}
}

type slaveBiDirectionalPipeEnd struct {
	io.ReadWriter
}

func (sbdpe slaveBiDirectionalPipeEnd) WindowTitleVariables() map[string]interface{} {
	return map[string]interface{}{}
}

func (sbdpe slaveBiDirectionalPipeEnd) ResizeTerminal(columns int, rows int) error {
	return nil
}

func withWebTTY(t *testing.T, f func(tty *WebTTY, master, slave io.ReadWriter)) {
	master := newBiDirectionalPipe()
	slave := newBiDirectionalPipe()

	tty, err := New(master.right, slaveBiDirectionalPipeEnd{slave.left}, WithPermitWrite())
	if err != nil {
		t.Fatalf("Unexpected error from New(): %s", err)
	}

	// start webtty in a goroutine and watch for errors
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		t.Log("Starting WebTTY.Run()")
		defer wg.Done()
		err := tty.Run(ctx)
		if err != nil {
			switch err {
			case context.Canceled:
				// The context is cancelled at the end of the test so ignore this error
				break
			default:
				t.Fatalf("Unexpected error from Run(): %s", err)
			}
		}
	}()

	// Read the initialize message off the master first. If we don't do this, it will block
	buf := make([]byte, 1024)
	_, err = master.left.Read(buf)
	if err != nil {
		t.Fatalf("Could not read initialize message: %s", err)
	}

	f(tty, master.left, slave.right)

	cancel()
	wg.Wait()
}

func TestWriteInputFromMaster(t *testing.T) {
	withWebTTY(t, func(tty *WebTTY, master, slave io.ReadWriter) {
		t.Log("Writing from master to slave")
		message := []byte("foobar")
		n, err := master.Write(makeInputMessage(message)) // write the message from the client side
		if err != nil {
			t.Fatalf("Unexpected error from Write(): %s", err)
		}
		if n != len(message)+1 { // +1 for the Input byte
			t.Fatalf("Write() accepted `%d` for message `%s`", n, message)
		}

		t.Log("Successfully wrote message from master to WebTTY")

		buf := make([]byte, 1024)
		n, err = slave.Read(buf) // read the message from the server side
		t.Logf("Successfully read %d bytes from slave", n)
		if err != nil {
			t.Fatalf("Unexpected error from Read(): %s", err)
		}
		if string(buf[:n]) != string(message) {
			t.Fatalf("Read() returned `%v` for message `%v`", buf[0:n], message)
		}
	})
}

func TestWriteFromSlave(t *testing.T) {
	withWebTTY(t, func(tty *WebTTY, master, slave io.ReadWriter) {
		t.Log("Writing from slave to master")
		message := []byte("0hello\n")  // line buffered canonical mode
		n, err := slave.Write(message) // write the message from the server side
		if err != nil {
			t.Fatalf("Unexpected error from Write(): %s", err)
		}
		if n != len(message) {
			t.Fatalf("Write() accepted `%d` for message `%s`", n, message)
		}

		t.Log("Successfully wrote message from slave to WebTTY")

		buf := make([]byte, 1024)
		n, err = master.Read(buf) // read the message from the client side
		t.Logf("Successfully read %d bytes from master", n)
		if err != nil {
			t.Fatalf("Unexpected error from Read(): %s", err)
		}
		decoded := make([]byte, 1024)
		n, err = base64.StdEncoding.Decode(decoded, buf[1:n])
		if err != nil {
			t.Fatalf("Unexpected error from base64 decode: %s", err)
		}
		if !bytes.Equal(decoded[:n], message) {
			t.Fatalf("Unexpected message from master: `%v`", decoded[:n])
		}

		// decoded := make([]byte, 1024)
		// n, err = base64.StdEncoding.Decode(decoded, buff.bytes())
		// if err != nil {
		// 	t.Fatalf("Unexpected error from Decode(): %s", err)
		// }
		// if !bytes.Equal(decoded[:n], message) {
		// 	t.Fatalf("Unexpected message received: `%s`", decoded[:n])
		// }
	})
}

func TestPing(t *testing.T) {
	withWebTTY(t, func(tty *WebTTY, master, slave io.ReadWriter) {
		t.Log("Sending ping")
		n, err := master.Write([]byte{Ping})
		if err != nil {
			t.Fatalf("Unexpected error from Write(): %s", err)
		}

		buff := make([]byte, 1024)
		n, err = master.Read(buff)
		if err != nil {
			t.Fatalf("Unexpected error from Read(): %s", err)
		}
		if n != 1 && buff[0] != Pong {
			t.Fatalf("Did not receive Pong message back")
		}
	})
}
