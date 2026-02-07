package services

import (
	"database/sql"
	"errors"
	"time"

	"stok-hadiah/repositories"
)

var (
	ErrNoWaitingTicket = errors.New("no waiting ticket")
	ErrNoCalledTicket  = errors.New("no called ticket")
	ErrNoSkippedTicket = errors.New("no skipped ticket")
)

type QueueService struct {
	Repo *repositories.QueueActionRepository
}

func (s *QueueService) CallNext(storeID, counterID, userID int) (int64, string, error) {
	tx, err := s.Repo.DB.Begin()
	if err != nil {
		return 0, "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	ticketDate := time.Now()
	ticketID, ticketNo, err := s.Repo.GetNextWaitingTicket(tx, storeID, counterID, ticketDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", ErrNoWaitingTicket
		}
		return 0, "", err
	}

	now := time.Now()
	currentID, _, currentErr := s.Repo.GetCurrentCalledTicket(tx, storeID, counterID, ticketDate)
	if currentErr != nil && !errors.Is(currentErr, sql.ErrNoRows) {
		return 0, "", currentErr
	}
	if currentID > 0 {
		if err = s.Repo.MarkDone(tx, currentID, now); err != nil {
			return 0, "", err
		}
		if err = s.Repo.InsertEvent(tx, currentID, "DONE", userID, "Selesai otomatis saat panggil berikutnya"); err != nil {
			return 0, "", err
		}
	}
	if err = s.Repo.MarkCalled(tx, ticketID, userID, now); err != nil {
		return 0, "", err
	}
	if err = s.Repo.InsertEvent(tx, ticketID, "CALL", userID, "Panggil nomor"); err != nil {
		return 0, "", err
	}

	if err = tx.Commit(); err != nil {
		return 0, "", err
	}
	return ticketID, ticketNo, nil
}

func (s *QueueService) Recall(storeID, counterID, userID int) (int64, string, error) {
	tx, err := s.Repo.DB.Begin()
	if err != nil {
		return 0, "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	ticketDate := time.Now()
	ticketID, ticketNo, err := s.Repo.GetCurrentCalledTicket(tx, storeID, counterID, ticketDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", ErrNoCalledTicket
		}
		return 0, "", err
	}

	now := time.Now()
	if err = s.Repo.MarkCalled(tx, ticketID, userID, now); err != nil {
		return 0, "", err
	}
	if err = s.Repo.InsertEvent(tx, ticketID, "CALL", userID, "Panggil ulang"); err != nil {
		return 0, "", err
	}

	if err = tx.Commit(); err != nil {
		return 0, "", err
	}
	return ticketID, ticketNo, nil
}

func (s *QueueService) Finish(storeID, counterID, userID int) (int64, string, error) {
	tx, err := s.Repo.DB.Begin()
	if err != nil {
		return 0, "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	ticketDate := time.Now()
	ticketID, ticketNo, err := s.Repo.GetCurrentCalledTicket(tx, storeID, counterID, ticketDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", ErrNoCalledTicket
		}
		return 0, "", err
	}

	now := time.Now()
	if err = s.Repo.MarkDone(tx, ticketID, now); err != nil {
		return 0, "", err
	}
	if err = s.Repo.InsertEvent(tx, ticketID, "DONE", userID, "Selesai layanan"); err != nil {
		return 0, "", err
	}

	if err = tx.Commit(); err != nil {
		return 0, "", err
	}
	return ticketID, ticketNo, nil
}

func (s *QueueService) Skip(storeID, counterID, userID int) (int64, string, error) {
	tx, err := s.Repo.DB.Begin()
	if err != nil {
		return 0, "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	ticketDate := time.Now()
	ticketID, ticketNo, err := s.Repo.GetCurrentCalledTicket(tx, storeID, counterID, ticketDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", ErrNoCalledTicket
		}
		return 0, "", err
	}

	now := time.Now()
	if err = s.Repo.MarkSkipped(tx, ticketID, now); err != nil {
		return 0, "", err
	}
	if err = s.Repo.InsertEvent(tx, ticketID, "SKIP", userID, "Lewati nomor"); err != nil {
		return 0, "", err
	}

	if err = tx.Commit(); err != nil {
		return 0, "", err
	}
	return ticketID, ticketNo, nil
}

func (s *QueueService) RecallSkipped(storeID, counterID, userID int, ticketID int64) (int64, string, error) {
	tx, err := s.Repo.DB.Begin()
	if err != nil {
		return 0, "", err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	ticketDate := time.Now()
	targetID, ticketNo, err := s.Repo.GetSkippedTicketByID(tx, storeID, counterID, ticketID, ticketDate)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", ErrNoSkippedTicket
		}
		return 0, "", err
	}

	now := time.Now()
	currentID, _, currentErr := s.Repo.GetCurrentCalledTicket(tx, storeID, counterID, ticketDate)
	if currentErr != nil && !errors.Is(currentErr, sql.ErrNoRows) {
		return 0, "", currentErr
	}
	if currentID > 0 {
		if err = s.Repo.MarkDone(tx, currentID, now); err != nil {
			return 0, "", err
		}
		if err = s.Repo.InsertEvent(tx, currentID, "DONE", userID, "Selesai otomatis saat panggil ulang terlewat"); err != nil {
			return 0, "", err
		}
	}

	if err = s.Repo.MarkCalled(tx, targetID, userID, now); err != nil {
		return 0, "", err
	}
	if err = s.Repo.InsertEvent(tx, targetID, "CALL", userID, "Panggil ulang setelah terlewat"); err != nil {
		return 0, "", err
	}

	if err = tx.Commit(); err != nil {
		return 0, "", err
	}
	return targetID, ticketNo, nil
}
