package main

import (
	"github.com/minlau/mdb-tool/store"
	"reflect"
	"testing"
)

func Test_LoadConfig(t *testing.T) {
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
			args{path: "testdata/test_read_config.json"},
			&Config{
				DataSources: []store.DataSource{{
					Query: "select 1",
					DatabaseConnConfig: store.DatabaseConnConfig{
						Hostname: "localhost",
						Port:     5432,
						Name:     "test-non-existing-1",
						Username: "postgres",
						Password: "admin",
						Type:     "postgresql",
					},
				}},
				DatabaseConfigs: []store.DatabaseConfig{
					{
						DatabaseGroup: store.DatabaseGroup{
							GroupName: "a",
							GroupType: "test-db",
						},
						DatabaseConnConfig: store.DatabaseConnConfig{
							Hostname: "localhost",
							Port:     5432,
							Name:     "test-non-existing-1",
							Username: "postgres",
							Password: "admin",
							Type:     "postgresql",
						},
						DatabaseConnPoolConfig: store.DatabaseConnPoolConfig{
							MaxOpenConns:             2,
							MaxIdleConns:             1,
							ConnMaxLifetimeInSeconds: 600,
							ConnMaxIdleTimeInSeconds: 60,
						},
					},
					{
						DatabaseGroup: store.DatabaseGroup{
							GroupName: "b",
							GroupType: "test-db",
						},
						DatabaseConnConfig: store.DatabaseConnConfig{
							Hostname: "localhost",
							Port:     5432,
							Name:     "test-non-existing-2",
							Username: "postgres",
							Password: "admin",
							Type:     "postgresql",
						},
						DatabaseConnPoolConfig: store.DatabaseConnPoolConfig{
							MaxOpenConns:             4,
							MaxIdleConns:             2,
							ConnMaxLifetimeInSeconds: 300,
							ConnMaxIdleTimeInSeconds: 30,
						},
					},
				},
			},
			false,
		},
		{
			"non existing file",
			args{path: "testdata/config_non_existing.json"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
