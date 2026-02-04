package repositories

import (
	"database/sql"
	"stok-hadiah/models"
	"time"
)

type GuestQueueRepository struct {
	DB *sql.DB
}

// GetCountersForGuest mengambil daftar counter aktif beserta kategori dan jumlah antrian menunggu.
func (r *GuestQueueRepository) GetCountersForGuest(storeID int, ticketDate time.Time) ([]models.GuestQueueCounter, error) {
	rows, err := r.DB.Query(`
		SELECT
			c.id,
			c.counter_name,
			c.counter_code,
			c.ticket_prefix,
			COALESCE(sc.category_name, '') AS category_name,
			(
				SELECT COUNT(*)
				FROM queue_tickets qt
				WHERE qt.store_id = c.store_id
				  AND qt.counter_id = c.id
				  AND qt.status = 'WAITING'
				  AND qt.ticket_date = ?
			) AS waiting_count
		FROM counters c
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE c.store_id = ? AND c.is_active = 1
		ORDER BY c.id ASC
	`, ticketDate.Format("2006-01-02"), storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counters []models.GuestQueueCounter
	for rows.Next() {
		var c models.GuestQueueCounter
		if err := rows.Scan(
			&c.CounterID,
			&c.CounterName,
			&c.CounterCode,
			&c.TicketPrefix,
			&c.CategoryName,
			&c.WaitingCount,
		); err != nil {
			return nil, err
		}
		if c.CategoryName == "" {
			c.CategoryName = c.CounterName
		}
		counters = append(counters, c)
	}

	return counters, rows.Err()
}
