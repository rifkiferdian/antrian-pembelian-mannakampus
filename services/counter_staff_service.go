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
