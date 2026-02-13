package services

import (
	"errors"
	"strings"

	"stok-hadiah/models"
	"stok-hadiah/repositories"
)

type CounterStaffService struct {
	Repo *repositories.CounterStaffRepository
}

func (s *CounterStaffService) GetByStoreIDs(storeIDs []int) ([]models.CounterStaff, error) {
	if len(storeIDs) == 0 {
		return []models.CounterStaff{}, nil
	}
	return s.Repo.GetByStoreIDs(storeIDs)
}

func (s *CounterStaffService) CreateCounterStaff(input models.CounterStaffCreateInput) error {
	if input.CounterID <= 0 {
		return errors.New("counter wajib dipilih")
	}
	if input.UserID <= 0 {
		return errors.New("staff wajib dipilih")
	}

	status, err := normalizeCounterStaffStatus(input.Status)
	if err != nil {
		return err
	}

	return s.Repo.Create(models.CounterStaffCreateInput{
		CounterID: input.CounterID,
		UserID:    input.UserID,
		Status:    status,
	})
}

func (s *CounterStaffService) UpdateCounterStaff(input models.CounterStaffUpdateInput) error {
	if input.ID <= 0 {
		return errors.New("counter staff tidak valid")
	}
	if input.CounterID <= 0 {
		return errors.New("counter wajib dipilih")
	}
	if input.UserID <= 0 {
		return errors.New("staff wajib dipilih")
	}

	status, err := normalizeCounterStaffStatus(input.Status)
	if err != nil {
		return err
	}

	return s.Repo.Update(models.CounterStaffUpdateInput{
		ID:        input.ID,
		CounterID: input.CounterID,
		UserID:    input.UserID,
		Status:    status,
	})
}

func (s *CounterStaffService) DeleteCounterStaff(id int) error {
	if id <= 0 {
		return errors.New("counter staff id tidak valid")
	}
	return s.Repo.Delete(id)
}

func (s *CounterStaffService) GetStatusByCounterAndUser(counterID, userID int) (string, error) {
	if counterID <= 0 || userID <= 0 {
		return "", errors.New("counter atau user tidak valid")
	}

	status, _, err := s.Repo.GetStatusByCounterAndUser(counterID, userID)
	if err != nil {
		return "", err
	}

	return status, nil
}

func (s *CounterStaffService) UpdateStatusByCounterAndUser(input models.CounterStaffStatusUpdateInput) (models.CounterStaffStatusDetail, error) {
	if input.CounterID <= 0 || input.UserID <= 0 {
		return models.CounterStaffStatusDetail{}, errors.New("counter atau user tidak valid")
	}

	normalized, err := normalizeCounterStaffStatus(input.Status)
	if err != nil {
		return models.CounterStaffStatusDetail{}, err
	}

	input.Status = normalized
	if normalized == "INACTIVE" {
		if input.InactiveStartedAt == nil || input.InactiveUntil == nil {
			return models.CounterStaffStatusDetail{}, errors.New("jadwal non-aktif wajib diisi")
		}
		if !input.InactiveUntil.After(*input.InactiveStartedAt) {
			return models.CounterStaffStatusDetail{}, errors.New("inactive_until harus setelah inactive_started_at")
		}
		input.InactiveAnnouncement = strings.TrimSpace(input.InactiveAnnouncement)
		if input.InactiveAnnouncement == "" {
			return models.CounterStaffStatusDetail{}, errors.New("pengumuman non-aktif wajib diisi")
		}
		if len(input.InactiveAnnouncement) > 255 {
			return models.CounterStaffStatusDetail{}, errors.New("pengumuman non-aktif maksimal 255 karakter")
		}
	} else {
		input.InactiveStartedAt = nil
		input.InactiveUntil = nil
		input.InactiveAnnouncement = ""
	}

	updated, err := s.Repo.UpdateStatusByCounterAndUser(input)
	if err != nil {
		return models.CounterStaffStatusDetail{}, err
	}
	if !updated {
		return models.CounterStaffStatusDetail{}, errors.New("staff belum terdaftar pada loket ini")
	}

	detail, exists, err := s.Repo.GetStatusDetailByCounterAndUser(input.CounterID, input.UserID)
	if err != nil {
		return models.CounterStaffStatusDetail{}, err
	}
	if !exists {
		return models.CounterStaffStatusDetail{}, errors.New("staff belum terdaftar pada loket ini")
	}

	return detail, nil
}

func normalizeCounterStaffStatus(status string) (string, error) {
	val := strings.TrimSpace(strings.ToUpper(status))

	switch val {
	case "", "ACTIVE", "AKTIF":
		return "ACTIVE", nil
	case "REST", "ISTIRAHAT":
		return "REST", nil
	case "INACTIVE", "NON_ACTIVE", "NON AKTIF", "NONAKTIF", "INAKTIF":
		return "INACTIVE", nil
	default:
		return "", errors.New("status tidak valid")
	}
}
