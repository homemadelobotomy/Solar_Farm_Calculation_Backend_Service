package repository

import (
	dto "lab/internal/app/DTO"
	"lab/internal/app/ds"
	"time"

	"gorm.io/gorm"
)

func (r *Repository) GetSolarPanelsInRequest(userId uint) (uint, int64, error) {
	//TODO Вернуть id заявки черновика и количество услуг в этой заявке
	var (
		requestId      uint
		numberOfPanels int64
	)
	err := r.db.Model(&ds.SolarPanelRequest{}).Where("creator_id = ? AND status = ? ", userId, "черновик").Select("id").First(&requestId).Error
	if err != nil {
		return 0, 0, err
	}
	err = r.db.Model(&ds.RequestPanels{}).Where("solar_panel_request_id = ?", requestId).Count(&numberOfPanels).Error
	if err != nil {
		return requestId, 0, err
	}

	return requestId, numberOfPanels, nil
}

func (r *Repository) GetFilteredSolarPanelRequests(userId uint, filter dto.SolarPanleRequestFilter) ([]ds.SolarPanelRequest, error) {
	//TODO Вернуть заявки для пользователя, кроме заявок со статусом удален и черновик, отфильтрованные по дате и статусу
	// так же заменить поля модератора и создателя на логины
	var solarPanelRequests []ds.SolarPanelRequest
	db := r.db.Model(&ds.SolarPanelRequest{}).Where("creator_id = ? AND status NOT IN ('черновик','удален')", userId)
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if !filter.Start_date.IsZero() {
		db = db.Where("formated_at >= ?", filter.Start_date)
	}
	if !filter.End_date.IsZero() {
		db = db.Where("formated_at <= ?", filter.End_date)
	}
	err := db.Order("formated_at DESC").
		Preload("Creator").
		Preload("Moderator").
		Find(&solarPanelRequests).Error
	if err != nil {
		return []ds.SolarPanelRequest{}, err
	}
	return solarPanelRequests, nil
}
func (r *Repository) GetAllFilteredSolarPanelRequests(filter dto.SolarPanleRequestFilter) ([]ds.SolarPanelRequest, error) {
	//TODO Вернуть заявки для пользователя, кроме заявок со статусом удален и черновик, отфильтрованные по дате и статусу
	// так же заменить поля модератора и создателя на логины
	var solarPanelRequests []ds.SolarPanelRequest
	db := r.db.Model(&ds.SolarPanelRequest{}).Where("status NOT IN ('черновик','удален')")
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if !filter.Start_date.IsZero() {
		db = db.Where("formated_at >= ?", filter.Start_date)
	}
	if !filter.End_date.IsZero() {
		db = db.Where("formated_at <= ?", filter.End_date)
	}
	err := db.
		Order("formated_at DESC").
		Preload("Creator").
		Preload("Moderator").
		Find(&solarPanelRequests).Error
	if err != nil {
		return []ds.SolarPanelRequest{}, err
	}
	return solarPanelRequests, nil
}

func (r *Repository) GetOneSolarPanelRequest(requestId uint, status string) (ds.SolarPanelRequest, error) {
	var solarPanelRequest ds.SolarPanelRequest
	err := r.db.Where("id = ? AND status <> 'удален' ", requestId).
		Preload("Panels.SolarPanel").
		Preload("Panels", func(db *gorm.DB) *gorm.DB {
			return db.Order("solar_panel_id ASC") // Сортировка по ID панели
		}).
		First(&solarPanelRequest).Error
	if err != nil {
		return ds.SolarPanelRequest{}, err
	}
	return solarPanelRequest, nil

}
func (r *Repository) UpdateTotalPower(requestId uint, totalPower float64) error {
	err := r.db.Model(&ds.SolarPanelRequest{}).
		Where("id = ? AND status = 'завершен'", requestId).
		Update("total_power", totalPower).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) ChangeSolarPanelRequest(requestId uint, insolation float64) error {
	//TODO Изменить значение инсоляции
	err := r.db.Model(&ds.SolarPanelRequest{}).
		Where("id = ? AND status = 'черновик'", requestId).
		Update("insolation", insolation).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) FormateSolarPanelRequest(requestId uint) error {
	//TODO Изменить статус черновика пользователя и проставить дату формирования
	err := r.db.Model(&ds.SolarPanelRequest{}).
		Where("id = ? AND status = 'черновик' ", requestId).
		Updates(map[string]any{
			"formated_at": time.Now(),
			"status":      "сформирован",
		}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) ModeratorAction(requestId uint, action string, totalPower float64, moderatorId uint) error {
	//TODO Отклонение/Завершение заявки модератором, проставить модератора, дату действия, рассчитать поле итоговой мощности
	err := r.db.Model(&ds.SolarPanelRequest{}).
		Where("id = ? AND status = 'сформирован'", requestId).
		Updates(map[string]any{
			"moderated_at": time.Now(),
			"status":       action,
			"total_power":  totalPower,
			"moderator_id": moderatorId,
		}).Error

	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) DeleteSolarPanelRequest(requestId uint) error {
	//TODO Проставить статус заявки удален и прописать дату удаления
	err := r.db.Model(&ds.SolarPanelRequest{}).
		Where("id = ? AND status = 'черновик'", requestId).
		Updates(map[string]any{
			"deleted_at": time.Now(),
			"status":     "удален",
		}).Error
	if err != nil {
		return err
	}
	return nil

}
