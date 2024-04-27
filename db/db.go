package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

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
	db.Db.Close()
}

func (db Db) CreateUser(email string, first_name string, last_name *string, password string) (string, error) {
	tx, err := db.Db.Begin()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	var user_id string
	err = tx.QueryRow("INSERT INTO single_user DEFAULT VALUES RETURNING id").Scan(&user_id)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return "", err
	}
	var pii_id string
	err = tx.QueryRow("INSERT INTO user_pii (user_id, email, first_name, last_name, password) VALUES ($1, $2, $3, $4, $5) RETURNING id", user_id, email, first_name, last_name, password).Scan(&pii_id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			tx.Rollback()
			return "", fmt.Errorf("Email already exists")
		}
		fmt.Println(err)
		tx.Rollback()
		return "", err
	}
	err = tx.QueryRow("UPDATE single_user SET pii_id = $1 WHERE id = $2 RETURNING id", pii_id, user_id).Scan(&user_id)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return user_id, nil
}

func (db Db) CreateProject(user_id string, name string, description *string) (string, error) {
	tx, err := db.Db.Begin()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	var project_id string
	err = tx.QueryRow("INSERT INTO project (user_id, name, description) VALUES ($1, $2, $3) RETURNING id", user_id, name, description).Scan(&project_id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			tx.Rollback()
			return "", fmt.Errorf("Project already exists")
		}
		fmt.Println(err)
		tx.Rollback()
		return "", err
	}
	_, err = tx.Exec("INSERT INTO permitted_project (user_id, project_id) VALUES ($1, $2)", user_id, project_id)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return project_id, nil
}
