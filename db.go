package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/pkg/errors"
)

type CompanyDatabase struct {
	CompanyID  int       `db:"company_id"`
	CreateDate time.Time `db:"create_date"`
	ModifyDate time.Time `db:"modify_date"`
	ID         int       `db:"id"`
	DbCfg
}

type DbInstance struct {
	Id  int
	Cfg DbCfg
	Db  *sqlx.DB
}

type DbCfg struct {
	Hostname string `db:"db_host"`
	Port     int    `db:"db_port"`
	Name     string `db:"db_name"`
	Username string `db:"db_username"`
	Password string `db:"db_password"`
	Type     int    `db:"db_type"`
}

type CompanyData struct {
	CompanyId int
	Data      []map[string]interface{}
}

var databases = make(map[int]DbInstance)
var maxId = 0

func initDatabasesFromDb() {
	coreDb, err := sqlx.Connect("postgres",
		"dbname=core user=postgres password=admin host=192.74.169.108 sslmode=disable")
	if err != nil {
		log.Error().Err(err).Msg("failed to connect to core db")
	}

	var databaseConfigs []CompanyDatabase
	err = coreDb.Select(&databaseConfigs, "select * from company_databases where db_type=0")
	if err != nil {
		log.Error().Err(err).Msg("failed to select company databases")
	}

	for _, element := range databaseConfigs {
		db := openDB(element.DbCfg)
		if db != nil {
			if _, ok := databases[element.CompanyID]; ok {
				log.Error().Msg(fmt.Sprintf("database is already registered under id: %d", element.CompanyID))
			} else {
				if maxId < element.CompanyID {
					maxId = element.CompanyID
				}
				databases[element.CompanyID] = DbInstance{
					element.CompanyID,
					element.DbCfg,
					db,
				}
			}
		}
	}
	log.Debug().Msg("Finished databases initialisation")
}

func initDatabasesFromJson() {
	var databaseConfigs []DbCfg

	configFile, err := os.Open("databases.json")
	if err != nil {
		log.Error().Err(err).Msg("opening file")
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&databaseConfigs); err != nil {
		log.Error().Err(err).Msg("parsing file")
	}

	for _, element := range databaseConfigs {
		db := openDB(element)
		if db != nil {
			maxId++
			databases[maxId] = DbInstance{
				maxId,
				element,
				db,
			}
		} else {
			log.Error().Err(err).Msg(fmt.Sprintf("failed to open db: %v", element))
		}
	}
	log.Debug().Msg("Finished databases initialisation")
}

func openDB(config DbCfg) *sqlx.DB {
	connectionUrl := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable host=%s port=%d",
		config.Username, config.Password, config.Name, config.Hostname, config.Port)
	db, err := sqlx.Connect("postgres", connectionUrl)
	if err != nil {
		log.Error().Err(err).Msg(fmt.Sprintf("failed to connect to '%s' db", config.Name))
		return nil
	}
	return db
}

func queryDatabase(id int, query string) ([]map[string]interface{}, error) {
	if db, ok := databases[id]; ok {
		return queryToMap(db.Db, query), nil
	} else {
		return nil, errors.New(fmt.Sprintf("no database registered with id: %d", id))
	}
}

func queryAllDatabases(query string) []CompanyData {
	var data []CompanyData
	var mutex = &sync.Mutex{}
	c := make(chan int, len(databases))
	for key, value := range databases {
		go func(key int, value *sqlx.DB) {
			res := queryToMap(value, query)
			var companyData = CompanyData{
				CompanyId: key,
				Data:      res,
			}
			mutex.Lock()
			data = append(data, companyData)
			mutex.Unlock()
			c <- key
		}(key, value.Db)
	}
	for range databases {
		<-c
	}
	return data
}

func queryToMap(db *sqlx.DB, query string) []map[string]interface{} {
	var data []map[string]interface{}
	rows, err := db.Queryx(query)
	if err != nil {
		log.Error().Err(err).Msgf("failed to execute query")

		result := make(map[string]interface{})
		result["error"] = err
		return append(data, result)
	}
	for rows.Next() {
		results := make(map[string]interface{})
		err = rows.MapScan(results)
		if err != nil {
			log.Error().Err(err).Msgf("failed to scan data")
			continue
		}

		data = append(data, results)
	}
	return data
}
