package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToInsert(t *testing.T) {
	type User struct {
		ID     int    `db:"-"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}

	query, binds := NewSQLBuilder(MySQL).Table("user").ToInsert(&User{
		Name:   "shenghui0779",
		Gender: "M",
		Age:    29,
	})

	assert.Equal(t, "INSERT INTO user ( name, gender, age ) VALUES ( ?, ?, ? )", query)
	assert.Equal(t, []interface{}{"shenghui0779", "M", 29}, binds)
}

func TestToBatchInsert(t *testing.T) {
	type User struct {
		ID     int    `db:"-"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}

	query, binds := NewSQLBuilder(MySQL).Table("user").ToBatchInsert([]*User{
		{
			Name:   "shenghui0779",
			Gender: "M",
			Age:    29,
		},
		{
			Name:   "test",
			Gender: "W",
			Age:    20,
		},
	})

	assert.Equal(t, "INSERT INTO user ( name, gender, age ) VALUES ( ?, ?, ? ), ( ?, ?, ? )", query)
	assert.Equal(t, []interface{}{"shenghui0779", "M", 29, "test", "W", 20}, binds)
}

func TestToUpdate(t *testing.T) {
	type User struct {
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}

	query, binds := NewSQLBuilder(MySQL).Table("user").Where("id = ?", 1).ToUpdate(&User{
		Name:   "shenghui0779",
		Gender: "M",
		Age:    29,
	})

	assert.Equal(t, "UPDATE user SET name = ?, gender = ?, age = ? WHERE id = ?", query)
	assert.Equal(t, []interface{}{"shenghui0779", "M", 29, 1}, binds)
}
