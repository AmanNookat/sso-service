package models

type App struct {
	ID     int
	Name   string
	Secret string // секрет нужен, чтобы подписывать токены и валидировать их в последующем
}
