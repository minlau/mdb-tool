package main

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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
			"username-t:password-t@tcp(hostname-t:1234)/name-t",
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
	tests := []struct {
		name    string
		cfg     DatabaseConnConfig
		want    bool
		wantErr bool
	}{
		{
			"connects to db",
			DatabaseConnConfig{
				Hostname: "localhost",
				Name:     "test",
				Port:     5432,
				Type:     "postgresql",
				Username: "postgres",
				Password: "admin",
			},
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
