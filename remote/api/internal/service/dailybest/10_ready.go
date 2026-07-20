package dailybest

import "context"

func (s *Service) Ready(ctx context.Context) error {
	return s.repository.Ping(ctx)
}
