package server

import (
	"context"
	"domain-scout/internal/ssh"
)

type Service struct{ detector *Detector }

func NewService() *Service { return &Service{detector: NewDetector()} }
func (s *Service) GetInfo(ctx context.Context, ex ssh.Executor) (Info, error) {
	return s.detector.Detect(ctx, ex)
}
