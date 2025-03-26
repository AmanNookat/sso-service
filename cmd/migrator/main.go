package main

import (
	"errors"
	"flag"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var storagePath, migrationsPath, migrationsTable string

	// Чтение флагов командной строки:
	flag.StringVar(&storagePath, "storage-path", "", "path to storage") // путь к файлу БД (например, SQLite)
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations") // путь к папке с миграциями
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table") // таблица, где будут храниться сведения о выполненных миграциях
	flag.Parse()


	// Проверка обязательных параметров
	if storagePath == "" {
		panic("storage-path is required")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}
	

	// Создаём экземпляр мигратора
	m, err := migrate.New(
		"file://"+migrationsPath, // откуда брать миграции
		fmt.Sprintf("sqlite3://%s?x-migrations-table=%s", 
		storagePath, migrationsTable), // к какой БД подключаться
	)
	if err != nil {
		panic(err)
	}


	// Применяем миграции (по порядку)
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply") // если нет новых миграций — просто выходим
			return
		}
		panic(err) // если ошибка — падаем
	}

	fmt.Println("migrations applied") // успешное завершение
}
