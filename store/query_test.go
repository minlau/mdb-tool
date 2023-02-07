package store

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type invalidNumberError struct {
	Number string
}

func (e *invalidNumberError) Error() string {
	return fmt.Sprintf("invalid number: %s", e.Number)
}

func TestNewQueryError(t *testing.T) {
	type args struct {
		err error
	}
	var tests = []struct {
		name string
		args args
		want *QueryError
	}{
		{
			name: "nil error",
			args: args{
				err: nil,
			},
			want: nil,
		},
		{
			name: "simple error",
			args: args{
				err: errors.New("test"),
			},
			want: &QueryError{
				Message: "test",
				Err:     nil,
			},
		},
		{
			name: "error with public fields",
			args: args{
				err: &invalidNumberError{"a"},
			},
			want: &QueryError{
				Message: "invalid number: a",
				Err:     &invalidNumberError{"a"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewQueryError(tt.args.err); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewQueryError() = %v, want %v", got, tt.want)
			}
		})
	}
}
