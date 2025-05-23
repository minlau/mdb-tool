package store

import (
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	_ "github.com/nakagami/firebirdsql"
	"github.com/pkg/errors"

	"github.com/minlau/mdb-tool/internal/utils/closer"
)

type DataSource struct {
	DatabaseConnConfig
	Query string
}

type DatabaseConfig struct {
	DatabaseGroup
	DatabaseConnConfig
	DatabaseConnPoolConfig
}

type DatabaseGroup struct {
	GroupName string `db:"groupName" json:"groupName"`
	GroupType string `db:"groupType" json:"groupType"`
}

type DatabaseConnConfig struct {
	Hostname string `db:"hostname"`
	Port     int    `db:"port"`
	Name     string `db:"name"`
	Username string `db:"username"`
	Password string `db:"password"`
	Type     string `db:"type"`
}

type DatabaseConnPoolConfig struct {
	MaxOpenConns             int `db:"maxOpenConns"`
	MaxIdleConns             int `db:"maxIdleConns"`
	ConnMaxLifetimeInSeconds int `db:"connMaxLifetimeInSeconds"`
	ConnMaxIdleTimeInSeconds int `db:"connMaxIdleTimeInSeconds"`
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
		// disable implicit prepared statement to enable execution of multiple queries at once
		connConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

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
	return db, nil
}

func GetDatabaseConfigsFromDataSources(dataSources []DataSource) ([]DatabaseConfig, []error) {
	databaseConfigs := make([]DatabaseConfig, 0)
	errs := make([]error, 0)
	var mutex sync.Mutex
	var wg sync.WaitGroup

	for _, item := range dataSources {
		wg.Add(1)
		go func(dataSource DataSource) {
			defer wg.Done()

			configs, err := GetDatabaseConfigsFromDataSource(dataSource)
			if err != nil {
				mutex.Lock()
				defer mutex.Unlock()
				errs = append(errs, err)
				return
			}

			mutex.Lock()
			defer mutex.Unlock()
			databaseConfigs = append(databaseConfigs, configs...)
		}(item)
	}
	wg.Wait()
	return databaseConfigs, errs
}

func GetDatabaseConfigsFromDataSource(dataSource DataSource) ([]DatabaseConfig, error) {
	db, err := OpenDatabase(dataSource.DatabaseConnConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open database. databaseConnConfig=%#v",
			dataSource.DatabaseConnConfig)
	}
	defer closer.Handle(db, "database")

	var databaseConfigs []DatabaseConfig
	err = db.Select(&databaseConfigs, dataSource.Query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select database configs. dataSource=%#v",
			dataSource)
	}
	return databaseConfigs, nil
}
