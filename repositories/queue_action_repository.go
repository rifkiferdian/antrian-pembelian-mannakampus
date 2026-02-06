package repositories

import (
	"database/sql"
	"time"
)

type QueueActionRepository struct {
	DB *sql.DB
}

func (r *QueueActionRepository) GetNextWaitingTicket(tx *sql.Tx, storeID, counterID int, ticketDate time.Time) (int64, string, error) {
	var (
		id       int64
		ticketNo string
	)
	err := tx.QueryRow(`
		SELECT id, ticket_no
		FROM queue_tickets
		WHERE store_id = ? AND counter_id = ? AND ticket_date = ? AND status IN ('WAITING', 'SKIPPED')
		ORDER BY
			CASE status
				WHEN 'WAITING' THEN 0
				WHEN 'SKIPPED' THEN 1
				ELSE 2
			END,
			queue_number ASC,
			id ASC
		LIMIT 1
		FOR UPDATE
	`, storeID, counterID, ticketDate.Format("2006-01-02")).Scan(&id, &ticketNo)
	if err != nil {
		return 0, "", err
	}
	return id, ticketNo, nil
}

func (r *QueueActionRepository) GetCurrentCalledTicket(tx *sql.Tx, storeID, counterID int, ticketDate time.Time) (int64, string, error) {
	var (
		id       int64
		ticketNo string
	)
	err := tx.QueryRow(`
		SELECT id, ticket_no
		FROM queue_tickets
		WHERE store_id = ? AND counter_id = ? AND ticket_date = ? AND status = 'CALLED'
		ORDER BY COALESCE(called_at, created_at) DESC, id DESC
		LIMIT 1
		FOR UPDATE
	`, storeID, counterID, ticketDate.Format("2006-01-02")).Scan(&id, &ticketNo)
	if err != nil {
		return 0, "", err
	}
	return id, ticketNo, nil
}

func (r *QueueActionRepository) MarkCalled(tx *sql.Tx, ticketID int64, userID int, calledAt time.Time) error {
	_, err := tx.Exec(`
		UPDATE queue_tickets
		SET status = 'CALLED',
			called_at = ?,
			called_by_user_id = ?,
			done_at = NULL,
			service_duration_seconds = NULL
		WHERE id = ?
	`, calledAt.Format("2006-01-02 15:04:05"), userID, ticketID)
	return err
}

func (r *QueueActionRepository) MarkDone(tx *sql.Tx, ticketID int64, doneAt time.Time) error {
	_, err := tx.Exec(`
		UPDATE queue_tickets
		SET status = 'DONE',
			done_at = ?,
			service_duration_seconds = CASE
				WHEN called_at IS NULL THEN 0
				ELSE GREATEST(TIMESTAMPDIFF(SECOND, called_at, ?), 0)
			END
		WHERE id = ?
	`, doneAt.Format("2006-01-02 15:04:05"), doneAt.Format("2006-01-02 15:04:05"), ticketID)
	return err
}

func (r *QueueActionRepository) MarkSkipped(tx *sql.Tx, ticketID int64, doneAt time.Time) error {
	_, err := tx.Exec(`
		UPDATE queue_tickets
		SET status = 'SKIPPED',
			done_at = ?,
			service_duration_seconds = CASE
				WHEN called_at IS NULL THEN 0
				ELSE GREATEST(TIMESTAMPDIFF(SECOND, called_at, ?), 0)
			END
		WHERE id = ?
	`, doneAt.Format("2006-01-02 15:04:05"), doneAt.Format("2006-01-02 15:04:05"), ticketID)
	return err
}

func (r *QueueActionRepository) InsertEvent(tx *sql.Tx, ticketID int64, eventType string, userID int, note string) error {
	_, err := tx.Exec(`
		INSERT INTO queue_events (ticket_id, event_type, user_id, note)
		VALUES (?, ?, ?, ?)
	`, ticketID, eventType, userID, note)
	return err
}
