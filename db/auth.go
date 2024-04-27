package db

import (
	"fmt"
	"database/sql"
)

func (db Db) Authenticate(user_id string, password string) (bool, error) {
	var storedPassword string
	err := db.Db.QueryRow("SELECT password FROM user_pii WHERE user_id = $1", user_id).Scan(&storedPassword)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	// TODO: Implement password hashing
	return storedPassword == password, nil
}

func (db Db) Authorize(user_id string, token string) (bool, error) {
	var storedToken string
	err := db.Db.QueryRow("SELECT token FROM single_user WHERE id = $1", user_id).Scan(&storedToken)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return storedToken == token, nil
}

// Pass nil as tx to execute with the default db connection
// Pass a transaction to execute within that transaction
func (db Db) UpdateUserToken(user_id string, token string, tx *sql.Tx) bool {
	if tx != nil {
		_, err := tx.Exec("UPDATE single_user SET token = $1 WHERE id = $2", token, user_id)
		if err != nil {
			fmt.Println(err)
			return false
		}
		return true
	}
	_, err := db.Db.Exec("UPDATE single_user SET token = $1 WHERE id = $2", token, user_id)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}
