package store

import (
	"context"
	stdjson "encoding/json"
	"fmt"
	goJson "github.com/goccy/go-json"
	iterJson "github.com/json-iterator/go"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
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
	cfg, err := readConfig("testdata/bench_config.json")
	if err != nil {
		panic(fmt.Sprintf("failed to read cfg. %v", err))
	}

	databaseStore := NewDatabaseStore()
	databaseStore.AddDatabases(cfg.DatabaseConfigs)
	databaseConfigs, errs := GetDatabaseConfigsFromDataSources(cfg.DataSources)
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
			if marshaledJson, err := json.Marshal(data); err != nil || len(marshaledJson) == 0 {
				panic("marshal error: " + err.Error())
			}
		}
	})
	b.Run("json-iterator/go", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if marshaledJson, err := iterJson.Marshal(data); err != nil || len(marshaledJson) == 0 {
				panic("marshal error: " + err.Error())
			}
		}
	})
	b.Run("ccy/go-json", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if marshaledJson, err := goJson.Marshal(data); err != nil || len(marshaledJson) == 0 {
				panic("marshal error: " + err.Error())
			}
		}
	})
	b.Run("encoding/json", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if marshaledJson, err := stdjson.Marshal(data); err != nil || len(marshaledJson) == 0 {
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

func Test_getFieldNames(t *testing.T) {
	type args struct {
		columnNames []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{
				columnNames: nil,
			},
			want: []string{},
		},
		{
			name: "no duplicates",
			args: args{
				columnNames: []string{"id", "name"},
			},
			want: []string{"id", "name"},
		},
		{
			name: "duplicates",
			args: args{
				columnNames: []string{"id", "name", "id", "name"},
			},
			want: []string{"id", "name", "id__1", "name__1"},
		},
		{
			name: "same duplicates and conflicting name at the start",
			args: args{
				columnNames: []string{"id__1", "id", "id", "id"},
			},
			want: []string{"id__1", "id", "id__2", "id__3"},
		},
		{
			name: "same duplicates and conflicting name at the end",
			args: args{
				columnNames: []string{"id", "id", "id", "id__1"},
			},
			want: []string{"id", "id__1", "id__2", "id__1__1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, getFieldNames(tt.args.columnNames), "getFieldNames(%v)", tt.args.columnNames)
		})
	}
}
