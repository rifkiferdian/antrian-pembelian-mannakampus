package repositories

import (
	"database/sql"
	"stok-hadiah/models"
	"time"
)

type ViewQueueRepository struct {
	DB *sql.DB
}

// GetCurrentCall mengambil tiket terakhir berstatus CALLED untuk store dan tanggal tertentu.
func (r *ViewQueueRepository) GetCurrentCall(storeID int, ticketDate time.Time) (models.QueueViewCall, error) {
	var call models.QueueViewCall
	var counterCode sql.NullString
	var counterName sql.NullString
	var categoryName sql.NullString

	err := r.DB.QueryRow(`
        SELECT
            qt.ticket_no,
            COALESCE(c.counter_code, '') AS counter_code,
            COALESCE(c.counter_name, '') AS counter_name,
            COALESCE(sc.category_name, '') AS category_name
        FROM queue_tickets qt
        LEFT JOIN counters c ON c.id = qt.counter_id
        LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
        WHERE qt.store_id = ? AND qt.ticket_date = ? AND qt.status = 'CALLED'
        ORDER BY qt.called_at DESC, qt.id DESC
        LIMIT 1
    `, storeID, ticketDate.Format("2006-01-02")).Scan(
		&call.TicketNo,
		&counterCode,
		&counterName,
		&categoryName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.QueueViewCall{}, nil
		}
		return models.QueueViewCall{}, err
	}

	call.CounterLabel = counterCode.String
	call.CounterName = counterName.String
	call.CategoryName = categoryName.String
	return call, nil
}

// GetCountersStatus mengambil status tiket terbaru per loket untuk tanggal tertentu.
func (r *ViewQueueRepository) GetCountersStatus(storeID int, ticketDate time.Time) ([]models.QueueViewCounter, error) {
	rows, err := r.DB.Query(`
        SELECT
            c.id,
            c.counter_name,
            c.counter_code,
            COALESCE(sc.category_name, '') AS category_name,
            (
                SELECT qt.ticket_no
                FROM queue_tickets qt
                WHERE qt.store_id = c.store_id
                  AND qt.counter_id = c.id
                  AND qt.ticket_date = ?
                  AND qt.status IN ('CALLED', 'WAITING', 'DONE', 'SKIPPED', 'CANCELLED')
                ORDER BY
                    CASE qt.status
                        WHEN 'CALLED' THEN 0
                        WHEN 'WAITING' THEN 1
                        WHEN 'DONE' THEN 2
                        WHEN 'SKIPPED' THEN 3
                        WHEN 'CANCELLED' THEN 4
                        ELSE 5
                    END,
                    COALESCE(qt.called_at, qt.created_at, qt.done_at) DESC,
                    qt.id DESC
                LIMIT 1
            ) AS ticket_no,
            (
                SELECT qt.status
                FROM queue_tickets qt
                WHERE qt.store_id = c.store_id
                  AND qt.counter_id = c.id
                  AND qt.ticket_date = ?
                  AND qt.status IN ('CALLED', 'WAITING', 'DONE', 'SKIPPED', 'CANCELLED')
                ORDER BY
                    CASE qt.status
                        WHEN 'CALLED' THEN 0
                        WHEN 'WAITING' THEN 1
                        WHEN 'DONE' THEN 2
                        WHEN 'SKIPPED' THEN 3
                        WHEN 'CANCELLED' THEN 4
                        ELSE 5
                    END,
                    COALESCE(qt.called_at, qt.created_at, qt.done_at) DESC,
                    qt.id DESC
                LIMIT 1
            ) AS ticket_status
        FROM counters c
        LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
        WHERE c.store_id = ? AND c.is_active = 1
        ORDER BY c.id ASC
    `, ticketDate.Format("2006-01-02"), ticketDate.Format("2006-01-02"), storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counters []models.QueueViewCounter
	for rows.Next() {
		var c models.QueueViewCounter
		var ticketNo sql.NullString
		var ticketStatus sql.NullString
		if err := rows.Scan(
			&c.CounterID,
			&c.CounterName,
			&c.CounterCode,
			&c.CategoryName,
			&ticketNo,
			&ticketStatus,
		); err != nil {
			return nil, err
		}
		c.TicketNo = ticketNo.String
		c.TicketStatus = ticketStatus.String
		if c.CategoryName == "" {
			c.CategoryName = c.CounterName
		}
		counters = append(counters, c)
	}

	return counters, rows.Err()
}
