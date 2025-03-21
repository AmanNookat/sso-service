package main

import (
	"log/slog"             // Новый логгер из стандартной библиотеки Go (Go 1.21+)
	"os"                   // Работа с операционной системой (файлы, переменные окружения и сигналы)
	"os/signal"            // Обработчик системных сигналов (например, завершение программы)
	app "sso/internal/app" // Импортируем пакет с логикой gRPC-сервера
	"sso/internal/config"  // Импортируем конфигурационный пакет
	"syscall"              // Используется для перехвата системных сигналов (SIGTERM, SIGINT)
)

// Константы, определяющие окружение
const (
	envLocal = "local" // Локальная среда (например, разработка)
	envDev   = "dev"   // Среда разработки
	envProd  = "prod"  // Продакшен-среда
)

func main() {
	// Загружаем конфигурацию из config.yaml
	cfg := config.MustLoad()

	// Настраиваем логгер в зависимости от окружения
	log := setupLogger(cfg.Env)

	// Логируем запуск приложения с загруженными настройками
	log.Info("starting application", slog.Any("config", cfg))

	// Создаем новый экземпляр gRPC-приложения
	application := app.New(log, cfg.GRPC.Port, cfg.StoragePath, cfg.TokenTTL)

	// Запускаем gRPC-сервер в отдельной горутине
	go application.GRPCSrv.MustRun()

	// Создаем канал для обработки системных сигналов (SIGINT, SIGTERM)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// Блокируем выполнение и ждем получения сигнала на остановку
	sign := <-stop

	// Логируем, что приложение завершает работу
	log.Info("stopping application", slog.String("signal", sign.String()))

	// Останавливаем gRPC-сервер
	application.GRPCSrv.Stop()

	// Логируем завершение работы
	log.Info("application stopped")
}

// setupLogger - настраивает логгер в зависимости от окружения
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		// Локальная среда: текстовый лог с DEBUG уровнем
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		// Среда разработки: JSON-лог с DEBUG уровнем
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		// Продакшен: JSON-лог с INFO уровнем (не логируем DEBUG)
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
