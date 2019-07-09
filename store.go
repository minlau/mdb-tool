package main

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
			Data:    QueryResult{},
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

func queryToMap(db *sqlx.DB, query string) (QueryResult, error) {
	tx, err := db.Begin()
	if err != nil {
		return QueryResult{}, err
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
		return QueryResult{}, err
	}
	var data QueryResult
	for rows.Next() {
		if len(data.Columns) == 0 {
			columns, err := rows.Columns()
			if err != nil {
				return data, err
			}
			data.Columns = columns
		}

		row, err := customMapScan(rows, len(data.Columns))
		if err != nil {
			return QueryResult{}, err
		}
		data.Rows = append(data.Rows, row)
	}
	err = tx.Commit()
	if err != nil {
		return data, err
	}
	return data, nil
}

//copy of sqlx.go func MapScan(r ColScanner, dest map[string]interface{}) error {}
func customMapScan(r sqlx.ColScanner, columnsLen int) ([]interface{}, error) {
	// ignore r.started, since we needn't use reflect for anything.
	values := make([]interface{}, columnsLen)
	for i := range values {
		values[i] = new(interface{})
	}

	err := r.Scan(values...)
	if err != nil {
		return values, err
	}

	for i := 0; i < columnsLen; i++ {
		switch (*(values[i].(*interface{}))).(type) {
		case []byte: //needed for mysql, some unsupported data types to convert byte array to string
			values[i] = string((*(values[i].(*interface{}))).([]byte))
		default:
			values[i] = *(values[i].(*interface{}))
		}
	}

	return values, r.Err()
}
