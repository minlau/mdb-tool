package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/nakagami/firebirdsql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"sync"
)

type DataSource struct {
	Query string
	DatabaseConnConfig
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

func (c DatabaseConnConfig) OpenDatabase() (*sqlx.DB, error) {
	driverName, err := c.getDriverName()
	if err != nil {
		return nil, err
	}

	connectionUrl, err := c.getConnectionUrl()
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Connect(driverName, connectionUrl)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to connect to database. config=%#v", c)
	}
	return db, nil
}

func GetDatabaseConfigsFromDataSources(dataSources []DataSource) []DatabaseConfig {
	var databaseConfigs []DatabaseConfig
	var mutex sync.Mutex
	var wg sync.WaitGroup

	for _, item := range dataSources {
		wg.Add(1)
		go func(dataSource DataSource) {
			configs, err := GetDatabaseConfigsFromDataSource(dataSource)
			if err != nil {
				log.Warn().Err(err).Msg("failed to get database configs from db")
			} else {
				mutex.Lock()
				databaseConfigs = append(databaseConfigs, configs...)
				mutex.Unlock()
			}
			wg.Done()
		}(item)
	}
	wg.Wait()
	return databaseConfigs
}

func GetDatabaseConfigsFromDataSource(dataSource DataSource) ([]DatabaseConfig, error) {
	db, err := dataSource.DatabaseConnConfig.OpenDatabase()
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}
	defer db.Close()

	var databaseConfigs []DatabaseConfig
	err = db.Select(&databaseConfigs, dataSource.Query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select database configs. config=%#v",
			dataSource.DatabaseConnConfig)
	}
	return databaseConfigs, nil
}
