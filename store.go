package main

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"sync"
	"time"
)

type DatabaseInstance struct {
	Config DatabaseConfig
	DB     *sqlx.DB
}

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
	db, err := OpenDatabase(config.DatabaseConnConfig)
	if err != nil {
		return errors.Wrap(err, "failed to open database")
	}
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(config.ConnMaxLifetimeInSeconds) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(config.ConnMaxIdleTimeInSeconds) * time.Second)

	s.m.Lock()
	defer s.m.Unlock()
	if _, ok := s.databases[config.DatabaseGroup]; ok {
		db.Close()
		return errors.Errorf("database is already added with groupId=%v, groupType=%v", config.GroupId,
			config.GroupType)
	}
	s.databases[config.DatabaseGroup] = DatabaseInstance{config, db}
	return nil
}

var selectPgTablesMetadata = `
SELECT
    t.table_name, c.column_name
FROM
    information_schema.tables AS t
    INNER JOIN information_schema.columns AS c ON t.table_name = c.table_name
WHERE
    t.table_schema = current_schema() AND t.table_type = 'BASE TABLE';
`
var selectFbTablesMetadata = `
SELECT
    trim(f.rdb$relation_name) AS table_name, trim(f.rdb$field_name) AS column_name
FROM
    rdb$relation_fields AS f
    JOIN rdb$relations AS r ON
            f.rdb$relation_name = r.rdb$relation_name
            AND r.rdb$view_blr IS NULL
            AND (r.rdb$system_flag IS NULL OR r.rdb$system_flag = 0)
ORDER BY
    1, f.rdb$field_position;
`
var selectMySqlTablesMetadata = `
SELECT
    table_name, column_name
FROM
    information_schema.columns
WHERE
    table_schema = database()
ORDER BY
    table_name, ordinal_position;
`

func getTablesMetadataSql(sqlType string) string {
	switch sqlType {
	case "postgresql":
		return selectPgTablesMetadata
	case "firebird":
		return selectFbTablesMetadata
	case "mysql":
		return selectMySqlTablesMetadata
	default:
		return ""
	}
}

func queryTablesMetadata(db *sqlx.DB, query string) (map[string][]string, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	data := make(map[string][]string)
	for rows.Next() {
		var tableName string
		var columnName string
		err := rows.Scan(&tableName, &columnName)
		if err != nil {
			return nil, err
		}

		if val, ok := data[tableName]; ok {
			data[tableName] = append(val, columnName)
		} else {
			data[tableName] = []string{columnName}
		}
	}
	return data, nil
}

func (s *DatabaseStore) GetTablesMetadata(groupId int, groupType string) (map[string][]string, error) {
	if databaseInstance, ok := s.databases[DatabaseGroup{groupId, groupType}]; ok {
		sqlType := databaseInstance.Config.Type
		query := getTablesMetadataSql(sqlType)
		data, err := queryTablesMetadata(databaseInstance.DB, query)
		return data, err
	} else {
		return nil, errors.Errorf("no database registered with groupId: %d, groupType: %s", groupId, groupType)
	}
}

func (s *DatabaseStore) QueryDatabase(groupId int, groupType string, query string) GroupQueryResult {
	if databaseInstance, ok := s.databases[DatabaseGroup{groupId, groupType}]; ok {
		data, err := executeQuery(databaseInstance.DB, query)
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
			defer wg.Done()

			data, err := executeQuery(db, query)
			var groupQueryResult = GroupQueryResult{
				groupId,
				data,
				NewQueryError(err),
			}
			mutex.Lock()
			defer mutex.Unlock()
			results = append(results, groupQueryResult)
		}(groupId, db)
	}
	wg.Wait()
	return results
}

type DatabaseItem struct {
	DatabaseDescription
	Type string `json:"type"`
}

func (s *DatabaseStore) GetDatabaseItems() []DatabaseItem {
	arr := make([]DatabaseItem, 0, len(s.databases))
	for _, value := range s.databases {
		arr = append(arr, DatabaseItem{
			DatabaseDescription: value.Config.DatabaseDescription,
			Type:                value.Config.Type,
		})
	}
	return arr
}

func executeQuery(db *sqlx.DB, query string) (*QueryData, error) {
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
	var data QueryData
	fieldNames := make([]string, 0)
	for rows.Next() {
		if len(data.Columns) == 0 {
			columnNames, err := rows.Columns()
			if err != nil {
				return &data, err
			}

			fieldNames = make([]string, len(columnNames))
			copy(fieldNames, columnNames)
			for i, fieldName := range fieldNames {
				if contains(fieldNames[:i], fieldName) {
					for j := 0; j < i; j++ {
						newFieldName := fieldName + "__" + strconv.Itoa(j+1)
						if !contains(fieldNames[:i], newFieldName) {
							fieldNames[i] = newFieldName
							break
						}
					}
				}
			}

			columns := make([]Column, 0, len(columnNames))
			for i := range columnNames {
				columns = append(columns, Column{
					Name:      columnNames[i],
					FieldName: fieldNames[i],
				})
			}

			data.Columns = columns
		}

		row, err := customMapScan(rows, fieldNames)
		if err != nil {
			return &data, err
		}
		data.Rows = append(data.Rows, row)
	}
	err = tx.Commit()
	if err != nil {
		return &data, err
	}
	return &data, nil
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
