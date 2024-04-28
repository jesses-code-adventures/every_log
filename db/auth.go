package db

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jesses-code-adventures/every_log/error_msgs"
)

func (db Db) Authenticate(user_id string, password string, tx *sql.Tx) error {
	var storedPassword string
	var err error
	if tx != nil {
		err = tx.QueryRow("SELECT password FROM user_pii WHERE user_id = $1", user_id).Scan(&storedPassword)
	} else {
		err = db.Db.QueryRow("SELECT password FROM user_pii WHERE user_id = $1", user_id).Scan(&storedPassword)
	}
	if err != nil {
		fmt.Println(err)
		return errors.New(error_msgs.DATABASE_ERROR)
	}
	// TODO: Implement password hashing
	if storedPassword != password {
		err = errors.New(error_msgs.UNAUTHORIZED)
		fmt.Println(err)
		return err
	}
	return nil
}

func (db Db) Authorize(user_id string, token string, tx *sql.Tx) error {
	var storedToken string
	var err error
	if tx != nil {
		err = tx.QueryRow("SELECT token FROM single_user WHERE id = $1", user_id).Scan(&storedToken)
	} else {
		err = db.Db.QueryRow("SELECT token FROM single_user WHERE id = $1", user_id).Scan(&storedToken)
	}
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return errors.New(error_msgs.DATABASE_ERROR)
	}
	if storedToken != token {
		err = errors.New(error_msgs.UNAUTHORIZED)
		fmt.Println(err)
		return err
	}
	return nil
}

// Pass nil as tx to execute with the default db connection
// Pass a transaction to execute within that transaction
func (db Db) UpdateUserToken(user_id string, token string, tx *sql.Tx) error {
	var err error
	if tx != nil {
		_, err = tx.Exec("UPDATE single_user SET token = $1 WHERE id = $2", token, user_id)
	} else {

		_, err = db.Db.Exec("UPDATE single_user SET token = $1 WHERE id = $2", token, user_id)
	}
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return errors.New(error_msgs.DATABASE_ERROR)
	}
	return nil
}
