package repositories

import (
	"database/sql"
	"fmt"
	"stok-hadiah/models"
	"time"
)

type QueueTicketRepository struct {
	DB *sql.DB
}

// CreateTicket membuat ticket baru untuk counter tertentu pada tanggal tertentu.
func (r *QueueTicketRepository) CreateTicket(storeID, counterID int, ticketPrefix string, ticketDate time.Time) (models.GuestTicket, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return models.GuestTicket{}, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var maxQueue int
	row := tx.QueryRow(`
		SELECT COALESCE(MAX(queue_number), 0)
		FROM queue_tickets
		WHERE store_id = ? AND counter_id = ? AND ticket_date = ?
		FOR UPDATE
	`, storeID, counterID, ticketDate.Format("2006-01-02"))
	if scanErr := row.Scan(&maxQueue); scanErr != nil {
		err = scanErr
		return models.GuestTicket{}, err
	}

	queueNumber := maxQueue + 1
	ticketNo := fmt.Sprintf("%s-%03d", ticketPrefix, queueNumber)

	res, execErr := tx.Exec(`
		INSERT INTO queue_tickets (store_id, counter_id, ticket_date, queue_number, ticket_no, status)
		VALUES (?, ?, ?, ?, ?, 'WAITING')
	`, storeID, counterID, ticketDate.Format("2006-01-02"), queueNumber, ticketNo)
	if execErr != nil {
		err = execErr
		return models.GuestTicket{}, err
	}

	ticketID, idErr := res.LastInsertId()
	if idErr != nil {
		err = idErr
		return models.GuestTicket{}, err
	}

	_, eventErr := tx.Exec(`
		INSERT INTO queue_events (ticket_id, event_type, note)
		VALUES (?, 'CREATE', 'Ambil nomor')
	`, ticketID)
	if eventErr != nil {
		err = eventErr
		return models.GuestTicket{}, err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		err = commitErr
		return models.GuestTicket{}, err
	}

	return models.GuestTicket{
		ID:          ticketID,
		StoreID:     storeID,
		CounterID:   counterID,
		TicketNo:    ticketNo,
		QueueNumber: queueNumber,
		TicketDate:  ticketDate.Format("2006-01-02"),
	}, nil
}

// GetTicketByID mengambil data tiket lengkap untuk halaman print.
func (r *QueueTicketRepository) GetTicketByID(id int64) (models.GuestTicket, error) {
	var ticket models.GuestTicket
	var ticketDate time.Time
	var createdAt time.Time

	err := r.DB.QueryRow(`
		SELECT
			qt.id,
			qt.store_id,
			COALESCE(s.store_name, '') AS store_name,
			qt.counter_id,
			COALESCE(c.counter_name, '') AS counter_name,
			COALESCE(c.counter_code, '') AS counter_code,
			COALESCE(sc.category_name, '') AS category_name,
			qt.ticket_no,
			qt.queue_number,
			qt.ticket_date,
			qt.created_at
		FROM queue_tickets qt
		LEFT JOIN counters c ON c.id = qt.counter_id
		LEFT JOIN stores s ON s.store_id = qt.store_id
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE qt.id = ?
	`, id).Scan(
		&ticket.ID,
		&ticket.StoreID,
		&ticket.StoreName,
		&ticket.CounterID,
		&ticket.CounterName,
		&ticket.CounterCode,
		&ticket.CategoryName,
		&ticket.TicketNo,
		&ticket.QueueNumber,
		&ticketDate,
		&createdAt,
	)
	if err != nil {
		return models.GuestTicket{}, err
	}

	if ticket.CategoryName == "" {
		ticket.CategoryName = ticket.CounterName
	}

	ticket.TicketDate = ticketDate.Format("2006-01-02")
	ticket.CreatedAt = createdAt.Format("2006-01-02 15:04:05")

	return ticket, nil
}
