package controllers

import (
	"strings"
	"time"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"
)

func buildQueueViewPayload(storeID int, action string) (models.QueueViewPayload, error) {
	currentCall, counters, today, err := buildQueueViewState(storeID)
	if err != nil {
		return models.QueueViewPayload{}, err
	}

	guestRepo := &repositories.GuestQueueRepository{DB: config.DB}
	guestCounters, err := guestRepo.GetCountersForGuest(storeID, today)
	if err != nil {
		return models.QueueViewPayload{}, err
	}

	return models.QueueViewPayload{
		Type:          "queue_update",
		Action:        action,
		StoreID:       storeID,
		Date:          today.Format("02 Jan 2006"),
		CurrentCall:   currentCall,
		Counters:      counters,
		GuestCounters: guestCounters,
	}, nil
}

func buildQueueViewState(storeID int) (models.QueueViewCall, []models.QueueViewCounter, time.Time, error) {
	repo := &repositories.ViewQueueRepository{DB: config.DB}
	today := time.Now()

	currentCall, err := repo.GetCurrentCall(storeID, today)
	if err != nil {
		return models.QueueViewCall{}, nil, today, err
	}

	counters, err := repo.GetCountersStatus(storeID, today)
	if err != nil {
		return models.QueueViewCall{}, nil, today, err
	}

	currentTicket := strings.TrimSpace(currentCall.TicketNo)
	for i := range counters {
		counters[i].CounterLabel = formatCounterLabel(counters[i].CounterCode, i)
		if counters[i].TicketNo == "" {
			counters[i].TicketNo = "-"
		}
		isCurrent := currentTicket != "" && counters[i].TicketNo == currentTicket
		counters[i].IsCurrent = isCurrent
		counters[i].StatusLabel, counters[i].StatusClass = resolveQueueStatus(counters[i].TicketStatus, isCurrent)
	}

	if currentTicket == "" {
		for _, counter := range counters {
			if counter.TicketStatus == "CALLED" && counter.TicketNo != "-" {
				currentCall.TicketNo = counter.TicketNo
				currentCall.CounterLabel = counter.CounterLabel
				currentCall.CounterName = counter.CounterName
				currentCall.CategoryName = counter.CategoryName
				currentTicket = counter.TicketNo
				break
			}
		}
	} else {
		if strings.TrimSpace(currentCall.CounterLabel) == "" {
			for _, counter := range counters {
				if counter.TicketNo == currentTicket {
					currentCall.CounterLabel = counter.CounterLabel
					if strings.TrimSpace(currentCall.CounterName) == "" {
						currentCall.CounterName = counter.CounterName
					}
					currentCall.CategoryName = counter.CategoryName
					break
				}
			}
		} else {
			currentCall.CounterLabel = formatCounterLabel(currentCall.CounterLabel, -1)
		}
	}

	currentCall.CounterName = formatCounterName(currentCall.CounterName)
	return currentCall, counters, today, nil
}

func formatCounterName(name string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "loket") {
		return trimmed
	}
	return "Loket " + trimmed
}
