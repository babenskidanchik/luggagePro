package otpdelivery

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type DevSender struct {
	path string
	mu   sync.Mutex
}

func NewDevSender(path string) *DevSender {
	return &DevSender{
		path: path,
	}
}

func (s *DevSender) SendOTP(
	ctx context.Context,
	phone string,
	code string,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.path)

	if err := os.MkdirAll(
		dir,
		0700,
	); err != nil {
		return fmt.Errorf(
			"create dev OTP directory: %w",
			err,
		)
	}

	if err := os.WriteFile(
		s.path,
		[]byte(code+"\n"),
		0600,
	); err != nil {
		return fmt.Errorf(
			"write dev OTP: %w",
			err,
		)
	}

	return nil
}
