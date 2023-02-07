package store

import (
	"context"
)

type DatabaseStoreMock struct {
	AddDatabasesFunc           func(databases []DatabaseConfig)
	AddDatabaseFunc            func(config DatabaseConfig) error
	GetTablesMetadataFunc      func(groupName string, groupType string) (map[string][]string, error)
	QueryDatabaseFunc          func(ctx context.Context, groupName string, groupType string, query string) GroupQueryResult
	QueryMultipleDatabasesFunc func(ctx context.Context, groupType string, query string) []GroupQueryResult
	GetDatabaseItemsFunc       func() []DatabaseItem
}

func (d DatabaseStoreMock) AddDatabases(databases []DatabaseConfig) {
	d.AddDatabasesFunc(databases)
}

func (d DatabaseStoreMock) AddDatabase(config DatabaseConfig) error {
	return d.AddDatabaseFunc(config)
}

func (d DatabaseStoreMock) GetTablesMetadata(groupName string, groupType string) (map[string][]string, error) {
	return d.GetTablesMetadataFunc(groupName, groupType)
}

func (d DatabaseStoreMock) QueryDatabase(ctx context.Context, groupName string, groupType string, query string) GroupQueryResult {
	return d.QueryDatabaseFunc(ctx, groupName, groupType, query)
}

func (d DatabaseStoreMock) QueryMultipleDatabases(ctx context.Context, groupType string, query string) []GroupQueryResult {
	return d.QueryMultipleDatabasesFunc(ctx, groupType, query)
}

func (d DatabaseStoreMock) GetDatabaseItems() []DatabaseItem {
	return d.GetDatabaseItemsFunc()
}
