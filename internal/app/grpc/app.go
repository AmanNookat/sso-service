package grpcapp

import (
	"fmt"
	"log/slog"
	"net"
	authgrpc "sso/internal/grpc/auth"

	"google.golang.org/grpc"
)

// App - структура, представляющая приложение
type App struct {
	log        *slog.Logger  // Логгер для записи событий
	gRPCServer *grpc.Server  // Экземпляр gRPC-сервера
	port       int           // Порт, на котором работает gRPC-сервер
}

// New - функция-конструктор для создания нового экземпляра App
func New(log *slog.Logger, authService authgrpc.Auth, port int) *App {
	// Создаем новый gRPC-сервер
	gRPCServer := grpc.NewServer()

	// Регистрируем сервис аутентификации в gRPC-сервере
	authgrpc.RegisterAuthServer(gRPCServer, authService)

	// Возвращаем экземпляр App со всеми необходимыми полями
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// MustRun - запускает gRPC-сервер и в случае ошибки завершает программу
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err) // Если `Run` вернет ошибку, программа завершится с паникой
	}
}

// Run - запускает gRPC-сервер и слушает входящие соединения
func (a *App) Run() error {
	const op = "grpcapp.Run" // Название операции для логирования

	// Создаем логгер с контекстной информацией (название операции + порт)
	log := a.log.With(slog.String("op", op), slog.Int("port", a.port))

	// Открываем TCP-соединение и слушаем входящие gRPC-запросы на указанном порту
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err) // Возвращаем ошибку, если порт не удалось открыть
	}

	// Логируем успешный запуск gRPC-сервера
	log.Info("gRPC server is running", slog.String("addr", l.Addr().String()))

	// Запускаем gRPC-сервер и начинаем обработку запросов
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err) // Возвращаем ошибку, если сервер не смог запуститься
	}

	return nil // Если сервер запустился без ошибок, возвращаем `nil`
}

// Stop - останавливает gRPC-сервер
func (a *App) Stop() {
	const op = "grpcapp.Stop" // Название операции для логирования

	// Логируем остановку сервера
	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	// Выполняем Graceful Shutdown (завершаем все активные соединения перед остановкой)
	a.gRPCServer.GracefulStop()
}
