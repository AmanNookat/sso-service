package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	"sso/internal/lib/jwt"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Auth - структура сервисного слоя, отвечающая за аутентификацию пользователей.
type AuthService struct {
	log         *slog.Logger    // Логгер для записи информации о работе сервиса.
	usrSaver    UserSaver       // Интерфейс для сохранения пользователей в базе.
	usrProvider UserProvider    // Интерфейс для получения данных о пользователях.
	appProvider AppProvider     // Интерфейс для работы с приложениями (если есть разные приложения, например, web и mobile).
	tokenTTL    time.Duration   // Время жизни токена (JWT, session и т. д.).
}

// UserSaver - интерфейс для сохранения пользователей в хранилище (например, в базе данных).
type UserSaver interface {
	SaveUser(
		ctx context.Context, // Контекст запроса (для отмены, таймаутов и т. д.).
		email string,        // Email пользователя.
		passHash []byte,     // Хеш пароля пользователя.
	) (uid int64, err error) // Возвращает ID созданного пользователя или ошибку.
}

// UserProvider - интерфейс для получения информации о пользователях.
type UserProvider interface {
	User(ctx context.Context, email string) (models.User, error) // Получает пользователя по email.
	IsAdmin(ctx context.Context, userID int64) (bool, error)     // Проверяет, является ли пользователь администратором.
	IsUserExists(ctx context.Context, userID int64) (bool, error)
}

// AppProvider - интерфейс для работы с данными о приложении (если у нас многосервисная архитектура).
type AppProvider interface {
	App(ctx context.Context, appID int) (models.App, error) // Получает информацию о приложении по его ID.
}

// Предопределенные ошибки, которые могут возникнуть в процессе работы с сервисным слоем.
var (
	ErrInvalidCredentials = errors.New("invalid credentials") // Ошибка, если логин/пароль неверные.
	ErrInvalidAppID       = errors.New("invalid app id")      // Ошибка, если передан несуществующий app_id.
	ErrUserExists         = errors.New("user already exists") // Ошибка, если пользователь с таким email уже зарегистрирован.
	ErrUserNotFound       = errors.New("user not found")      // Ошибка, если пользователь не найден.
)

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration) *AuthService {
	return &AuthService{
		usrSaver:    userSaver,
		usrProvider: userProvider,
		log:         log,
		appProvider: appProvider,
		tokenTTL:    tokenTTL,
	}
}

func (a *AuthService) Login(ctx context.Context, email string, password string, appID int) (string, error) {
	const op = "Auth.Login"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email))

	log.Info("attempting to login user")

	user, err := a.usrProvider.User(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			a.log.Warn("user not found", slog.String("error", err.Error()))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		a.log.Error("failed to get user", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.log.Info("invalid credentials", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	app, err := a.appProvider.App(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user logged in successfully")

	token, err := jwt.NewToken(user, app, a.tokenTTL)
	if err != nil {
		a.log.Error("failed to create token", slog.String("error", err.Error()))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (a *AuthService) RegisterNewUser(ctx context.Context, email string, pass string) (int64, error) {
	const op = "auth.RegisterNewUser"

	log := a.log.With(
		slog.String("op", op),
		slog.String("email", email),
	)

	log.Info("registering new user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed ot generate password hash", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := a.usrSaver.SaveUser(ctx, email, passHash)
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			a.log.Warn("user not found", slog.String("error", err.Error()))

			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}

		log.Error("failed ot save user", slog.String("error", err.Error()))

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("user registered")

	return id, nil
}

func (a *AuthService) IsAdmin(ctx context.Context, userID int64) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID))

	log.Info("checking if user is admin")

	isAdmin, err := a.usrProvider.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", slog.String("error", err.Error()))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user is admin", slog.Bool("is_admin", isAdmin))

	return isAdmin, nil
}

func (a *AuthService) IsUserExists(ctx context.Context, userID int64) (bool, error) {
	const op = "Auth.IsUserExists"

	log := a.log.With(
		slog.String("op", op),
		slog.Int64("user_id", userID))

	log.Info("checking if user exists")

	isUserExists, err := a.usrProvider.IsUserExists(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Warn("user not found", slog.String("error", err.Error()))

			return false, fmt.Errorf("%s: %w", op, ErrInvalidAppID)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("checked if user exists", slog.Bool("exists", isUserExists))

	return isUserExists, nil
}