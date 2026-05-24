package server

import (
	"backend-skripsi/internal/handler"
	"backend-skripsi/internal/mailer"
	"backend-skripsi/internal/repository"
	"backend-skripsi/internal/security"
	"backend-skripsi/internal/service"
)

type handlers struct {
	auth *handler.AuthHandler
}

func (s *Server) initDependencies() *handlers {
	userRepo := repository.NewUserRepository(s.db)
	authLogRepo := repository.NewAuthLogRepository(s.db)

	s.cacheRepo = repository.NewCacheRepository(s.rdb)

	mailClient := mailer.NewSMTPMailer()

	s.jwtProvider = security.NewJWTProvider()

	authService := service.NewAuthService(userRepo, s.cacheRepo, authLogRepo, mailClient, s.jwtProvider)
	authHdl := handler.NewAuthHandler(authService)

	return &handlers{
		auth: authHdl,
	}
}
