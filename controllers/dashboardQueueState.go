package controllers

import (
	"fmt"
	"time"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"
)

func buildDashboardQueueState(counter models.Counter) (models.DashboardServing, []models.DashboardQueueItem, int, error) {
	repo := &repositories.DashboardRepository{DB: config.DB}
	today := time.Now()

	serving, err := repo.GetServingTicket(counter.StoreID, counter.ID, today)
	if err != nil {
		return models.DashboardServing{}, nil, 0, err
	}

	if serving.CategoryName == "" {
		serving.CategoryName = counter.CounterName
	}
	if serving.CounterName == "" {
		serving.CounterName = counter.CounterName
	}

	serving.DurationLabel = formatServiceDuration(serving.CalledAtUnix)

	waitingItems, waitingTotal, err := repo.GetWaitingTickets(counter.StoreID, counter.ID, today, 5)
	if err != nil {
		return models.DashboardServing{}, nil, 0, err
	}

	for i := range waitingItems {
		label, className := resolveQueueStatus(waitingItems[i].Status, false)
		waitingItems[i].StatusLabel = label
		waitingItems[i].StatusClass = className
	}

	return serving, waitingItems, waitingTotal, nil
}

func formatServiceDuration(calledAtUnix int64) string {
	if calledAtUnix <= 0 {
		return "00:00"
	}

	seconds := int(time.Since(time.Unix(calledAtUnix, 0)).Seconds())
	if seconds < 0 {
		seconds = 0
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	}
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
