package yiigo

import (
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestInsertSQL(t *testing.T) {
	type args struct {
		table string
		data  interface{}
	}
	type Person struct {
		ID     int    `db:"id"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []interface{}
	}{
		{
			name: "t1",
			args: args{
				table: "person",
				data: &Person{
					ID:     1,
					Name:   "IIInsomnia",
					Gender: "M",
					Age:    29,
				},
			},
			want:  "INSERT INTO `person` (`id`, `name`, `gender`, `age`) VALUES (?, ?, ?, ?)",
			want1: []interface{}{1, "IIInsomnia", "M", 29},
		},
		{
			name: "t2",
			args: args{
				table: "person",
				data: []*Person{
					{
						ID:     1,
						Name:   "IIInsomnia",
						Gender: "M",
						Age:    29,
					},
					{
						ID:     2,
						Name:   "test",
						Gender: "W",
						Age:    20,
					},
				},
			},
			want:  "INSERT INTO `person` (`id`, `name`, `gender`, `age`) VALUES (?, ?, ?, ?), (?, ?, ?, ?)",
			want1: []interface{}{1, "IIInsomnia", "M", 29, 2, "test", "W", 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := InsertSQL(tt.args.table, tt.args.data)
			if got != tt.want {
				t.Errorf("InsertSQL() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("InsertSQL() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestUpdateSQL(t *testing.T) {
	type args struct {
		query string
		data  interface{}
		args  []interface{}
	}
	type Person struct {
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []interface{}
	}{
		{
			name: "t1",
			args: args{
				query: "UPDATE `person` SET ? WHERE `id` = ?",
				data: &Person{
					Name:   "IIInsomnia",
					Gender: "M",
					Age:    29,
				},
				args: []interface{}{1},
			},
			want:  "UPDATE `person` SET `name` = ?, `gender` = ?, `age` = ? WHERE `id` = ?",
			want1: []interface{}{"IIInsomnia", "M", 29, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := UpdateSQL(tt.args.query, tt.args.data, tt.args.args...)
			if got != tt.want {
				t.Errorf("UpdateSQL() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("UpdateSQL() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPGInsertSQL(t *testing.T) {
	type args struct {
		table string
		data  interface{}
	}
	type Person struct {
		ID     int    `db:"id"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []interface{}
	}{
		{
			name: "t1",
			args: args{
				table: "person",
				data: &Person{
					ID:     1,
					Name:   "IIInsomnia",
					Gender: "M",
					Age:    29,
				},
			},
			want:  `INSERT INTO "person" ("id", "name", "gender", "age") VALUES ($1, $2, $3, $4) RETURNING "id"`,
			want1: []interface{}{1, "IIInsomnia", "M", 29},
		},
		{
			name: "t2",
			args: args{
				table: "person",
				data: []*Person{
					{
						ID:     1,
						Name:   "IIInsomnia",
						Gender: "M",
						Age:    29,
					},
					{
						ID:     2,
						Name:   "test",
						Gender: "W",
						Age:    20,
					},
				},
			},
			want:  `INSERT INTO "person" ("id", "name", "gender", "age") VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)`,
			want1: []interface{}{1, "IIInsomnia", "M", 29, 2, "test", "W", 20},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := PGInsertSQL(tt.args.table, tt.args.data)
			if got != tt.want {
				t.Errorf("PGInsertSQL() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("PGInsertSQL() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestPGUpdateSQL(t *testing.T) {
	type args struct {
		query string
		data  interface{}
		args  []interface{}
	}
	type Person struct {
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []interface{}
	}{
		{
			name: "t1",
			args: args{
				query: `UPDATE "person" SET $1 WHERE "id" = $2`,
				data: &Person{
					Name:   "IIInsomnia",
					Gender: "M",
					Age:    29,
				},
				args: []interface{}{1},
			},
			want:  `UPDATE "person" SET "name" = $1, "gender" = $2, "age" = $3 WHERE "id" = $4`,
			want1: []interface{}{"IIInsomnia", "M", 29, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := PGUpdateSQL(tt.args.query, tt.args.data, tt.args.args...)
			if got != tt.want {
				t.Errorf("PGUpdateSQL() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("PGUpdateSQL() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
