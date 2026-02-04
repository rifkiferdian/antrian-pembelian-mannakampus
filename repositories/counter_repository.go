package repositories

import (
	"database/sql"
	"stok-hadiah/models"
)

type CounterRepository struct {
	DB *sql.DB
}

// GetAll mengambil seluruh data counter beserta nama store.
func (r *CounterRepository) GetAll() ([]models.Counter, error) {
	rows, err := r.DB.Query(`
		SELECT 
			c.id,
			c.store_id,
			COALESCE(s.store_name, '') AS store_name,
			c.counter_code,
			c.counter_name,
			c.ticket_prefix,
			c.is_active,
			c.created_at
		FROM counters c
		LEFT JOIN stores s ON s.store_id = c.store_id
		ORDER BY c.created_at DESC, c.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counters []models.Counter

	for rows.Next() {
		var (
			counter   models.Counter
			isActive  int
			createdAt sql.NullTime
		)

		if err := rows.Scan(
			&counter.ID,
			&counter.StoreID,
			&counter.StoreName,
			&counter.CounterCode,
			&counter.CounterName,
			&counter.TicketPrefix,
			&isActive,
			&createdAt,
		); err != nil {
			return nil, err
		}

		counter.IsActive = isActive == 1
		if counter.IsActive {
			counter.StatusLabel = "Aktif"
		} else {
			counter.StatusLabel = "Non Aktif"
		}

		if createdAt.Valid {
			counter.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			counter.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			counter.CreatedAt = "-"
			counter.CreatedAtDisplay = "-"
		}

		if counter.StoreName == "" {
			counter.StoreName = "-"
		}

		counters = append(counters, counter)
	}

	return counters, rows.Err()
}

// Create menyimpan data counter baru.
func (r *CounterRepository) Create(params models.CounterCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO counters (store_id, counter_code, counter_name, ticket_prefix, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, params.StoreID, params.CounterCode, params.CounterName, params.TicketPrefix, boolToInt(params.IsActive))
	return err
}

// Update memperbarui data counter.
func (r *CounterRepository) Update(params models.CounterUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE counters
		SET store_id = ?, counter_code = ?, counter_name = ?, ticket_prefix = ?, is_active = ?
		WHERE id = ?
	`, params.StoreID, params.CounterCode, params.CounterName, params.TicketPrefix, boolToInt(params.IsActive), params.ID)
	return err
}

// Delete menghapus counter berdasarkan ID.
func (r *CounterRepository) Delete(id int) error {
	_, err := r.DB.Exec(`DELETE FROM counters WHERE id = ?`, id)
	return err
}

func boolToInt(val bool) int {
	if val {
		return 1
	}
	return 0
}
