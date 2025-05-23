package web

import (
	"context"
	"github.com/minlau/mdb-tool/store"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func BenchmarkRequest(b *testing.B) {
	var rows []map[string]interface{}
	for i := 0; i < 100000; i++ {
		rows = append(rows, map[string]interface{}{
			"id":   i,
			"name": strconv.Itoa(i),
		})
	}

	databaseStore := &store.DatabaseStoreMock{
		QueryMultipleDatabasesFunc: func(ctx context.Context, groupType string, query string) []store.GroupQueryResult {
			return []store.GroupQueryResult{
				{
					GroupName: "bench1",
					Data: &store.QueryData{
						Columns: []store.Column{
							{Name: "id", FieldName: "id"},
							{Name: "name", FieldName: "name"},
						},
						Rows: rows,
					},
					Error: nil,
				},
			}
		},
	}

	u, err := url.Parse("localhost/query")
	if err != nil {
		b.Fatalf("failed to parse url. %e", err)
	}
	q := u.Query()
	q.Set("groupType", "bench")
	q.Set("query", "bench")
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.RequestURI(), nil)
	if err != nil {
		b.Fatalf("failed to create new request. %e", err)
	}

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
