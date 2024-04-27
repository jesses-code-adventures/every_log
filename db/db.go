package db

import (
	"database/sql"
	"fmt"
	"os"
	"encoding/json"

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
	var pii_id string;
	err = tx.QueryRow("INSERT INTO user_pii (user_id, email, first_name, last_name, password) VALUES ($1, $2, $3, $4, $5) RETURNING id", user_id, email, first_name, last_name, password).Scan(&pii_id)
	if err != nil {
		if err.Error() == "pq: duplicate key value violates unique constraint \"user_pii_email_key\"" {
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

func (db Db) CheckTableExists(name string) ([]byte, error) {
	query := fmt.Sprintf("SELECT table_name FROM information_schema.tables WHERE table_name = '%s';", name)
	var table_name string
	err := db.Db.QueryRow(query).Scan(&table_name)
	if err != nil {
		return []byte{}, err
	}
	resp := struct {
		Exists bool `json:"exists"`
	} {
		Exists: table_name == name,
	}
	return json.Marshal(resp)
}

func (db Db) ShowTables() ([]byte, error) {
	query := `SELECT c.relname as "Name"
FROM pg_catalog.pg_class c
    LEFT JOIN pg_catalog.pg_namespace n ON n.oid = c.relnamespace
WHERE c.relkind IN ('r','')
    AND n.nspname <> 'pg_catalog'
    AND n.nspname <> 'information_schema'
    AND n.nspname !~ '^pg_toast'
AND pg_catalog.pg_table_is_visible(c.oid)
ORDER BY 1;`
	rows, err := db.Db.Query(query)
	resp := []byte{}
	names := []string{}
	if err != nil {
		fmt.Println(err)
		return resp, err
	}
	defer rows.Close()
	for rows.Next() {
		var table_name string
		err := rows.Scan(&table_name)
		if err != nil {
			fmt.Println(err)
		}
		names = append(names, table_name)
	}
	resp, err = json.Marshal(names)
	if err != nil {
		fmt.Println(err)
		return resp, err
	}
	return resp, nil
}
