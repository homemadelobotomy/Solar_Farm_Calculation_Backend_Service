package handler

import (
	"errors"
	dto "lab/internal/app/DTO"
	"lab/internal/app/ds"
	"lab/internal/app/role"
	"lab/internal/app/service"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

func (h *Handler) RegisterSolarPanelHandlers(router *gin.Engine) {
	panelsGroup := router.Group("/api/panels")
	{
		panelsGroup.GET("", h.GetSolarPanels)
		panelsGroup.GET("/:id", h.GetOneSolarPanel)
		panelsGroup.POST("", h.WithAuthCheck(role.Moderator), h.AddNewSolarPanel)
		panelsGroup.PUT("/:id", h.WithAuthCheck(role.Moderator), h.ChangeSolarPanel)
		panelsGroup.DELETE("/:id", h.WithAuthCheck(role.Moderator), h.DeleteSolarPanel)
		panelsGroup.POST("/:id", h.WithAuthCheck(role.User, role.Moderator), h.AddSolarPanelToRequest)
		panelsGroup.POST("/:id/image", h.WithAuthCheck(role.Moderator), h.AddImageToSolarPanel)
	}
}

// GetSolarPanels godoc
// @Summary Получить список солнечных панелей
// @Description Возвращает список всех солнечных панелей с возможностью фильтрации по мощности
// @Tags SolarPanels
// @Accept json
// @Produce json
// @Param start_value query number false "Минимальная мощность для фильтрации"
// @Param end_value query number false "Максимальная мощность для фильтрации"
// @Success 200 {array}  ds.SolarPanel "Список солнечных панелей"
// @Failure 400 {object} map[string]string "Некорректные параметры фильтрации"
// @Failure 404 {object} map[string]string "Панели не найдены"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /panels [get]
func (h *Handler) GetSolarPanels(ctx *gin.Context) {
	var (
		startValue float64
		endValue   float64
		err        error
	)
	startValueStr := ctx.Query("start_value")
	endValueStr := ctx.Query("end_value")

	if startValueStr != "" {
		startValue, err = strconv.ParseFloat(startValueStr, 64)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest,
				"введено некорректное значение start_value")
		}
	} else {
		startValue = 0
	}

	if endValueStr != "" {
		endValue, err = strconv.ParseFloat(endValueStr, 64)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest,
				"введено некорректное значение end_value")
		}
	} else {
		endValue = 0
	}

	response, err := h.Service.GetSolarPanels(startValue, endValue)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найден")
		} else if errors.Is(err, service.ErrNoRecords) {
			ctx.JSON(http.StatusOK, []ds.SolarPanel{})
		} else if errors.Is(err, service.ErrBadRequest) {
			h.errorHandler(ctx, http.StatusBadRequest,
				"введено недопсутимое значение мощности для фильтрации")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// GetOneSolarPanel godoc
// @Summary Получить солнечную панель по ID
// @Description Возвращает детальную информацию о конкретной солнечной панели
// @Tags SolarPanels
// @Accept json
// @Produce json
// @Param id path int true "ID солнечной панели"
// @Success 200 {object} ds.SolarPanel "Данные солнечной панели"
// @Failure 404 {object} map[string]string "Панель не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /panels/{id} [get]
func (h *Handler) GetOneSolarPanel(ctx *gin.Context) {
	solarPanelId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id солнечной панели")
		return
	}
	response, err := h.Service.GetSolarPanel(uint(solarPanelId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"солнечной панели с таким id не найдено")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// AddNewSolarPanel godoc
// @Summary Добавить новую солнечную панель (Модератор)
// @Description Создает новую солнечную панель. Требуется роль модератора.
// @Tags SolarPanels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param solarPanel body dto.AddSolarPanel true "Данные новой панели"
// @Success 201 {object} ds.SolarPanel "Созданная панель"
// @Failure 400 {object} map[string]string "Некорректные данные"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /panels [post]
func (h *Handler) AddNewSolarPanel(ctx *gin.Context) {
	var solarPanel dto.AddSolarPanel
	if err := ctx.BindJSON(&solarPanel); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest,
			"введен неправильный формат тела запроса")
		return
	}
	response, err := h.Service.AddNewSolarPanel(solarPanel)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusCreated, response)
}

// ChangeSolarPanel godoc
// @Summary Изменить солнечную панель (Модератор)
// @Description Обновляет данные существующей солнечной панели. Требуется роль модератора.
// @Tags SolarPanels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID солнечной панели"
// @Param solarPanel body dto.ChangeSolarPanel true "Обновленные данные панели"
// @Success 200 {object} ds.SolarPanel "Обновленная панель"
// @Failure 400 {object} map[string]string "Некорректные данные"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 404 {object} map[string]string "Панель не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /panels/{id} [put]
func (h *Handler) ChangeSolarPanel(ctx *gin.Context) {
	var solarPanel dto.ChangeSolarPanel
	solarPanelId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id солнечной панели")
		return
	}
	if err := ctx.BindJSON(&solarPanel); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest,
			"введен неправильный формат тела запроса: "+err.Error())
		return
	}
	response, err := h.Service.ChangeSolarPanel(uint(solarPanelId), solarPanel)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"солнечной панели с таким id не найдено")
		} else if errors.Is(err, service.ErrBadRequest) {
			h.errorHandler(ctx, http.StatusBadRequest,
				"числовые значения должны быть положительными")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// DeleteSolarPanel godoc
// @Summary Удалить солнечную панель (Модератор)
// @Description Помечает солнечную панель как удаленную. Требуется роль модератора.
// @Tags SolarPanels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID солнечной панели"
// @Success 200 {object} map[string]string "Панель успешно удалена"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 404 {object} map[string]string "Панель не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /panels/{id} [delete]
func (h *Handler) DeleteSolarPanel(ctx *gin.Context) {
	solarPanelId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id солнечной панели")
		return
	}
	err = h.Service.DeleteSolarPanel(uint(solarPanelId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"солнечной панели с таким id не найдено")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "200 OK",
		"message": "солнечная панель успешно удалена",
	})
}

// AddSolarPanelToRequest godoc
// @Summary Добавить панель в корзину
// @Description Добавляет солнечную панель в текущую заявку пользователя (корзину)
// @Tags SolarPanels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID солнечной панели"
// @Success 200 {object} map[string]string "Панель добавлена в заявку"
// @Failure 400 {object} map[string]string "Панель уже в заявке или удалена"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 404 {object} map[string]string "Панель не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /panels/{id} [post]
func (h *Handler) AddSolarPanelToRequest(ctx *gin.Context) {
	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}

	solarPanelId, err := strconv.Atoi(ctx.Param("id"))
	var pgErr *pgconn.PgError
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id солнечной панели")
		return
	}
	err = h.Service.AddSolarPanelToRequest(uint(solarPanelId), userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявка по слонечным панелям не найдена")
		} else if errors.Is(err, service.ErrNoRecords) {
			h.errorHandler(ctx, http.StatusNotFound,
				"солнечной панели с таким id не найдено")
		} else if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			h.errorHandler(ctx, http.StatusBadRequest,
				"такая солнечная панель уже добавлена в заявку")
		} else if errors.Is(err, service.ErrSolarPanelDeleted) {
			h.errorHandler(ctx, http.StatusBadRequest,
				"эта солнечная панель удалена")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "201 Created",
		"message": "солнечная панель добавлена в заявку",
	})
}

// AddImageToSolarPanel godoc
// @Summary Загрузить изображение панели (Модератор)
// @Description Загружает изображение для солнечной панели. Требуется роль модератора.
// @Tags SolarPanels
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID солнечной панели"
// @Param image formData file true "Файл изображения"
// @Success 201 {object} ds.SolarPanel "Панель с обновленным изображением"
// @Failure 400 {object} map[string]string "Файл не найден"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 404 {object} map[string]string "Панель не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /panels/{id}/image [post]
func (h *Handler) AddImageToSolarPanel(ctx *gin.Context) {
	solarPanelId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id солнечной панели")
		return
	}
	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "файл не найден")
		return
	}

	filename := uuid.New().String() + filepath.Ext(file.Filename)

	response, err := h.Service.AddImageToSolarPanel(uint(solarPanelId), file, filename)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"солнечной панели с таким id не найдено")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusCreated, response)

}
