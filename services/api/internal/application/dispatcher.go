package application

import (
	"context"
	"hms.local/api/internal/domain"
	"time"
)

type DispatchRepository interface {
	RecoverMessages(context.Context) (int64, error)
	ClaimMessage(context.Context) (domain.MessageLease, error)
	RenewMessage(context.Context, domain.MessageLease) error
	FinishMessage(context.Context, domain.MessageLease, string, string) error
}
type MessageTransport interface {
	Send(context.Context, domain.MessageLease) (string, string, error)
}
type Dispatcher struct {
	Store     DispatchRepository
	Transport MessageTransport
}

func (d Dispatcher) Process(ctx context.Context) error {
	if _, err := d.Store.RecoverMessages(ctx); err != nil {
		return err
	}
	m, err := d.Store.ClaimMessage(ctx)
	if err != nil {
		return err
	}
	sendCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-sendCtx.Done():
				return
			case <-ticker.C:
				renewCtx, stop := context.WithTimeout(sendCtx, 3*time.Second)
				err := d.Store.RenewMessage(renewCtx, m)
				stop()
				if err != nil {
					cancel()
					return
				}
			}
		}
	}()
	status, provider, sendErr := d.Transport.Send(sendCtx, m)
	close(done)
	<-stopped
	// Unknown outcomes are never converted into a retriable pending message.
	if status != "sent" && status != "captured" && status != "failed" && status != "uncertain" {
		status = "uncertain"
	}
	finishCtx, finishCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer finishCancel()
	if err = d.Store.FinishMessage(finishCtx, m, status, provider); err != nil {
		return err
	}
	return sendErr
}
