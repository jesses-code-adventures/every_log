package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/jesses-code-adventures/every_log/error_msgs"
)

func (db Db) getPermittedProjectIdFromUserProject(userId string, project_id string) (string, error) {
	var permittedProjectId string
	err := db.Db.QueryRow(`SELECT permitted_project_id
FROM api_key
LEFT JOIN permitted_project
ON permitted_project.id = api_key.permitted_project_id
WHERE user_id = $1
AND permitted_project.project_id = $2`, userId, project_id).Scan(&permittedProjectId)
	if err != nil {
		db.Logger.Println(err)
		return "", errors.New(error_msgs.UNAUTHORIZED)
	}
	return permittedProjectId, nil
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
		db.Logger.Println(err)
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
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	if matchingUserId != userId {
		db.Logger.Println("User ID does not match")
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
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	permittedId, err := db.getPermittedProjectId(userId, projectId, tx)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", errors.New(error_msgs.UNAUTHORIZED)
	}
	apiKey, err := GenerateRandomAPIKey(32)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}

	_, err = tx.Exec("INSERT INTO api_key (permitted_project_id, key) VALUES ($1, $2) ON CONFLICT(permitted_project_id) DO UPDATE SET key=EXCLUDED.key", permittedId, apiKey)

	if err != nil {
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
	return apiKey, nil

}

func (db Db) CreateProject(user_id string, name string, description *string) (string, error) {
	tx, err := db.Db.Begin()
	if err != nil {
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	var project_id string
	row := tx.QueryRow("INSERT INTO project (user_id, name, description) VALUES ($1, $2, $3) RETURNING id", user_id, name, description)
	err = row.Scan(&project_id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			innerErr := tx.Rollback()
			if innerErr != nil {
				db.Logger.Println(innerErr)
				return "", errors.New(error_msgs.DATABASE_ERROR)
			}
			return "", errors.New(error_msgs.PROJECT_EXISTS)
		}
		db.Logger.Println(err)
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		fmt.Println("got to 4")
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	_, err = tx.Exec("INSERT INTO permitted_project (user_id, project_id) VALUES ($1, $2)", user_id, project_id)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	err = tx.Commit()
	if err != nil {
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	return project_id, nil
}
