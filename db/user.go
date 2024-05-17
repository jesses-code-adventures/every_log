package db

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jesses-code-adventures/every_log/error_msgs"
)

func (db Db) CreateUser(email string, first_name string, last_name *string, password string) (string, error) {
	tx, err := db.Db.Begin()
	if err != nil {
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	var user_id string
	err = tx.QueryRow("INSERT INTO single_user DEFAULT VALUES RETURNING id").Scan(&user_id)
	if err != nil {
		db.Logger.Println(err)
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	var pii_id string
	err = tx.QueryRow("INSERT INTO user_pii (user_id, email, first_name, last_name, password) VALUES ($1, $2, $3, $4, $5) RETURNING id", user_id, email, first_name, last_name, password).Scan(&pii_id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			fmt.Println(error_msgs.EMAIL_EXISTS)
			innerErr := tx.Rollback()
			if innerErr != nil {
				db.Logger.Println(innerErr)
				return "", errors.New(error_msgs.DATABASE_ERROR)
			}
			return "", errors.New(error_msgs.EMAIL_EXISTS)
		}
		fmt.Println(err)
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", err
	}
	err = tx.QueryRow("UPDATE single_user SET pii_id = $1 WHERE id = $2 RETURNING id", pii_id, user_id).Scan(&user_id)
	if err != nil {
		db.Logger.Println(err)
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	err = tx.Commit()
	if err != nil {
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	return user_id, nil
}
