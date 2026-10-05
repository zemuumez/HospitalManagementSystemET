package privatefiles

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"hms.local/api/internal/domain"
	"io"
	"net"
	"testing"
)

func TestScannerProtocol(t *testing.T) {
	for _, tc := range []struct {
		name, reply string
		want        error
	}{{"clean", "stream: OK\x00", nil}, {"infected", "stream: Synthetic-Signature FOUND\x00", domain.ErrValidation}, {"error", "stream: unavailable ERROR\x00", domain.ErrUnavailable}, {"malformed", "OK\x00", domain.ErrUnavailable}} {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			done := make(chan error, 1)
			go func() {
				c, err := listener.Accept()
				if err != nil {
					done <- err
					return
				}
				defer c.Close()
				r := bufio.NewReader(c)
				command, err := r.ReadString(0)
				if err != nil || command != "zINSTREAM\x00" {
					done <- errors.New("invalid command")
					return
				}
				var received []byte
				for {
					var h [4]byte
					if _, err = io.ReadFull(r, h[:]); err != nil {
						done <- err
						return
					}
					n := binary.BigEndian.Uint32(h[:])
					if n == 0 {
						break
					}
					if n > 64<<10 {
						done <- errors.New("oversized chunk")
						return
					}
					chunk := make([]byte, n)
					if _, err = io.ReadFull(r, chunk); err != nil {
						done <- err
						return
					}
					received = append(received, chunk...)
				}
				if string(received) != "synthetic upload" {
					done <- errors.New("changed bytes")
					return
				}
				_, err = io.WriteString(c, tc.reply)
				done <- err
			}()
			err = (Scanner{Address: listener.Addr().String()}).Scan(context.Background(), []byte("synthetic upload"))
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v got %v", tc.want, err)
			}
			if err = <-done; err != nil {
				t.Fatal(err)
			}
		})
	}
	if err := (Scanner{Address: "192.0.2.1:3310"}).Scan(context.Background(), nil); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatal("nonlocal scanner allowed")
	}
}
