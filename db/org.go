package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type Org struct {
	Id          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	LocationId  *string   `json:"location_id"`
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
		db.Logger.Println(err)
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", respErr
	}
	query = "INSERT INTO user_org (user_id, org_id, level) VALUES ($1, $2, 500)"
	_, err = tx.Exec(query, userId, orgId)
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
	return orgId, nil
}

func (db Db) GetOrgs(userId string, orgId *string, name *string, from *time.Time, to *time.Time) ([]Org, error) {
	var orgs []Org
	tx, err := db.Db.Begin()
	if err != nil {
		db.Logger.Println(err)
		return nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	var rows *sql.Rows
	query := "SELECT org.id, org.created_at, org.name, org.description, org.location_id FROM org INNER JOIN user_org WHERE user_org.user_id = $1"
	variableIndex := 2
	args := make([]any, 0)
	args = append(args, userId)
	// TODO: I kind of hate this
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
		var org Org
		err = rows.Scan(&org.Id, &org.CreatedAt, &org.Name, &org.Description, &org.LocationId)
		if err != nil {
			db.Logger.Println(err)
			innerErr := tx.Rollback()
			if innerErr != nil {
				db.Logger.Println(innerErr)
				return nil, errors.New(error_msgs.DATABASE_ERROR)
			}
			return nil, errors.New(error_msgs.DATABASE_ERROR)
		}
		orgs = append(orgs, org)
	}
	return orgs, nil
}
