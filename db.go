package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/nakagami/firebirdsql"
	"github.com/pkg/errors"
	"os"
)

type GroupQueryResult struct {
	GroupId int                      `json:"groupId"`
	Data    []map[string]interface{} `json:"data"`
	Error   error                    `json:"error"`
}

type DatabaseInstance struct {
	Config DatabaseConfig
	DB     *sqlx.DB
}

type DatabaseConfig struct {
	DatabaseDescription
	DatabaseConnConfig
}

type DatabaseDescription struct {
	Title string `json:"title"`
	DatabaseGroup
}

type DatabaseGroup struct {
	GroupId   int    `json:"groupId"`
	GroupType string `json:"groupType"`
}

type DatabaseConnConfig struct {
	Hostname string `db:"hostname"`
	Port     int    `db:"port"`
	Name     string `db:"name"`
	Username string `db:"username"`
	Password string `db:"password"`
	Type     string `db:"type"`
}

func (c DatabaseConnConfig) getConnectionUrl() (string, error) {
	switch c.Type {
	case "postgresql":
		return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
			c.Username, c.Password, c.Hostname, c.Port, c.Name), nil
	case "firebird":
		return fmt.Sprintf("%s:%s@%s:%d/%s",
			c.Username, c.Password, c.Hostname, c.Port, c.Name), nil
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
			c.Username, c.Password, c.Hostname, c.Port, c.Name), nil
	default:
		return "", errors.Errorf("unknown type: %s", c.Type)
	}
}

func (c DatabaseConnConfig) getDriverName() (string, error) {
	switch c.Type {
	case "postgresql":
		return "postgres", nil
	case "firebird":
		return "firebirdsql", nil
	case "mysql":
		return "mysql", nil
	default:
		return "", errors.Errorf("unknown type: %s", c.Type)
	}
}

const DbSelect = "select company_id as groupId || '', db_host as hostname, db_port as port, db_name as name, " +
	"db_username as username, db_password as password, " +
	"case when (db_type == 0) then 'postgresql' else 'firebird' end as groupType " +
	"from company_databases"

func getDatabaseConfigsFromDb() ([]DatabaseConfig, error) {
	db, err := sqlx.Connect("postgres",
		"dbname=core user=postgres password=admin host=192.74.169.108 sslmode=disable")
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to db")
	}
	defer db.Close()

	var databaseConfigs []DatabaseConfig
	err = db.Select(&databaseConfigs, DbSelect)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select database configs")
	}
	return databaseConfigs, nil
}

func getDatabaseConfigsFromFile(path string) ([]DatabaseConfig, error) {
	var databaseConfigs []DatabaseConfig

	configFile, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open file")
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&databaseConfigs); err != nil {
		return nil, errors.Wrap(err, "failed to parse file")
	}
	return databaseConfigs, nil
}
