package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/services/auth"
	"sso/internal/storage/sqlite"
	"time"
)

// основная структура приложения
type App struct {
	GRPCSrv *grpcapp.App
}

// конструктор
func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	// инициализация хранилище (подключаемся)
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	// инициализация сервиса авторизации
	authService := auth.New(log, storage, storage, storage, tokenTTL)

	// инициализация grpc сервиса 
	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
