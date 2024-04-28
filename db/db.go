package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jesses-code-adventures/every_log/error_msgs"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type dbCredentials struct {
	Name     string
	User     string
	Password string
	Host     string
	Port     string
}

func getDbCredentials() dbCredentials {
	godotenv.Load()
	return dbCredentials{
		Name:     os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
	}
}

type Db struct {
	Db *sql.DB
}

func NewDb() Db {
	credentials := getDbCredentials()
	db, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", credentials.User, credentials.Password, credentials.Name))
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("SET search_path TO public;")
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return Db{db}
}

func (db Db) Close() {
	err := db.Db.Close()
	if err != nil {
		fmt.Println("failed to close db!")
		panic(err)
	}
}

func (db Db) CreateUser(email string, first_name string, last_name *string, password string) (string, error) {
	tx, err := db.Db.Begin()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	var user_id string
	err = tx.QueryRow("INSERT INTO single_user DEFAULT VALUES RETURNING id").Scan(&user_id)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
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
				fmt.Println(innerErr) // TODO: Use a logger
				return "", errors.New(error_msgs.DATABASE_ERROR)
			}
			return "", errors.New(error_msgs.EMAIL_EXISTS)
		}
		fmt.Println(err)
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", err
	}
	err = tx.QueryRow("UPDATE single_user SET pii_id = $1 WHERE id = $2 RETURNING id", pii_id, user_id).Scan(&user_id)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	return user_id, nil
}

func (db Db) CreateProject(user_id string, name string, description *string) (string, error) {
	tx, err := db.Db.Begin()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	var project_id string
	row := tx.QueryRow("INSERT INTO project (user_id, name, description) VALUES ($1, $2, $3) RETURNING id", user_id, name, description)
	err = row.Scan(&project_id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			innerErr := tx.Rollback()
			if innerErr != nil {
				fmt.Println(innerErr) // TODO: Use a logger
				return "", errors.New(error_msgs.DATABASE_ERROR)
			}
			return "", errors.New(error_msgs.PROJECT_EXISTS)
		}
		fmt.Println(err) // TODO: Use a logger
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		fmt.Println("got to 4")
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	_, err = tx.Exec("INSERT INTO permitted_project (user_id, project_id) VALUES ($1, $2)", user_id, project_id)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	return project_id, nil
}
