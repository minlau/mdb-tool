package main

import (
	"fmt"
	"github.com/segmentio/encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const benchQuery = "select * from messages limit 100000"
const benchGroupType = "main-db"

func initDatabaseStore() *DatabaseStore {
	config, err := readConfig("config.json")
	if err != nil {
		fmt.Errorf("failed to read config. %e", err)
		return nil
	}

	databaseStore := NewDatabaseStore()
	databaseStore.AddDatabases(config.DatabaseConfigs)
	databaseConfigs := GetDatabaseConfigsFromDataSources(config.DataSources)
	databaseStore.AddDatabases(databaseConfigs)
	return databaseStore
}

func BenchmarkQuery(b *testing.B) {
	databaseStore := initDatabaseStore()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		result := databaseStore.QueryMultipleDatabases(benchGroupType, benchQuery)
		if len(result) > -1 {
			continue
		}
	}
}

func BenchmarkQueryAndJSON(b *testing.B) {
	databaseStore := initDatabaseStore()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		json, _ := json.Marshal(databaseStore.QueryMultipleDatabases(benchGroupType, benchQuery))
		if json != nil {
			continue
		}
	}
}

func BenchmarkRequestQuery(b *testing.B) {
	u, err := url.Parse("localhost/query")
	if err != nil {
		fmt.Errorf("failed to parse url. %e", err)
	}
	q := u.Query()
	q.Set("groupType", benchGroupType)
	q.Set("query", benchQuery)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.RequestURI(), nil)
	if err != nil {
		fmt.Errorf("failed to create new request. %e", err)
	}

	databaseStore := initDatabaseStore()
	handler := query(databaseStore)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != 200 {
			fmt.Errorf("code is not 200. %d", rr.Code)
		}
	}
}
