package server

import (
	"backend-skripsi/internal/handler"
	"backend-skripsi/internal/mailer"
	"backend-skripsi/internal/repository"
	"backend-skripsi/internal/service"
)

type handlers struct {
	auth *handler.AuthHandler
}

func (s *Server) initDependencies() *handlers {
	userRepo := repository.NewUserRepository(s.db)
	cacheRepo := repository.NewCacheRepository(s.rdb)

	mailClient := mailer.NewSMTPMailer()

	authService := service.NewAuthService(userRepo, cacheRepo, mailClient)
	authHdl := handler.NewAuthHandler(authService)

	return &handlers{
		auth: authHdl,
	}
}
