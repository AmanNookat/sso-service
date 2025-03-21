package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv" // Библиотека для загрузки конфигурации из YAML/ENV
)

// Config - структура, содержащая настройки приложения
type Config struct {
	Env         string        `yaml:"env" env-default:"local"`  // Окружение (local, dev, prod)
	StoragePath string        `yaml:"storage_path" env-required:"true"` // Путь к файлу хранения (например, SQLite)
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"` // Время жизни токена
	GRPC        GRPCConfig    `yaml:"grpc"` // Вложенная структура с настройками gRPC
}

// GRPCConfig - структура с параметрами gRPC
type GRPCConfig struct {
	Port    int           `yaml:"port"`    // Порт gRPC-сервера
	Timeout time.Duration `yaml:"timeout"` // Таймаут gRPC-запросов
}

// MustLoad - загружает конфигурацию из файла, указанного в аргументе `-config` или переменной окружения `CONFIG_PATH`
func MustLoad() *Config {
	path := fetchConfigPath() // Получаем путь к конфигурационному файлу
	if path == "" {
		panic("config path is empty") // Если путь пуст, завершаем выполнение с ошибкой
	}

	// Проверяем, существует ли файл конфигурации
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path) // Если нет файла — ошибка
	}

	var cfg Config

	// Читаем конфигурацию из YAML-файла
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error()) // Если ошибка — завершаем выполнение
	}

	return &cfg // Возвращаем загруженную конфигурацию
}

// MustLoadPath - загружает конфигурацию из указанного пути
func MustLoadPath(configPath string) *Config {
	// Проверяем, существует ли файл конфигурации
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath) // Если файла нет, завершаем выполнение
	}

	var cfg Config

	// Читаем конфигурацию из YAML-файла
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error()) // Если ошибка — завершаем выполнение
	}

	return &cfg // Возвращаем загруженную конфигурацию
}

// fetchConfigPath - определяет путь к файлу конфигурации
func fetchConfigPath() string {
	var res string

	// Читаем путь к конфигу из флага `-config`
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse() // Разбираем аргументы командной строки

	// Если путь не указан через флаг, пробуем прочитать переменную окружения `CONFIG_PATH`
	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res // Возвращаем путь к конфигурации
}
