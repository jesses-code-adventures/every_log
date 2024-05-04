package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jesses-code-adventures/every_log/error_msgs"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Log struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UserId    string    `json:"user_id"`
	ProjectId string    `json:"project_id"`
	LevelId   int       `json:"level_id"`
	ProcessId *string   `json:"process_id"`
	Message   *string   `json:"message"`
	Traceback *string   `json:"traceback"`
}

type Org struct {
	Id          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	LocationId  *string   `json:"location_id"`
}

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

func (db Db) CreateOrg(userId string, name string, description *string, location_id *string) (string, error) {
	var orgId string
	tx, err := db.Db.Begin()
	query := "INSERT INTO org (name, owner"
	values := "VALUES ($1, $2"
	idx := 3
	args := make([]any, 0)
	args = append(args, name, userId)
	if description != nil {
		query += ", description"
		values += ", $" + fmt.Sprint(idx)
		args = append(args, *description)
		idx++
	}
	if location_id != nil {
		query += ", location_id"
		values += ", $" + fmt.Sprint(idx)
		args = append(args, *location_id)
		idx++
	}
	query += fmt.Sprintf(") %s) RETURNING id", values)
	row := tx.QueryRow(query, args...)
	err = row.Scan(&orgId)
	if err != nil {
		var respErr error
		if strings.Contains(err.Error(), "duplicate") {
			respErr = errors.New(error_msgs.ORG_EXISTS)
		} else {
			respErr = errors.New(error_msgs.DATABASE_ERROR)
		}
		fmt.Println(err) // TODO: Use a logger
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", respErr
	}
	query = "INSERT INTO user_org (user_id, org_id, level) VALUES ($1, $2, 500)"
	_, err = tx.Exec(query, userId, orgId)
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
	return orgId, nil
}

func (db Db) GetOrgs(userId string, orgId *string, name *string, from *time.Time, to *time.Time) ([]Org, error) {
	var orgs []Org
	tx, err := db.Db.Begin()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger,
		return nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	var rows *sql.Rows
	query := "SELECT org.id, org.created_at, org.name, org.description, org.location_id FROM org INNER JOIN user_org WHERE user_org.user_id = $1"
	variableIndex := 2
	args := make([]any, 0)
	args = append(args, userId)
	// I kind of hate this
	if orgId != nil {
		query += fmt.Sprintf(" AND project_id = $%d", variableIndex)
		args = append(args, *orgId)
		variableIndex++
	}
	if name != nil {
		query += fmt.Sprintf(" AND name LIKE $%d", variableIndex)
		args = append(args, *name)
		variableIndex++
	}
	if from != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", variableIndex)
		args = append(args, *from)
		variableIndex++
	}
	if to != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", variableIndex)
		args = append(args, *to)
		variableIndex++
	}
	rows, err = tx.Query(query, args...)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return nil, errors.New(error_msgs.DATABASE_ERROR)
		}
		return nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	defer rows.Close()
	for rows.Next() {
		var org Org
		err = rows.Scan(&org.Id, &org.CreatedAt, &org.Name, &org.Description, &org.LocationId)
		if err != nil {
			fmt.Println(err) // TODO: Use a logger
			innerErr := tx.Rollback()
			if innerErr != nil {
				fmt.Println(innerErr) // TODO: Use a logger
				return nil, errors.New(error_msgs.DATABASE_ERROR)
			}
			return nil, errors.New(error_msgs.DATABASE_ERROR)
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}

func (db Db) CreateLog(userId string, project_id string, level_id int, process_id *string, message string, traceback *string, apiKey string) (string, error) {
	var logId string
	tx, err := db.Db.Begin()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	_, err = db.getPermittedProjectIdFromApiKey(userId, apiKey)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", err
	}
	row := tx.QueryRow("INSERT INTO log (user_id, project_id, level_id, process_id, message, traceback) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", userId, project_id, level_id, process_id, message, traceback)
	err = row.Scan(&logId)
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
	return logId, nil
}

func (db Db) getPermittedProjectIdFromUserProject(userId string, project_id string) (string, error) {
	var permittedProjectId string
	err := db.Db.QueryRow(`SELECT permitted_project_id
FROM api_key
LEFT JOIN permitted_project
ON permitted_project.id = api_key.permitted_project_id
WHERE user_id = $1
AND permitted_project.project_id = $2`, userId, project_id).Scan(&permittedProjectId)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.UNAUTHORIZED)
	}
	return permittedProjectId, nil
}

func (db Db) GetLogs(userId string, projectId *string, levelId *int, processId *string, orgId *string, from *time.Time, to *time.Time) ([]Log, error) {
	var logs []Log
	tx, err := db.Db.Begin()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger,
		return nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	var rows *sql.Rows
	query := "SELECT * FROM log WHERE user_id = $1"
	variableIndex := 2
	args := make([]any, 0)
	args = append(args, userId)
	// I kind of hate this
	if projectId != nil {
		query += fmt.Sprintf(" AND project_id = $%d", variableIndex)
		args = append(args, *projectId)
		variableIndex++
	}
	if levelId != nil {
		query += fmt.Sprintf(" AND level_id = $%d", variableIndex)
		args = append(args, *levelId)
		variableIndex++
	}
	if processId != nil {
		query += fmt.Sprintf(" AND process_id = $%d", variableIndex)
		args = append(args, *processId)
		variableIndex++
	}
	if orgId != nil {
		query += fmt.Sprintf(" AND org_id = $%d", variableIndex)
		args = append(args, *orgId)
		variableIndex++
	}
	if from != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", variableIndex)
		args = append(args, *from)
		variableIndex++
	}
	if to != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", variableIndex)
		args = append(args, *to)
		variableIndex++
	}
	rows, err = tx.Query(query, args...)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return nil, errors.New(error_msgs.DATABASE_ERROR)
		}
		return nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	defer rows.Close()
	for rows.Next() {
		var log Log
		err = rows.Scan(&log.Id, &log.CreatedAt, &log.UserId, &log.ProjectId, &log.LevelId, &log.ProcessId, &log.Message, &log.Traceback)
		if err != nil {
			fmt.Println(err) // TODO: Use a logger
			innerErr := tx.Rollback()
			if innerErr != nil {
				fmt.Println(innerErr) // TODO: Use a logger
				return nil, errors.New(error_msgs.DATABASE_ERROR)
			}
			return nil, errors.New(error_msgs.DATABASE_ERROR)
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (db Db) getPermittedProjectId(userId string, projectId string, tx *sql.Tx) (string, error) {
	var id string
	var err error
	if tx != nil {
		err = tx.QueryRow("SELECT id FROM permitted_project WHERE user_id = $1 AND project_id = $2", userId, projectId).Scan(&id)
	} else {
		err = db.Db.QueryRow("SELECT id FROM permitted_project WHERE user_id = $1 AND project_id = $2", userId, projectId).Scan(&id)
	}
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	return id, nil
}

func (db Db) getPermittedProjectIdFromApiKey(userId string, apiKey string) (string, error) {
	var id string
	var matchingUserId string
	err := db.Db.QueryRow(`SELECT api_key.permitted_project_id, permitted_project.user_id
FROM api_key
LEFT JOIN permitted_project
ON permitted_project.id = api_key.permitted_project_id
WHERE api_key.key = $1;`, apiKey).Scan(&id, &matchingUserId)
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	if matchingUserId != userId {
		fmt.Println("User ID does not match") // TODO: Use a logger
		return "", errors.New(error_msgs.UNAUTHORIZED)
	}
	return id, nil
}

// GenerateRandomAPIKey generates a random alphanumeric API key of the given length
func GenerateRandomAPIKey(length int) (string, error) {
	bytes := make([]byte, length/2) // Using hex encoding, so each byte gives two characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (db Db) CreateApiKey(userId string, projectId string) (string, error) {
	tx, err := db.Db.Begin()
	if err != nil {
		fmt.Println(err) // TODO: Use a logger
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	permittedId, err := db.getPermittedProjectId(userId, projectId, tx)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", errors.New(error_msgs.UNAUTHORIZED)
	}
	apiKey, err := GenerateRandomAPIKey(32)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			fmt.Println(innerErr) // TODO: Use a logger
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}

	_, err = tx.Exec("INSERT INTO api_key (permitted_project_id, key) VALUES ($1, $2) ON CONFLICT(permitted_project_id) DO UPDATE SET key=EXCLUDED.key", permittedId, apiKey)

	if err != nil {
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
	return apiKey, nil

}
