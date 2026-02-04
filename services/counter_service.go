package services

import (
	"errors"
	"strings"

	"stok-hadiah/models"
	"stok-hadiah/repositories"
)

type CounterService struct {
	Repo *repositories.CounterRepository
}

func (s *CounterService) GetCounters() ([]models.Counter, error) {
	return s.Repo.GetAll()
}

func (s *CounterService) CreateCounter(input models.CounterCreateInput) error {
	counterCode := strings.TrimSpace(input.CounterCode)
	counterName := strings.TrimSpace(input.CounterName)
	ticketPrefix := strings.TrimSpace(input.TicketPrefix)

	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if counterCode == "" {
		return errors.New("kode counter wajib diisi")
	}
	if counterName == "" {
		return errors.New("nama counter wajib diisi")
	}
	if ticketPrefix == "" {
		return errors.New("ticket prefix wajib diisi")
	}

	return s.Repo.Create(models.CounterCreateInput{
		StoreID:      input.StoreID,
		CounterCode:  counterCode,
		CounterName:  counterName,
		TicketPrefix: ticketPrefix,
		IsActive:     input.IsActive,
	})
}

func (s *CounterService) UpdateCounter(input models.CounterUpdateInput) error {
	counterCode := strings.TrimSpace(input.CounterCode)
	counterName := strings.TrimSpace(input.CounterName)
	ticketPrefix := strings.TrimSpace(input.TicketPrefix)

	if input.ID <= 0 {
		return errors.New("counter tidak valid")
	}
	if input.StoreID <= 0 {
		return errors.New("store wajib dipilih")
	}
	if counterCode == "" {
		return errors.New("kode counter wajib diisi")
	}
	if counterName == "" {
		return errors.New("nama counter wajib diisi")
	}
	if ticketPrefix == "" {
		return errors.New("ticket prefix wajib diisi")
	}

	return s.Repo.Update(models.CounterUpdateInput{
		ID:           input.ID,
		StoreID:      input.StoreID,
		CounterCode:  counterCode,
		CounterName:  counterName,
		TicketPrefix: ticketPrefix,
		IsActive:     input.IsActive,
	})
}

func (s *CounterService) DeleteCounter(id int) error {
	if id <= 0 {
		return errors.New("counter id tidak valid")
	}
	return s.Repo.Delete(id)
}
