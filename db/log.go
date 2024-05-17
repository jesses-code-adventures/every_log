package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jesses-code-adventures/every_log/error_msgs"
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

func (db Db) CreateLog(userId string, project_id string, level_id int, process_id *string, message string, traceback *string, apiKey string) (string, error) {
	var logId string
	tx, err := db.Db.Begin()
	if err != nil {
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	_, err = db.getPermittedProjectIdFromApiKey(userId, apiKey)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", err
	}
	row := tx.QueryRow("INSERT INTO log (user_id, project_id, level_id, process_id, message, traceback) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id", userId, project_id, level_id, process_id, message, traceback)
	err = row.Scan(&logId)
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
	return logId, nil
}

func (db Db) GetLogs(userId string, projectId *string, levelId *int, processId *string, orgId *string, from *time.Time, to *time.Time) ([]Log, error) {
	var logs []Log
	tx, err := db.Db.Begin()
	if err != nil {
		db.Logger.Println(err)
		return nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	var rows *sql.Rows
	query := "SELECT * FROM log WHERE user_id = $1"
	variableIndex := 2
	args := make([]any, 0)
	args = append(args, userId)
	// TODO: I kind of hate this
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
		db.Logger.Println(err)
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return nil, errors.New(error_msgs.DATABASE_ERROR)
		}
		return nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	defer rows.Close()
	for rows.Next() {
		var log Log
		err = rows.Scan(&log.Id, &log.CreatedAt, &log.UserId, &log.ProjectId, &log.LevelId, &log.ProcessId, &log.Message, &log.Traceback)
		if err != nil {
			db.Logger.Println(err)
			innerErr := tx.Rollback()
			if innerErr != nil {
				db.Logger.Println(innerErr)
				return nil, errors.New(error_msgs.DATABASE_ERROR)
			}
			return nil, errors.New(error_msgs.DATABASE_ERROR)
		}
		logs = append(logs, log)
	}
	return logs, nil
}
