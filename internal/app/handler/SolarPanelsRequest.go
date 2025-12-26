package handler

import (
	"errors"
	dto "lab/internal/app/DTO"
	"lab/internal/app/role"
	"lab/internal/app/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) RegisterSolarPanelsRequestHandlers(router *gin.Engine) {
	solarPanelRequestGroups := router.Group("/api/solarpanel-requests")
	{
		solarPanelRequestGroups.GET("/info", h.WithOptionalAuth(), h.GetSolarPanelsInRequest)
		solarPanelRequestGroups.GET("", h.WithAuthCheck(role.Moderator, role.User), h.GetFilteredSolarPanelRequests)
		solarPanelRequestGroups.GET("/:id", h.WithAuthCheck(role.Moderator, role.User), h.GetOneSolarPanelRequest)
		solarPanelRequestGroups.PUT("/:id", h.WithAuthCheck(role.User, role.Moderator), h.ChangeSolarPanelRequest)
		solarPanelRequestGroups.PUT("/:id/formate", h.WithAuthCheck(role.User, role.Moderator), h.FormateSolarPanelRequest)
		solarPanelRequestGroups.PUT("/:id/moderate", h.WithAuthCheck(role.Moderator), h.ModeratorAction)
		solarPanelRequestGroups.PUT("/:id/update-total-power", h.UpdateCalculationResult)
		solarPanelRequestGroups.DELETE("/:id", h.WithAuthCheck(role.User, role.Moderator), h.DeleteSolarPanelRequest)
	}
}

// GetSolarPanelsInRequest godoc
// @Summary Получить информацию о корзине
// @Description Возвращает ID текущей корзины и количество панелей в ней
// @Tags SolarPanelRequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.NumberOfPanelsResponse "Информация о корзине"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests/info [get]
func (h *Handler) GetSolarPanelsInRequest(ctx *gin.Context) {

	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusOK, dto.NumberOfPanelsResponse{
			RequestId:      0,
			NumberOfPanels: 0,
		})
		return
	}

	requestId, numberOfPanels, err := h.Service.GetSolarPanelsInRequest(userId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusOK, dto.NumberOfPanelsResponse{
				RequestId:      0,
				NumberOfPanels: 0,
			})
		} else {
			ctx.JSON(http.StatusOK, dto.NumberOfPanelsResponse{
				RequestId:      0,
				NumberOfPanels: 0,
			})
		}
		return
	}

	ctx.JSON(http.StatusOK, dto.NumberOfPanelsResponse{
		RequestId:      requestId,
		NumberOfPanels: numberOfPanels,
	})
}

// GetFilteredSolarPanelRequests godoc
// @Summary Получить отфильтрованные заявки
// @Description Возвращает список заявок пользователя с фильтрацией по дате и статусу
// @Tags SolarPanelRequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Начальная дата (формат: dd-mm-yyyy hh:mm:ss)"
// @Param end_date query string false "Конечная дата (формат: dd-mm-yyyy hh:mm:ss)"
// @Param status query string false "Статус заявки (сформирован, завершен, отклонен)"
// @Success 200 {array} dto.SolarPanelsRequestsResponse "Список заявок"
// @Failure 400 {object} map[string]string "Некорректный формат даты"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 404 {object} map[string]string "Заявки не найдены"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests [get]
func (h *Handler) GetFilteredSolarPanelRequests(ctx *gin.Context) {
	var (
		startDate time.Time
		endDate   time.Time
		err       error
	)

	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}
	userRole, exists := GetUserRoleFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}
	start_date := ctx.Query("start_date")
	end_date := ctx.Query("end_date")
	status := ctx.Query("status")

	layout := "02-01-2006 15:04:05"

	if start_date != "" {
		startDate, err = time.Parse(layout, start_date)
		if err != nil {

			h.errorHandler(ctx, http.StatusBadRequest,
				"введен неверный формат start_date, правильный формат: dd-mm-yyyy hh:mm:ss")
			return
		}
	} else {
		startDate = time.Time{}
	}

	if end_date != "" {
		endDate, err = time.Parse(layout, end_date)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest,
				"введен неверный формат end_date, правильный формат: dd-mm-yyyy hh:mm:ss")
			return
		}
	} else {
		endDate = time.Time{}
	}

	filter := dto.SolarPanleRequestFilter{
		Status:     status,
		Start_date: startDate,
		End_date:   endDate,
	}
	response, err := h.Service.GetFilteredSolarPanelRequests(userId, filter, userRole)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, service.ErrNoRecords) {
			ctx.JSON(http.StatusOK, []dto.SolarPanelsRequestsResponse{})
		} else {
			ctx.JSON(http.StatusOK, []dto.SolarPanelsRequestsResponse{})
		}
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// GetOneSolarPanelRequest godoc
// @Summary Получить заявку по ID
// @Description Возвращает детальную информацию о заявке со списком панелей
// @Tags SolarPanelRequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} dto.OneSolarPanelRequestResponse "Данные заявки"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Заявка недоступна этому пользователю"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests/{id} [get]
func (h *Handler) GetOneSolarPanelRequest(ctx *gin.Context) {
	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}

	requestId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id заявки по солнечным панелям")
	}
	response, err := h.Service.GetOneSolarPanelRequest(uint(requestId), userId)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найден")
		} else if errors.Is(err, service.ErrForbidden) {
			h.errorHandler(ctx, http.StatusForbidden,
				"заявка по солнечым панелям не доступна этому пользователю")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, response)

}

// ChangeSolarPanelRequest godoc
// @Summary Изменить инсоляцию в заявке
// @Description Обновляет значение инсоляции для заявки со статусом черновик
// @Tags SolarPanelRequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param insolation body dto.ChangeSolarPanelRequest true "Новое значение инсоляции"
// @Success 200 {object} dto.OneSolarPanelRequestResponse "Обновленная заявка"
// @Failure 400 {object} map[string]string "Некорректное значение инсоляции"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Заявка недоступна этому пользователю"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests/{id} [put]
func (h *Handler) ChangeSolarPanelRequest(ctx *gin.Context) {
	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}

	var (
		insolationRequest dto.ChangeSolarPanelRequest
		response          dto.OneSolarPanelRequestResponse
	)
	if err := ctx.BindJSON(&insolationRequest); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest,
			"введен неправильный формат тела запроса")
		return
	}
	requestId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id заявки по солнечным панелям")
		return
	}
	response, err = h.Service.ChangeSolarPanelRequest(userId, uint(requestId), insolationRequest.Insolation)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найдено")
		} else if errors.Is(err, service.ErrForbidden) {
			h.errorHandler(ctx, http.StatusForbidden,
				"заявка по солнечым панелям не доступна этому пользователю")
		} else if errors.Is(err, service.ErrBadRequest) {
			h.errorHandler(ctx, http.StatusBadRequest,
				"введено некорректное значение для инсоляции")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// FormateSolarPanelRequest godoc
// @Summary Сформировать заявку
// @Description Переводит заявку из статуса черновик в статус сформирован
// @Tags SolarPanelRequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} dto.OneSolarPanelRequestResponse "Сформированная заявка"
// @Failure 400 {object} map[string]string "Некорректные данные в заявке"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Заявка недоступна этому пользователю"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests/{id}/formate [put]
func (h *Handler) FormateSolarPanelRequest(ctx *gin.Context) {
	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}
	requestId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id заявки по солнечным панелям")
		return
	}
	response, err := h.Service.FormateSolarPanelRequest(uint(requestId), userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найдено")
		} else if errors.Is(err, service.ErrForbidden) {
			h.errorHandler(ctx, http.StatusForbidden,
				"заявка по солнечым панелям не доступна этому пользователю")
		} else if errors.Is(err, service.ErrBadRequest) {
			h.errorHandler(ctx, http.StatusBadRequest,
				"введено некорректное значение для инсоляции или значения площади  < 0")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// ModeratorAction godoc
// @Summary Действие модератора (Модератор)
// @Description Модератор завершает или отклоняет заявку. Требуется роль модератора.
// @Tags SolarPanelRequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param action body dto.ModeratorAction true "Действие модератора (завершен/отклонен)"
// @Success 200 {object} dto.OneSolarPanelRequestResponse "Обработанная заявка"
// @Failure 400 {object} map[string]string "Некорректное действие"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Недостаточно прав"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests/{id}/moderate [put]
func (h *Handler) ModeratorAction(ctx *gin.Context) {
	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}
	requestId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id заявки по солнечным панелям")
		return
	}
	var action dto.ModeratorAction
	if err = ctx.BindJSON(&action); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest,
			"введен неправильный формат тела запроса")
		return
	}
	response, err := h.Service.ModeratorAction(uint(requestId), action.Action, userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найдено")
		} else if errors.Is(err, service.ErrForbidden) {
			h.errorHandler(ctx, http.StatusForbidden,
				"пользователь не является модератором")
		} else if errors.Is(err, service.ErrBadRequest) {
			h.errorHandler(ctx, http.StatusBadRequest,
				"действие модератора должно быть 'отклонен|завершен'")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, response)
}

// DeleteSolarPanelRequest godoc
// @Summary Удалить заявку
// @Description Помечает заявку как удаленную (доступно только для черновиков)
// @Tags SolarPanelRequests
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Success 200 {object} map[string]string "Заявка успешно удалена"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Заявка недоступна этому пользователю"
// @Failure 404 {object} map[string]string "Заявка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests/{id} [delete]
func (h *Handler) DeleteSolarPanelRequest(ctx *gin.Context) {
	userId, exists := GetUserIdFromContext(ctx)
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, "не авторизован")
		return
	}
	requestId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id заявки по солнечным панелям")
		return
	}
	err = h.Service.DeleteSolarPanelRequest(uint(requestId), userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найдено")
		} else if errors.Is(err, service.ErrForbidden) {
			h.errorHandler(ctx, http.StatusForbidden,
				"заявка по солнечым панелям не доступна этому пользователю")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "200 OK",
		"message": "заявка по солнечным панелям успешно удалена",
	})
}

func (h *Handler) UpdateCalculationResult(ctx *gin.Context) {
	const SERVICE_TOKEN = "12345678"

	requestId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "некорректный id заявки")
		return
	}

	var updateData dto.CalculationResultUpdate
	if err := ctx.BindJSON(&updateData); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, "неверный формат данных")
		return
	}
	if updateData.Token != SERVICE_TOKEN {
		h.errorHandler(ctx, http.StatusUnauthorized, "неверный токен сервиса")
		return
	}

	err = h.Service.UpdateCalculationResult(uint(requestId), updateData.TotalPower)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "результат обновлен"})
}
