package jwt

import (
	"sso/internal/domain/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func NewToken(user models.User, app models.App, duration time.Duration) (string, error) {
	// Создаем новый JWT токен с методом подписи HMAC-SHA256
	token := jwt.New(jwt.SigningMethodHS256)

	// Получаем map-клеймы токена (ключ-значение)
	claims := token.Claims.(jwt.MapClaims)

	// Добавляем в токен информацию о пользователе и приложении
	claims["uid"] = user.ID           // ID пользователя
	claims["email"] = user.Email      // Email пользователя
	claims["exp"] = time.Now().Add(duration).Unix() // Время истечения токена (в UNIX формате)
	claims["app_id"] = app.ID         // ID приложения

	// Подписываем токен с использованием секрета приложения
	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err // Возвращаем ошибку, если не удалось подписать токен
	}

	// Возвращаем готовый токен
	return tokenString, nil
}
