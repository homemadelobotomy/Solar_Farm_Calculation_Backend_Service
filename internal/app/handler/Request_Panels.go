package handler

import (
	"errors"
	dto "lab/internal/app/DTO"
	_ "lab/internal/app/ds"
	"lab/internal/app/role"
	"lab/internal/app/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) RegisterRequestPanelsHandlers(router *gin.Engine) {
	router.DELETE("/api/solarpanel-requests/:id/:solarpanelId", h.WithAuthCheck(role.User, role.Moderator), h.DeleteSolarPanelFromRequest)
	router.PUT("/api/solarpanel-requests/:id/:solarpanelId", h.WithAuthCheck(role.User, role.Moderator), h.ChangeSolarPanelArea)
}

// DeleteSolarPanelFromRequest godoc
// @Summary Удалить панель из заявки
// @Description Удаляет солнечную панель из заявки пользователя
// @Tags RequestPanels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param solarpanelId path int true "ID солнечной панели"
// @Success 200 {object} map[string]string "Панель успешно удалена"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Заявка недоступна / Панель отсутствует"
// @Failure 404 {object} map[string]string "Заявка или панель не найдены"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests/{id}/{solarpanelId} [delete]
func (h *Handler) DeleteSolarPanelFromRequest(ctx *gin.Context) {
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
	solarPanelId, err := strconv.Atoi(ctx.Param("solarpanelId"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id солнечной панели")
		return
	}

	err = h.Service.DeleteSolarPanelFromRequest(userId, uint(requestId), uint(solarPanelId))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найдено")
		} else if errors.Is(err, service.ErrForbidden) {
			h.errorHandler(ctx, http.StatusForbidden,
				"заявка по солнечым панелям не доступна этому пользователю")
		} else if errors.Is(err, service.ErrNoRecords) {
			h.errorHandler(ctx, http.StatusForbidden,
				"солнечная панель с таким id отсутсвует в заявке")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "200 OK",
		"message": "солнечная панель успешно удалена из заявки по солнечным панелям",
	})
}

// ChangeSolarPanelArea godoc
// @Summary Изменить площадь панели в заявке
// @Description Обновляет значение площади для панели в заявке
// @Tags RequestPanels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID заявки"
// @Param solarpanelId path int true "ID солнечной панели"
// @Param area body dto.ChangeSolarPanelAreaRequest true "Новое значение площади"
// @Success 200 {object} ds.RequestPanels "Обновленная панель в заявке"
// @Failure 400 {object} map[string]string "Некорректное значение площади"
// @Failure 401 {object} map[string]string "Не авторизован"
// @Failure 403 {object} map[string]string "Заявка недоступна этому пользователю"
// @Failure 404 {object} map[string]string "Заявка или панель не найдены"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /solarpanel-requests/{id}/{solarpanelId} [put]
func (h *Handler) ChangeSolarPanelArea(ctx *gin.Context) {

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
	solarPanelId, err := strconv.Atoi(ctx.Param("solarpanelId"))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound,
			"введен некорректный id солнечной панели")
		return
	}

	var area dto.ChangeSolarPanelAreaRequest

	if err = ctx.BindJSON(&area); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest,
			"введен неправильный формат тела запроса")
		return
	}
	response, err := h.Service.ChangeSolarPanelAreaInRequest(userId, uint(requestId), uint(solarPanelId), area.Area)

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.errorHandler(ctx, http.StatusNotFound,
				"заявки по солнечным панелям с таким id не найдено")
		} else if errors.Is(err, service.ErrForbidden) {
			h.errorHandler(ctx, http.StatusForbidden,
				"заявка по солнечым панелям не доступна этому пользователю")
		} else if errors.Is(err, service.ErrNoRecords) {
			h.errorHandler(ctx, http.StatusNotFound,
				"солнечная панель с таким id отсутсвует в заявке")
		} else if errors.Is(err, service.ErrBadRequest) {
			h.errorHandler(ctx, http.StatusBadRequest,
				"введено неверное значение для площади")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	ctx.JSON(http.StatusOK, response)
}
