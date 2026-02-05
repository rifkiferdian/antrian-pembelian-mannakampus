package repositories

import (
	"database/sql"
	"time"

	"stok-hadiah/models"
)

type DashboardRepository struct {
	DB *sql.DB
}

// GetServingTicket mengambil tiket yang sedang dipanggil untuk counter tertentu.
func (r *DashboardRepository) GetServingTicket(storeID, counterID int, ticketDate time.Time) (models.DashboardServing, error) {
	var (
		serving      models.DashboardServing
		calledAt     sql.NullTime
		categoryName sql.NullString
		counterName  sql.NullString
	)

	err := r.DB.QueryRow(`
		SELECT
			qt.ticket_no,
			qt.called_at,
			COALESCE(sc.category_name, '') AS category_name,
			COALESCE(c.counter_name, '') AS counter_name
		FROM queue_tickets qt
		LEFT JOIN counters c ON c.id = qt.counter_id
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE qt.store_id = ? AND qt.counter_id = ? AND qt.ticket_date = ? AND qt.status = 'CALLED'
		ORDER BY qt.called_at DESC, qt.id DESC
		LIMIT 1
	`, storeID, counterID, ticketDate.Format("2006-01-02")).Scan(
		&serving.TicketNo,
		&calledAt,
		&categoryName,
		&counterName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.DashboardServing{}, nil
		}
		return models.DashboardServing{}, err
	}

	serving.CategoryName = categoryName.String
	serving.CounterName = counterName.String
	if calledAt.Valid {
		serving.CalledAt = calledAt.Time.Format("2006-01-02 15:04:05")
		serving.CalledAtUnix = calledAt.Time.Unix()
	}

	return serving, nil
}

// GetWaitingTickets mengambil daftar antrian WAITING untuk counter tertentu (beserta total).
func (r *DashboardRepository) GetWaitingTickets(storeID, counterID int, ticketDate time.Time, limit int) ([]models.DashboardQueueItem, int, error) {
	var total int
	if err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM queue_tickets
		WHERE store_id = ? AND counter_id = ? AND ticket_date = ? AND status = 'WAITING'
	`, storeID, counterID, ticketDate.Format("2006-01-02")).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.DB.Query(`
		SELECT
			qt.ticket_no,
			qt.queue_number,
			qt.created_at,
			COALESCE(sc.category_name, '') AS category_name,
			COALESCE(c.counter_name, '') AS counter_name
		FROM queue_tickets qt
		LEFT JOIN counters c ON c.id = qt.counter_id
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE qt.store_id = ? AND qt.counter_id = ? AND qt.ticket_date = ? AND qt.status = 'WAITING'
		ORDER BY qt.queue_number ASC, qt.id ASC
		LIMIT ?
	`, storeID, counterID, ticketDate.Format("2006-01-02"), limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	now := time.Now()
	items := make([]models.DashboardQueueItem, 0)
	for rows.Next() {
		var (
			item        models.DashboardQueueItem
			createdAt   time.Time
			category    sql.NullString
			counterName sql.NullString
		)
		if err := rows.Scan(
			&item.TicketNo,
			&item.QueueNumber,
			&createdAt,
			&category,
			&counterName,
		); err != nil {
			return nil, 0, err
		}

		item.CategoryName = category.String
		if item.CategoryName == "" {
			item.CategoryName = counterName.String
		}

		waitMinutes := int(now.Sub(createdAt).Minutes())
		if waitMinutes < 0 {
			waitMinutes = 0
		}
		item.WaitingMinutes = waitMinutes

		items = append(items, item)
	}

	return items, total, rows.Err()
}
