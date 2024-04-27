package db

import (
	"encoding/json"
	"fmt"
)

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
