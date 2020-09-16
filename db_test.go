package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToQuery(t *testing.T) {
	query, binds := builder.Table("user").Where("id = ?", 1).ToQuery()
	assert.Equal(t, "SELECT * FROM user WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Table("user").Where("name = ? AND age > ?", "shenghui0779", 20).ToQuery()
	assert.Equal(t, "SELECT * FROM user WHERE name = ? AND age > ?", query)
	assert.Equal(t, []interface{}{"shenghui0779", 20}, binds)

	query, binds = builder.Table("user").Where("age IN (?)", []int{20, 30}).ToQuery()
	assert.Equal(t, "SELECT * FROM user WHERE age IN (?, ?)", query)
	assert.Equal(t, []interface{}{20, 30}, binds)

	query, binds = builder.Table("user").Select("id", "name", "age").Where("id = ?", 1).ToQuery()
	assert.Equal(t, "SELECT id, name, age FROM user WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Table("user").Distinct("name").Where("id = ?", 1).ToQuery()
	assert.Equal(t, "SELECT DISTINCT name FROM user WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Table("user").LeftJoin("address", "user.id = address.user_id").Where("user.id = ?", 1).ToQuery()
	assert.Equal(t, "SELECT * FROM user LEFT JOIN address ON user.id = address.user_id WHERE user.id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Table("address").Select("user_id", "COUNT(*) AS total").Group("user_id").Having("user_id = ?", 1).ToQuery()
	assert.Equal(t, "SELECT user_id, COUNT(*) AS total FROM address GROUP BY user_id HAVING user_id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Table("user").Where("age > ?", 20).Order("id DESC").Offset(5).Limit(10).ToQuery()
	assert.Equal(t, "SELECT * FROM user WHERE age > ? ORDER BY id DESC OFFSET 5 LIMIT 10", query)
	assert.Equal(t, []interface{}{20}, binds)
}

func TestToInsert(t *testing.T) {
	type User struct {
		ID     int    `db:"-"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}

	query, binds := builder.Table("user").ToInsert(&User{
		Name:   "shenghui0779",
		Gender: "M",
		Age:    29,
	})

	assert.Equal(t, "INSERT INTO user (name, gender, age) VALUES (?, ?, ?)", query)
	assert.Equal(t, []interface{}{"shenghui0779", "M", 29}, binds)
}

func TestToBatchInsert(t *testing.T) {
	type User struct {
		ID     int    `db:"-"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}

	query, binds := builder.Table("user").ToBatchInsert([]*User{
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

	assert.Equal(t, "INSERT INTO user (name, gender, age) VALUES (?, ?, ?), (?, ?, ?)", query)
	assert.Equal(t, []interface{}{"shenghui0779", "M", 29, "test", "W", 20}, binds)
}

func TestToUpdate(t *testing.T) {
	type User struct {
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
	}

	query, binds := builder.Table("user").Where("id = ?", 1).ToUpdate(&User{
		Name:   "shenghui0779",
		Gender: "M",
		Age:    29,
	})

	assert.Equal(t, "UPDATE user SET name = ?, gender = ?, age = ? WHERE id = ?", query)
	assert.Equal(t, []interface{}{"shenghui0779", "M", 29, 1}, binds)
}

func TestToDelete(t *testing.T) {
	query, binds := builder.Table("user").Where("id = ?", 1).ToDelete()
	assert.Equal(t, "DELETE FROM user WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)
}
