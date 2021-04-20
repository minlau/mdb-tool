package main

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestGetConnectionUrl(t *testing.T) {
	tests := []struct {
		config  DatabaseConnConfig
		want    string
		wantErr error
	}{
		{
			DatabaseConnConfig{
				Hostname: "hostname-t",
				Name:     "name-t",
				Port:     1234,
				Type:     "postgresql",
				Username: "username-t",
				Password: "password-t",
			},
			"user=username-t password=password-t host=hostname-t port=1234 dbname=name-t sslmode=disable",
			nil,
		},
		{
			DatabaseConnConfig{
				Hostname: "hostname-t",
				Name:     "name-t",
				Port:     1234,
				Type:     "mysql",
				Username: "username-t",
				Password: "password-t",
			},
			"username-t:password-t@tcp(hostname-t:1234)/name-t?multiStatements=true",
			nil,
		},
		{
			DatabaseConnConfig{
				Hostname: "hostname-t",
				Name:     "name-t",
				Port:     1234,
				Type:     "firebird",
				Username: "username-t",
				Password: "password-t",
			},
			"username-t:password-t@hostname-t:1234/name-t",
			nil,
		},
		{
			DatabaseConnConfig{
				Hostname: "hostname-t",
				Name:     "name-t",
				Port:     1234,
				Type:     "",
				Username: "username-t",
				Password: "password-t",
			},
			"",
			errors.Errorf("unknown type: %s", ""),
		},
	}

	for _, tt := range tests {
		url, err := getConnectionUrl(tt.config)

		assert.Equal(t, tt.want, url)
		assert.IsType(t, tt.wantErr, err)
	}
}

func TestGetDriverName(t *testing.T) {
	tests := []struct {
		config  DatabaseConnConfig
		want    string
		wantErr error
	}{
		{
			DatabaseConnConfig{
				Type: "postgresql",
			},
			"pgx",
			nil,
		},
		{
			DatabaseConnConfig{
				Type: "mysql",
			},
			"mysql",
			nil,
		},
		{
			DatabaseConnConfig{
				Type: "firebird",
			},
			"firebirdsql",
			nil,
		},
		{
			DatabaseConnConfig{
				Type: "",
			},
			"",
			errors.Errorf("unknown type: %s", ""),
		},
	}

	for _, tt := range tests {
		url, err := getDriverName(tt.config)

		assert.Equal(t, tt.want, url)
		assert.IsType(t, tt.wantErr, err)
	}
}

func TestOpenDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	config, err := readConfig("testdata/integration_config.json")
	if err != nil {
		t.Fatalf("failed to parse config file. Error: %v", err)
	}

	tests := []struct {
		name    string
		cfg     DatabaseConnConfig
		want    bool
		wantErr bool
	}{
		{
			"connects to db",
			config.DatabaseConfigs[0].DatabaseConnConfig,
			true,
			false,
		},
		{
			"fails to connect to db",
			DatabaseConnConfig{
				Hostname: "localhost",
				Name:     "test-does-not-exist",
				Port:     5432,
				Type:     "postgresql",
				Username: "postgres",
				Password: "admin",
			},
			false,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OpenDatabase(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenDatabase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got != nil) != tt.want {
				t.Errorf("OpenDatabase() got = %v, want not nil", got)
			}
		})
	}
}

func TestGetDatabaseConfigsFromDataSources(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	config, err := readConfig("testdata/integration_config.json")
	if err != nil {
		t.Fatalf("failed to parse config file. Error: %v", err)
	}

	type args struct {
		dataSources []DataSource
	}
	tests := []struct {
		name         string
		args         args
		want         []DatabaseConfig
		wantErrsSize int
	}{
		{
			name: "datasource select with valid query",
			args: args{
				dataSources: []DataSource{
					{
						DatabaseConnConfig: config.DatabaseConfigs[0].DatabaseConnConfig,
						Query: `
select 1 as "groupId", 'test' as "groupType", 'test db' as title, 
'testhost' as hostname, 5432 as port, 'test-name' as name, 'test-username' as username, 'test-password' as password, 'postgresql' as type,
2 as "maxOpenConns", 1 as "maxIdleConns", 600 as "connMaxLifetimeInSeconds", 60 as "connMaxIdleTimeInSeconds"
`,
					},
				},
			},
			want: []DatabaseConfig{
				{
					DatabaseDescription: DatabaseDescription{
						DatabaseGroup: DatabaseGroup{
							GroupId:   1,
							GroupType: "test",
						},
						Title: "test db",
					},
					DatabaseConnConfig: DatabaseConnConfig{
						Hostname: "testhost",
						Port:     5432,
						Name:     "test-name",
						Username: "test-username",
						Password: "test-password",
						Type:     "postgresql",
					},
					DatabaseConnPoolConfig: DatabaseConnPoolConfig{
						MaxOpenConns:             2,
						MaxIdleConns:             1,
						ConnMaxLifetimeInSeconds: 600,
						ConnMaxIdleTimeInSeconds: 60,
					},
				},
			},
			wantErrsSize: 0,
		},
		{
			name: "datasource select with invalid query",
			args: args{
				dataSources: []DataSource{
					{
						DatabaseConnConfig: config.DatabaseConfigs[0].DatabaseConnConfig,
						Query:              `select 1`,
					},
				},
			},
			want:         []DatabaseConfig{},
			wantErrsSize: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetDatabaseConfigsFromDataSources(tt.args.dataSources)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDatabaseConfigsFromDataSources() got = %v, want %v", got, tt.want)
			}
			if len(got1) != tt.wantErrsSize {
				t.Errorf("GetDatabaseConfigsFromDataSources() got1 = %v, want %v", got1, tt.wantErrsSize)
			}
		})
	}
}

func TestGetDatabaseConfigsFromDataSource(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	config, err := readConfig("testdata/integration_config.json")
	if err != nil {
		t.Fatalf("failed to parse config file. Error: %v", err)
	}

	type args struct {
		dataSource DataSource
	}
	tests := []struct {
		name    string
		args    args
		want    []DatabaseConfig
		wantErr bool
	}{
		{
			name: "datasource select with valid query",
			args: args{
				dataSource: DataSource{
					DatabaseConnConfig: config.DatabaseConfigs[0].DatabaseConnConfig,
					Query: `
select 1 as "groupId", 'test' as "groupType", 'test db' as title, 
'testhost' as hostname, 5432 as port, 'test-name' as name, 'test-username' as username, 'test-password' as password, 'postgresql' as type,
2 as "maxOpenConns", 1 as "maxIdleConns", 600 as "connMaxLifetimeInSeconds", 60 as "connMaxIdleTimeInSeconds"
`,
				},
			},
			want: []DatabaseConfig{
				{
					DatabaseDescription: DatabaseDescription{
						DatabaseGroup: DatabaseGroup{
							GroupId:   1,
							GroupType: "test",
						},
						Title: "test db",
					},
					DatabaseConnConfig: DatabaseConnConfig{
						Hostname: "testhost",
						Port:     5432,
						Name:     "test-name",
						Username: "test-username",
						Password: "test-password",
						Type:     "postgresql",
					},
					DatabaseConnPoolConfig: DatabaseConnPoolConfig{
						MaxOpenConns:             2,
						MaxIdleConns:             1,
						ConnMaxLifetimeInSeconds: 600,
						ConnMaxIdleTimeInSeconds: 60,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "datasource select with invalid query",
			args: args{
				dataSource: DataSource{
					DatabaseConnConfig: config.DatabaseConfigs[0].DatabaseConnConfig,
					Query:              `select 1`,
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDatabaseConfigsFromDataSource(tt.args.dataSource)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDatabaseConfigsFromDataSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetDatabaseConfigsFromDataSource() got = %v, want %v", got, tt.want)
			}
		})
	}
}
