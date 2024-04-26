package db

import (
	"database/sql"
	"fmt"
	"os"
	"time"

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
	return Db{db}
}

func (db Db) Close() {
	db.Db.Close()
}

func (db Db) CreateUser(email string, first_name string, last_name *string, password string) (string, string, error) {
	result := db.Db.QueryRow("INSERT INTO user (email, first_name, last_name, password) VALUES ($1, $2, $3, $4) RETURNING id, created_at", email, first_name, last_name, password)
	var id string;
	var created_at time.Time;
	err := result.Scan(&id, &created_at)
	if err != nil {
		return "", "", err
	}
	return id, nil
}
