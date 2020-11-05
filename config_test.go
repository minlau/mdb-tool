package main

import (
	"reflect"
	"testing"
)

func Test_readConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			"valid file",
			args{path: "test/config.json"},
			&Config{DataSources: []DataSource{{Query: "select 1", DatabaseConnConfig: DatabaseConnConfig{Hostname: "localhost", Port: 5432, Name: "test-non-existing-1", Username: "postgres", Password: "admin", Type: "postgresql"}}}, DatabaseConfigs: []DatabaseConfig{{DatabaseDescription: DatabaseDescription{Title: "test 1", DatabaseGroup: DatabaseGroup{GroupId: 1, GroupType: "test-db"}}, DatabaseConnConfig: DatabaseConnConfig{Hostname: "localhost", Port: 5432, Name: "test-non-existing-1", Username: "postgres", Password: "admin", Type: "postgresql"}}, {DatabaseDescription: DatabaseDescription{Title: "test 2", DatabaseGroup: DatabaseGroup{GroupId: 2, GroupType: "test-db"}}, DatabaseConnConfig: DatabaseConnConfig{Hostname: "localhost", Port: 5432, Name: "test-non-existing-2", Username: "postgres", Password: "admin", Type: "postgresql"}}}},
			false,
		},
		{
			"non existing file",
			args{path: "test/config_non_existing.json"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
