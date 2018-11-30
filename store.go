package main

import (
	"bytes"
	"encoding/json"
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
		return GroupQueryResult{GroupId: groupId, Data: data, Error: err}
	} else {
		return GroupQueryResult{
			GroupId: groupId,
			Data:    nil,
			Error:   errors.Errorf("no database registered with groupId=%d, groupType=%s", groupId, groupType),
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
				err,
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

type OrderedMap struct {
	Order []string
	Map   map[string]interface{}
}

func (om OrderedMap) MarshalJSON() ([]byte, error) {
	var b []byte
	buf := bytes.NewBuffer(b)
	buf.WriteRune('{')
	l := len(om.Order)
	for i, key := range om.Order {
		km, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}
		buf.Write(km)
		buf.WriteRune(':')
		vm, err := json.Marshal(om.Map[key])
		if err != nil {
			return nil, err
		}
		buf.Write(vm)
		if i != l-1 {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}

type DatabaseInstance struct {
	Config DatabaseConfig
	DB     *sqlx.DB
}

type GroupQueryResult struct {
	GroupId int          `json:"groupId"`
	Data    []OrderedMap `json:"data"`
	Error   error        `json:"error"`
}

func queryToMap(db *sqlx.DB, query string) ([]OrderedMap, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				log.Error().Err(rollbackErr).Msg("failed to rollback transaction")
			}
		}
	}()

	rows, err := tx.Query(query)
	if err != nil {
		return nil, err
	}
	var data []OrderedMap
	for rows.Next() {
		results := OrderedMap{Map: make(map[string]interface{})}
		err = customMapScan(rows, &results)
		if err != nil {
			return nil, err
		}
		data = append(data, results)
	}
	err = tx.Commit()
	if err != nil {
		return data, err
	}
	return data, nil
}

//copy of sqlx.go func MapScan(r ColScanner, dest map[string]interface{}) error {}
func customMapScan(r sqlx.ColScanner, dest *OrderedMap) error {
	// ignore r.started, since we needn't use reflect for anything.
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	dest.Order = columns

	values := make([]interface{}, len(columns))
	for i := range values {
		values[i] = new(interface{})
	}

	err = r.Scan(values...)
	if err != nil {
		return err
	}

	for i, column := range columns {
		if _, ok := dest.Map[column]; ok {
			column = column + "__" + strconv.Itoa(i)
			columns[i] = column
		}
		dest.Map[column] = *(values[i].(*interface{}))
	}

	return r.Err()
}
