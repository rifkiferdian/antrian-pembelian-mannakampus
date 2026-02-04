package repositories

import (
	"database/sql"
	"stok-hadiah/models"
	"strings"
)

type StoreRepository struct {
	DB *sql.DB
}

// GetFirstActive mengambil store aktif pertama.
func (r *StoreRepository) GetFirstActive() (models.Store, error) {
	var store models.Store
	err := r.DB.QueryRow(`
		SELECT store_id, store_name
		FROM stores
		WHERE is_active = 1
		ORDER BY store_id ASC
		LIMIT 1
	`).Scan(&store.StoreID, &store.StoreName)
	if err != nil {
		return models.Store{}, err
	}

	return store, nil
}

// GetByID mengambil store berdasarkan ID.
func (r *StoreRepository) GetByID(id int) (models.Store, error) {
	var store models.Store
	err := r.DB.QueryRow(`
		SELECT store_id, store_name
		FROM stores
		WHERE store_id = ?
		LIMIT 1
	`, id).Scan(&store.StoreID, &store.StoreName)
	if err != nil {
		return models.Store{}, err
	}

	return store, nil
}

// GetAll mengambil seluruh data store.
func (r *StoreRepository) GetAll() ([]models.Store, error) {
	rows, err := r.DB.Query(`
		SELECT store_id, store_name
		FROM stores
		ORDER BY store_id asc
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var s models.Store
		if err := rows.Scan(&s.StoreID, &s.StoreName); err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}

	return stores, rows.Err()
}

// GetByIDs mengambil daftar store berdasarkan id yang diberikan.
func (r *StoreRepository) GetByIDs(ids []int) ([]models.Store, error) {
	if len(ids) == 0 {
		return []models.Store{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT store_id, store_name
		FROM stores
		WHERE store_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY store_name
	`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var s models.Store
		if err := rows.Scan(&s.StoreID, &s.StoreName); err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}

	return stores, nil
}
