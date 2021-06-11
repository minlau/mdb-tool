package main

import (
	"context"
	stdjson "encoding/json"
	"fmt"
	iterJson "github.com/json-iterator/go"
	"github.com/segmentio/encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const benchPrepareSchema = `
CREATE TABLE messages (
    id integer NOT NULL,
    text character varying(10000) NOT NULL,
    is_read boolean NOT NULL DEFAULT false,
    sender_id integer NOT NULL,
    receiver_id integer NOT NULL,
    create_date timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT messages_pk PRIMARY KEY (id)
)
`
const benchGenerateData = `
INSERT INTO messages(id, text, is_read, sender_id, receiver_id, create_date) 
SELECT gen, 'gen'||gen, false, 0, 0, now() from generate_series(1,100000) gen
`
const benchClearSchema = `
drop TABLE messages;
`
const benchQuery = `select * from messages limit 100000`
const benchGroupType = "test"

func initDatabaseStore() *DatabaseStore {
	config, err := readConfig("testdata/bench_config.json")
	if err != nil {
		panic(fmt.Sprintf("failed to read config. %v", err))
		return nil
	}

	databaseStore := NewDatabaseStore()
	databaseStore.AddDatabases(config.DatabaseConfigs)
	databaseConfigs, errs := GetDatabaseConfigsFromDataSources(config.DataSources)
	if len(errs) > 0 {
		panic(fmt.Sprintf("GetDatabaseConfigsFromDataSources contains errors. Errors: %v", err))
	}
	databaseStore.AddDatabases(databaseConfigs)
	return databaseStore
}

func initData(dataStore *DatabaseStore, groupType string) {
	dataStore.QueryMultipleDatabases(context.Background(), groupType, benchPrepareSchema)
	dataStore.QueryMultipleDatabases(context.Background(), groupType, benchGenerateData)
}

func clearData(dataStore *DatabaseStore, groupType string) {
	dataStore.QueryMultipleDatabases(context.Background(), groupType, benchClearSchema)
}

func BenchmarkEncodeJson(b *testing.B) {
	databaseStore := initDatabaseStore()
	initData(databaseStore, benchGroupType)
	defer clearData(databaseStore, benchGroupType)

	data := databaseStore.QueryMultipleDatabases(context.Background(), benchGroupType, benchQuery)
	b.ResetTimer()
	b.Run("segmentio/encoding/json", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if json, err := json.Marshal(data); err != nil || len(json) == 0 {
				panic("marshal error: " + err.Error())
			}
		}
	})
	b.Run("json-iterator/go", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if json, err := iterJson.Marshal(data); err != nil || len(json) == 0 {
				panic("marshal error: " + err.Error())
			}
		}
	})
	b.Run("encoding/json", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if json, err := stdjson.Marshal(data); err != nil || len(json) == 0 {
				panic("marshal error: " + err.Error())
			}
		}
	})
}

func BenchmarkQuery(b *testing.B) {
	databaseStore := initDatabaseStore()
	initData(databaseStore, benchGroupType)
	defer clearData(databaseStore, benchGroupType)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		result := databaseStore.QueryMultipleDatabases(context.Background(), benchGroupType, benchQuery)
		if len(result) > -1 {
			continue
		}
	}
}

func BenchmarkRequest(b *testing.B) {
	u, err := url.Parse("localhost/query")
	if err != nil {
		b.Fatalf("failed to parse url. %e", err)
	}
	q := u.Query()
	q.Set("groupType", benchGroupType)
	q.Set("query", benchQuery)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.RequestURI(), nil)
	if err != nil {
		b.Fatalf("failed to create new request. %e", err)
	}

	databaseStore := initDatabaseStore()
	initData(databaseStore, benchGroupType)
	defer clearData(databaseStore, benchGroupType)

	handler := query(databaseStore)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != 200 {
			b.Errorf("code is not 200. %d", rr.Code)
		}
	}
}
