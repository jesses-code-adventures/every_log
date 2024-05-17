package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jesses-code-adventures/every_log/error_msgs"
)

type ProjectInvite struct {
	Id         string `json:"id"`
	FromUserId string `json:"from_user_id"`
	ToUserId   string `json:"to_user_id"`
	Status     string `json:"status"`
	ProjectId  string `json:"project_id"`
}

type OrgInvite struct {
	Id         string `json:"id"`
	FromUserId string `json:"from_user_id"`
	ToUserId   string `json:"to_user_id"`
	Status     string `json:"status"`
	OrgId      string `json:"org_id"`
}

func (db Db) getUserOrgIdFromUserOrg(userId string, orgId string) (string, error) {
	var permittedProjectId string
	err := db.Db.QueryRow(`SELECT id
FROM user_org
WHERE user_id = $1
AND org_id = $2`, userId, orgId).Scan(&permittedProjectId)
	if err != nil {
		db.Logger.Println("User org db error")
		db.Logger.Println(err)
		return "", errors.New(error_msgs.UNAUTHORIZED)
	}
	return permittedProjectId, nil
}

func (db Db) CreateOrgInvite(fromUserId string, toUserId string, orgId string) (string, error) {
	var inviteId string
	tx, err := db.Db.Begin()
	if err != nil {
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	_, err = db.getUserOrgIdFromUserOrg(fromUserId, orgId)
	if err != nil {
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", err
	}
	row := tx.QueryRow("INSERT INTO org_invite (from_user_id, to_user_id, org_id) VALUES ($1, $2, $3) RETURNING id", fromUserId, toUserId, orgId)
	err = row.Scan(&inviteId)
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
	return inviteId, nil
}

func (db Db) CreateProjectInvite(fromUserId string, toUserId string, projectId string, apiKey string) (string, error) {
	var inviteId string
	tx, err := db.Db.Begin()
	if err != nil {
		db.Logger.Println(err)
		return "", errors.New(error_msgs.DATABASE_ERROR)
	}
	_, err = db.getPermittedProjectIdFromApiKey(fromUserId, apiKey)
	if err != nil {
		// TODO: Add unauthorized vs other error handling
		innerErr := tx.Rollback()
		if innerErr != nil {
			db.Logger.Println(innerErr)
			return "", errors.New(error_msgs.DATABASE_ERROR)
		}
		return "", err
	}
	row := tx.QueryRow("INSERT INTO project_invite (from_user_id, to_user_id, project_id) VALUES ($1, $2, $3) RETURNING id", fromUserId, toUserId, projectId)
	err = row.Scan(&inviteId)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			innerErr := tx.Rollback()
			if innerErr != nil {
				db.Logger.Println(innerErr)
				return "", errors.New(error_msgs.DATABASE_ERROR)
			}
			return "", errors.New(error_msgs.GetExistsMessage(fmt.Sprintf("invite from %s to %s for project %s", fromUserId, toUserId, projectId)))
		}
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
	return inviteId, nil
}

func (db Db) GetProjectInvites(requestingUserId string, fromUserId *string, toUserId *string, projectId *string, status *string, from *time.Time, to *time.Time) ([]ProjectInvite, error) {
	var project_invites []ProjectInvite
	tx, err := db.Db.Begin()
	if err != nil {
		db.Logger.Println(err)
		return nil, errors.New(error_msgs.DATABASE_ERROR)
	}
	var rows *sql.Rows
	query := "SELECT * FROM log WHERE user_id = $1"
	variableIndex := 2
	args := make([]any, 0)
	args = append(args, fromUserId)
	// TODO: I kind of hate this
	if projectId != nil {
		query += fmt.Sprintf(" AND project_id = $%d", variableIndex)
		args = append(args, *projectId)
		variableIndex++
	}
	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", variableIndex)
		args = append(args, *status)
		variableIndex++
	}
	if fromUserId != nil {
		query += fmt.Sprintf(" AND from_user_id = $%d", variableIndex)
		args = append(args, *fromUserId)
		variableIndex++
	}
	if toUserId != nil {
		query += fmt.Sprintf(" AND to_user_id = $%d", variableIndex)
		args = append(args, *toUserId)
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
		var invite ProjectInvite
		err = rows.Scan(&invite)
		if err != nil {
			db.Logger.Println(err)
			innerErr := tx.Rollback()
			if innerErr != nil {
				db.Logger.Println(innerErr)
				return nil, errors.New(error_msgs.DATABASE_ERROR)
			}
			return nil, errors.New(error_msgs.DATABASE_ERROR)
		}
		project_invites = append(project_invites, invite)
	}
	return project_invites, nil
}
