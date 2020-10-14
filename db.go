package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/nakagami/firebirdsql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type DataSource struct {
	DatabaseConnConfig
	Query string
}

type DatabaseConfig struct {
	DatabaseDescription
	DatabaseConnConfig
}

type DatabaseDescription struct {
	DatabaseGroup
	Title string `json:"title"`
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

func getConnectionUrl(c DatabaseConnConfig) (string, error) {
	switch c.Type {
	case "postgresql":
		return fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s sslmode=disable",
			c.Username, c.Password, c.Hostname, c.Port, c.Name), nil
	case "firebird":
		return fmt.Sprintf("%s:%s@%s:%d/%s",
			c.Username, c.Password, c.Hostname, c.Port, c.Name), nil
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true",
			c.Username, c.Password, c.Hostname, c.Port, c.Name), nil
	default:
		return "", errors.Errorf("unknown type: %s", c.Type)
	}
}

func getDriverName(c DatabaseConnConfig) (string, error) {
	switch c.Type {
	case "postgresql":
		return "pgx", nil
	case "firebird":
		return "firebirdsql", nil
	case "mysql":
		return "mysql", nil
	default:
		return "", errors.Errorf("unknown type: %s", c.Type)
	}
}

func OpenDatabase(c DatabaseConnConfig) (*sqlx.DB, error) {
	driverName, err := getDriverName(c)
	if err != nil {
		return nil, err
	}

	connectionUrl, err := getConnectionUrl(c)
	if err != nil {
		return nil, err
	}

	var db *sqlx.DB
	if driverName == "pgx" {
		connConfig, err := pgx.ParseConfig(connectionUrl)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse pgx config. config=%#v", c)
		}
		//disable implicit prepared statement to enable execution of multiple queries at once
		connConfig.PreferSimpleProtocol = true

		openDB := stdlib.OpenDB(*connConfig)
		db = sqlx.NewDb(openDB, driverName)
		err = db.Ping()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to connect to database. config=%#v", c)
		}
	} else {
		db, err = sqlx.Connect(driverName, connectionUrl)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to connect to database. config=%#v", c)
		}
	}
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

func GetDatabaseConfigsFromDataSources(dataSources []DataSource) []DatabaseConfig {
	var databaseConfigs []DatabaseConfig
	var mutex sync.Mutex
	var wg sync.WaitGroup

	for _, item := range dataSources {
		wg.Add(1)
		go func(dataSource DataSource) {
			defer wg.Done()

			configs, err := GetDatabaseConfigsFromDataSource(dataSource)
			if err != nil {
				log.Warn().Err(err).Msg("failed to get database configs from db")
				return
			}

			mutex.Lock()
			defer mutex.Unlock()
			databaseConfigs = append(databaseConfigs, configs...)
		}(item)
	}
	wg.Wait()
	return databaseConfigs
}

func GetDatabaseConfigsFromDataSource(dataSource DataSource) ([]DatabaseConfig, error) {
	db, err := OpenDatabase(dataSource.DatabaseConnConfig)
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
