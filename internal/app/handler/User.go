package handler

import (
	"encoding/json"
	"errors"
	dto "lab/internal/app/DTO"
	"lab/internal/app/ds"
	"lab/internal/app/role"
	"lab/internal/app/service"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func (h *Handler) RegisterUserHandlers(router *gin.Engine) {
	router.POST("/api/user/registration", h.Registration)
	router.GET("/api/user/:id", h.WithAuthCheck(role.User, role.Moderator), h.GetUserData)
	router.PUT("/api/user", h.WithAuthCheck(role.User, role.Moderator), h.ChangeUserData)
	router.POST("/api/login", h.Login)
	router.POST("/api/logout", h.WithAuthCheck(role.User, role.Moderator), h.Logout)
}

// Registration godoc
// @Summary Регистрация пользователя
// @Description Создает нового пользователя в системе
// @Tags Users
// @Accept json
// @Produce json
// @Param user body dto.UserRegistration true "Данные для регистрации"
// @Success 201 {object} dto.UserDataResposne "Данные созданного пользователя"
// @Failure 400 {object} map[string]string "Некорректные данные или пользователь уже существует"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user/registration [post]
func (h *Handler) Registration(ctx *gin.Context) {
	var user dto.UserRegistration
	var pgErr *pgconn.PgError
	if err := ctx.BindJSON(&user); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest,
			"введен неправильный формат тела запроса")
		return
	}

	response, err := h.Service.AddNewUser(user)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"")
		} else if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			h.errorHandler(ctx, http.StatusBadRequest,
				"пользователь с таким логином уже существует")
		} else if errors.Is(err, service.ErrBadRequest) {
			h.errorHandler(ctx, http.StatusBadRequest, "длина пароля должна быть не менее 8 символов")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

// GetUserData godoc
// @Summary Получить данные пользователя
// @Description Возвращает информацию о пользователе по ID
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID пользователя"
// @Success 200 {object} dto.UserDataResposne "Данные пользователя"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user/{id} [get]
func (h *Handler) GetUserData(ctx *gin.Context) {
	userId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id заявки по солнечным панелям")
		return
	}

	response, err := h.Service.GetUserData(uint(userId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найдено")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, response)

}

// ChangeUserData godoc
// @Summary Изменить данные пользователя
// @Description Обновляет данные текущего авторизованного пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body dto.ChangeUserData true "Обновленные данные пользователя"
// @Success 200 {object} dto.UserDataResposne "Обновленные данные пользователя"
// @Failure 400 {object} map[string]string "Некорректные данные"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /user [put]
func (h *Handler) ChangeUserData(ctx *gin.Context) {
	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}
	var user dto.ChangeUserData
	if err := ctx.BindJSON(&user); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest,
			"введен неправильный формат тела запроса")
		return
	}
	response, err := h.Service.ChangeUserData(user, userId)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// Login godoc
// @Summary Авторизация пользователя
// @Description Выполняет вход в систему и возвращает JWT токен
// @Tags Users
// @Accept json
// @Produce json
// @Param credentials body dto.LoginReq true "Логин и пароль"
// @Success 200 {object} dto.LoginRes "JWT токен и информация о сессии"
// @Failure 400 {object} map[string]string "Некорректные данные"
// @Failure 403 {object} map[string]string "Неверный логин или пароль"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /login [post]
func (h *Handler) Login(ctx *gin.Context) {
	req := dto.LoginReq{}
	err := json.NewDecoder(ctx.Request.Body).Decode(&req)

	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	response, err := h.Service.Login(req, h.Config.JWT)
	if err != nil {
		if errors.Is(err, service.ErrInternalServerError) {
			h.errorHandler(ctx, http.StatusInternalServerError, "не удалось создать токен")
		} else if errors.Is(err, service.ErrForbidden) {
			h.errorHandler(ctx, http.StatusForbidden, "пароль неверный")
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusForbidden, "такого пользователя не существует")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// Logout godoc
// @Summary Выход из системы
// @Description Добавляет JWT токен в черный список
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 "Успешный выход"
// @Failure 400 {object} map[string]string "Некорректный токен"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /logout [post]
func (h *Handler) Logout(ctx *gin.Context) {
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		ctx.AbortWithStatus(http.StatusBadRequest)

		return
	}
	jwtStr = jwtStr[len(jwtPrefix):]

	_, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.Config.JWT.Token), nil
	})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		log.Println(err)

		return
	}

	err = h.Redis.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, h.Config.JWT.ExpiresIn)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusOK)

}
