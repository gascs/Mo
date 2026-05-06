package service

import (
	"mo/internal/database"
)

// GetAllSettings returns all key-value pairs from the settings table.
func GetAllSettings() (map[string]string, error) {
	rows, err := database.DB.Query("SELECT key, value FROM settings")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			continue
		}
		result[k] = v
	}
	return result, nil
}

// UpdateSettings applies a batch of key-value updates.
func UpdateSettings(updates map[string]string) error {
	tx, err := database.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k, v := range updates {
		if _, err := stmt.Exec(k, v); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetSetting returns a single setting value.
func GetSetting(key string) (string, error) {
	var v string
	err := database.DB.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&v)
	return v, err
}
