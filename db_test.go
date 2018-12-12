package main

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConnectionUrl(t *testing.T) {
	var tests = []struct {
		config      DatabaseConnConfig
		expected    string
		expectedErr error
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
		url, err := tt.config.getConnectionUrl()

		assert.Equal(t, tt.expected, url)
		assert.IsType(t, tt.expectedErr, err)
	}
}

func TestGetDriverName(t *testing.T) {
	var tests = []struct {
		config      DatabaseConnConfig
		expected    string
		expectedErr error
	}{
		{
			DatabaseConnConfig{
				Type: "postgresql",
			},
			"postgres",
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
		url, err := tt.config.getDriverName()

		assert.Equal(t, tt.expected, url)
		assert.IsType(t, tt.expectedErr, err)
	}
}
