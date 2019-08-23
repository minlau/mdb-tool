package main

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"sync"
)

type DatabaseStore struct {
	m         *sync.Mutex
	databases map[DatabaseGroup]DatabaseInstance
}

func NewDatabaseStore() *DatabaseStore {
	return &DatabaseStore{&sync.Mutex{}, make(map[DatabaseGroup]DatabaseInstance)}
}

func (s *DatabaseStore) AddDatabases(databases []DatabaseConfig) {
	var wg sync.WaitGroup
	for _, item := range databases {
		wg.Add(1)
		go func(config DatabaseConfig) {
			defer wg.Done()
			err := s.AddDatabase(config)
			if err != nil {
				log.Warn().Err(err).Msg("failed to add database")
			}
		}(item)
	}
	wg.Wait()
}

func (s *DatabaseStore) AddDatabase(config DatabaseConfig) error {
	s.m.Lock()
	if _, ok := s.databases[config.DatabaseGroup]; ok {
		s.m.Unlock()
		return errors.Errorf("database is already added with groupId=%v, groupType=%v", config.GroupId,
			config.GroupType)
	}
	s.databases[config.DatabaseGroup] = DatabaseInstance{}
	s.m.Unlock()

	db, err := config.DatabaseConnConfig.OpenDatabase()
	if err != nil {
		s.m.Lock()
		delete(s.databases, config.DatabaseGroup)
		s.m.Unlock()
		return errors.Wrap(err, "failed to open database")
	}
	s.m.Lock()
	s.databases[config.DatabaseGroup] = DatabaseInstance{config, db}
	s.m.Unlock()
	return nil
}

func (s *DatabaseStore) QueryDatabase(groupId int, groupType string, query string) GroupQueryResult {
	if databaseInstance, ok := s.databases[DatabaseGroup{groupId, groupType}]; ok {
		data, err := queryToMap(databaseInstance.DB, query)
		return GroupQueryResult{GroupId: groupId, Data: data, Error: NewQueryError(err)}
	} else {
		return GroupQueryResult{
			GroupId: groupId,
			Data:    nil,
			Error:   NewQueryError(errors.Errorf("no database registered with groupId: %d, groupType: %s", groupId, groupType)),
		}
	}
}

//does not have timeout, might be a problem
func (s *DatabaseStore) QueryMultipleDatabases(groupType string, query string) []GroupQueryResult {
	var results []GroupQueryResult
	var mutex = &sync.Mutex{}
	var filteredDatabases = make(map[int]*sqlx.DB)

	for key, value := range s.databases {
		if key.GroupType == groupType {
			filteredDatabases[key.GroupId] = value.DB
		}
	}

	var wg sync.WaitGroup
	for groupId, db := range filteredDatabases {
		wg.Add(1)
		go func(groupId int, db *sqlx.DB) {
			result, err := queryToMap(db, query)
			var groupQueryResult = GroupQueryResult{
				groupId,
				result,
				NewQueryError(err),
			}
			mutex.Lock()
			results = append(results, groupQueryResult)
			mutex.Unlock()
			wg.Done()
		}(groupId, db)
	}
	wg.Wait()
	return results
}

func (s *DatabaseStore) GetDatabaseItems() []DatabaseDescription {
	arr := make([]DatabaseDescription, 0, len(s.databases))
	for _, value := range s.databases {
		arr = append(arr, value.Config.DatabaseDescription)
	}
	return arr
}

type DatabaseInstance struct {
	Config DatabaseConfig
	DB     *sqlx.DB
}

func queryToMap(db *sqlx.DB, query string) (*QueryResult, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil && rollbackErr != sql.ErrTxDone {
				log.Error().Err(rollbackErr).Msg("failed to rollback transaction")
			}
		}
	}()

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	var result QueryResult
	for rows.Next() {
		if len(result.Columns) == 0 {
			columns, err := rows.Columns()
			if err != nil {
				return &result, err
			}

			for i, column := range columns {
				if contains(columns[:i], column) {
					for j := 0; j < i; j++ {
						newColumn := column + "__" + strconv.Itoa(j)
						if !contains(columns[:i], newColumn) {
							columns[i] = newColumn
						}
					}
				}
			}

			result.Columns = columns
		}

		row, err := customMapScan(rows, result.Columns)
		if err != nil {
			return &result, err
		}
		result.Rows = append(result.Rows, row)
	}
	err = tx.Commit()
	if err != nil {
		return &result, err
	}
	return &result, nil
}

func contains(arr []string, value string) bool {
	for _, arrValue := range arr {
		if arrValue == value {
			return true
		}
	}
	return false
}

func customMapScan(r sqlx.ColScanner, columns []string) (map[string]interface{}, error) {
	// ignore r.started, since we needn't use reflect for anything.
	valuesArr := make([]interface{}, len(columns))
	for i := range valuesArr {
		valuesArr[i] = new(interface{})
	}

	valuesMap := make(map[string]interface{})

	err := r.Scan(valuesArr...)
	if err != nil {
		return valuesMap, err
	}

	for i, column := range columns {
		switch (*(valuesArr[i].(*interface{}))).(type) {
		case []byte: //needed for mysql, some unsupported data types to convert byte array to string
			valuesMap[column] = string((*(valuesArr[i].(*interface{}))).([]byte))
		default:
			valuesMap[column] = *(valuesArr[i].(*interface{}))
		}
	}

	return valuesMap, r.Err()
}
