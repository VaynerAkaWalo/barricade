package healthcheck

type Service struct{}

func (s *Service) IsSystemHealthy() Health {
	return Health{Status: "ok"}
}
