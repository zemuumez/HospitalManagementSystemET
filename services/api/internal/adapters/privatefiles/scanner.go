package privatefiles

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"hms.local/api/internal/domain"
	"io"
	"net"
	"strings"
	"time"
)

// ClamAV uses an authenticated local boundary supplied by deployment, not by
// the protocol. Only loopback addresses are accepted here.
type Scanner struct{ Address string }

func (s Scanner) Scan(ctx context.Context, data []byte) error {
	host, _, err := net.SplitHostPort(s.Address)
	ip := net.ParseIP(host)
	if err != nil || ip == nil || !ip.IsLoopback() {
		return domain.ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", s.Address)
	if err != nil {
		return domain.ErrUnavailable
	}
	defer conn.Close()
	deadline, _ := ctx.Deadline()
	if err = conn.SetDeadline(deadline); err != nil {
		return domain.ErrUnavailable
	}
	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()
	if _, err = io.Copy(conn, strings.NewReader("zINSTREAM\x00")); err != nil {
		return domain.ErrUnavailable
	}
	for len(data) > 0 {
		n := min(len(data), 64<<10)
		var head [4]byte
		binary.BigEndian.PutUint32(head[:], uint32(n))
		if _, err = io.Copy(conn, bytes.NewReader(head[:])); err != nil {
			return domain.ErrUnavailable
		}
		if _, err = io.Copy(conn, bytes.NewReader(data[:n])); err != nil {
			return domain.ErrUnavailable
		}
		data = data[n:]
	}
	if _, err = conn.Write([]byte{0, 0, 0, 0}); err != nil {
		return domain.ErrUnavailable
	}
	reply, err := bufio.NewReader(io.LimitReader(conn, 4096)).ReadString(0)
	if err != nil {
		return domain.ErrUnavailable
	}
	if reply == "stream: OK\x00" {
		return nil
	}
	if strings.HasSuffix(reply, " FOUND\x00") {
		return domain.ErrValidation
	}
	return domain.ErrUnavailable
}
