package yiigo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func warpper(opts ...SQLOption) *sqlWrapper {
	wrapper := &sqlWrapper{
		columns: []string{"*"},
	}

	for _, fn := range opts {
		fn(wrapper)
	}

	return wrapper
}

func TestToQuery(t *testing.T) {
	sql, args, err := warpper(
		Table("user"),
		Where("id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user WHERE id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("user"),
		Where("name = ? AND age > ?", "yiigo", 20),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user WHERE name = ? AND age > ?", sql)
	assert.Equal(t, []any{"yiigo", 20}, args)

	sql, args, err = warpper(
		Table("user"),
		WhereIn("age IN (?)", []int{20, 30}),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user WHERE age IN (?, ?)", sql)
	assert.Equal(t, []any{20, 30}, args)

	sql, args, err = warpper(
		Table("user"),
		Select("id", "name", "age"),
		Where("id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT id, name, age FROM user WHERE id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("user"),
		Distinct("name"),
		Where("id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT DISTINCT name FROM user WHERE id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("user"),
		Join("address", "user.id = address.user_id"),
		Where("user.id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user INNER JOIN address ON user.id = address.user_id WHERE user.id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("user"),
		LeftJoin("address", "user.id = address.user_id"),
		Where("user.id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user LEFT JOIN address ON user.id = address.user_id WHERE user.id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("user"),
		RightJoin("address", "user.id = address.user_id"),
		Where("user.id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user RIGHT JOIN address ON user.id = address.user_id WHERE user.id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("user"),
		FullJoin("address", "user.id = address.user_id"),
		Where("user.id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user FULL JOIN address ON user.id = address.user_id WHERE user.id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, _, err = warpper(
		Table("sizes"),
		CrossJoin("colors"),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM sizes CROSS JOIN colors", sql)

	sql, args, err = warpper(
		Table("user"),
		LeftJoin("address", "user.id = address.user_id"),
		RightJoin("company", "user.id = company.user_id"),
		Where("user.id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user LEFT JOIN address ON user.id = address.user_id RIGHT JOIN company ON user.id = company.user_id WHERE user.id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("address"),
		Select("user_id", "COUNT(*) AS total"),
		GroupBy("user_id"),
		Having("user_id = ?", 1),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT user_id, COUNT(*) AS total FROM address GROUP BY user_id HAVING user_id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("user"),
		Where("age > ?", 20),
		OrderBy("age ASC", "id DESC"),
		Offset(5),
		Limit(10),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "SELECT * FROM user WHERE age > ? ORDER BY age ASC, id DESC LIMIT ? OFFSET ?", sql)
	assert.Equal(t, []any{20, 10, 5}, args)

	sql, args, err = warpper(
		Table("user_0"),
		Where("id = ?", 1),
		Union(warpper(Table("user_1"), Where("id = ?", 2))),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "(SELECT * FROM user_0 WHERE id = ?) UNION (SELECT * FROM user_1 WHERE id = ?)", sql)
	assert.Equal(t, []any{1, 2}, args)

	sql, args, err = warpper(
		Table("user_0"),
		Where("id = ?", 1),
		UnionAll(warpper(Table("user_1"), Where("id = ?", 2))),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "(SELECT * FROM user_0 WHERE id = ?) UNION ALL (SELECT * FROM user_1 WHERE id = ?)", sql)
	assert.Equal(t, []any{1, 2}, args)

	sql, args, err = warpper(
		Table("user_0"),
		WhereIn("age IN (?)", []int{10, 20}),
		Limit(5),
		Union(
			warpper(
				Table("user_1"),
				WhereIn("age IN (?)", []int{30, 40}),
				Limit(5),
			),
		),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "(SELECT * FROM user_0 WHERE age IN (?, ?) LIMIT ?) UNION (SELECT * FROM user_1 WHERE age IN (?, ?) LIMIT ?)", sql)
	assert.Equal(t, []any{10, 20, 5, 30, 40, 5}, args)

	sql, args, err = warpper(
		Table("user_0"),
		Where("id = ?", 1),
		Union(warpper(Table("user_1"), Where("id = ?", 2))),
		UnionAll(warpper(Table("user_2"), Where("id = ?", 3))),
	).querySQL()
	assert.Nil(t, err)
	assert.Equal(t, "(SELECT * FROM user_0 WHERE id = ?) UNION (SELECT * FROM user_1 WHERE id = ?) UNION ALL (SELECT * FROM user_2 WHERE id = ?)", sql)
	assert.Equal(t, []any{1, 2, 3}, args)
}

func TestToInsert(t *testing.T) {
	type User struct {
		ID     int    `db:"-"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
		Phone  string `db:"phone,omitempty"`
	}

	sql, args, err := warpper(Table("user")).insertSQL(&User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
	})

	assert.Nil(t, err)
	assert.Equal(t, "INSERT INTO user (name, gender, age) VALUES (?, ?, ?)", sql)
	assert.Equal(t, []any{"yiigo", "M", 29}, args)

	sql, args, err = warpper(Table("user"), Returning("id")).insertSQL(&User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
	})

	assert.Nil(t, err)
	assert.Equal(t, "INSERT INTO user (name, gender, age) VALUES (?, ?, ?) RETURNING id", sql)
	assert.Equal(t, []any{"yiigo", "M", 29}, args)

	sql, args, err = warpper(Table("user")).insertSQL(&User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
		Phone:  "13605109425",
	})

	assert.Nil(t, err)
	assert.Equal(t, "INSERT INTO user (name, gender, age, phone) VALUES (?, ?, ?, ?)", sql)
	assert.Equal(t, []any{"yiigo", "M", 29, "13605109425"}, args)

	// map 字段顺序不一定
	// sql, args, err = warpper(Table("user")).insertSQL(X{
	// 	"age":    29,
	// 	"gender": "M",
	// 	"name":   "yiigo",
	// })
	//
	// assert.Equal(t, "INSERT INTO user (age, gender, name) VALUES (?, ?, ?)", sql)
	// assert.Equal(t, []any{29, "M", "yiigo"}, args)
}

func TestToBatchInsert(t *testing.T) {
	type User struct {
		ID     int    `db:"-"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
		Phone  string `db:"phone,omitempty"`
	}

	sql, args, err := warpper(Table("user")).batchInsertSQL([]*User{
		{
			Name:   "yiigo",
			Gender: "M",
			Age:    29,
		},
		{
			Name:   "test",
			Gender: "W",
			Age:    20,
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "INSERT INTO user (name, gender, age) VALUES (?, ?, ?), (?, ?, ?)", sql)
	assert.Equal(t, []any{"yiigo", "M", 29, "test", "W", 20}, args)

	sql, args, err = warpper(Table("user")).batchInsertSQL([]*User{
		{
			Name:   "yiigo",
			Gender: "M",
			Age:    29,
			Phone:  "13605109425",
		},
		{
			Name:   "test",
			Gender: "W",
			Age:    20,
			Phone:  "13605105471",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "INSERT INTO user (name, gender, age, phone) VALUES (?, ?, ?, ?), (?, ?, ?, ?)", sql)
	assert.Equal(t, []any{"yiigo", "M", 29, "13605109425", "test", "W", 20, "13605105471"}, args)

	// map 字段顺序不一定
	// sql, args, err = warpper(Table("user")).batchInsertSQL([]X{
	// 	{
	// 		"age":    29,
	// 		"gender": "M",
	// 		"name":   "yiigo",
	// 	},
	// 	{
	// 		"age":    20,
	// 		"gender": "W",
	// 		"name":   "test",
	// 	},
	// })
	//
	// assert.Equal(t, "INSERT INTO user (age, gender, name) VALUES (?, ?, ?), (?, ?, ?)", sql)
	// assert.Equal(t, []any{29, "M", "yiigo", 20, "W", "test"}, args)
}

func TestToUpdate(t *testing.T) {
	type User struct {
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
		Phone  string `db:"phone,omitempty"`
	}

	sql, args, err := warpper(
		Table("user"),
		Where("id = ?", 1),
	).updateSQL(&User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
	})

	assert.Nil(t, err)
	assert.Equal(t, "UPDATE user SET name = ?, gender = ?, age = ? WHERE id = ?", sql)
	assert.Equal(t, []any{"yiigo", "M", 29, 1}, args)

	sql, args, err = warpper(
		Table("user"),
		Where("id = ?", 1),
	).updateSQL(&User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
		Phone:  "13605109425",
	})

	assert.Nil(t, err)
	assert.Equal(t, "UPDATE user SET name = ?, gender = ?, age = ?, phone = ? WHERE id = ?", sql)
	assert.Equal(t, []any{"yiigo", "M", 29, "13605109425", 1}, args)

	// map 字段顺序不一定
	// sql, args, err = warpper(
	// 	Table("user"),
	// 	Where("id = ?", 1),
	// ).updateSQL(X{
	// 	"age":    29,
	// 	"gender": "M",
	// 	"name":   "yiigo",
	// })
	//
	// assert.Equal(t, "UPDATE user SET age = ?, gender = ?, name = ? WHERE id = ?", sql)
	// assert.Equal(t, []any{29, "M", "yiigo", 1}, args)

	sql, args, err = warpper(
		Table("user"),
		WhereIn("id IN (?)", []int{1, 2}),
	).updateSQL(&User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
	})

	assert.Nil(t, err)
	assert.Equal(t, "UPDATE user SET name = ?, gender = ?, age = ? WHERE id IN (?, ?)", sql)
	assert.Equal(t, []any{"yiigo", "M", 29, 1, 2}, args)

	sql, args, err = warpper(
		Table("product"),
		Where("id = ?", 1),
	).updateSQL(X{"price": SQLExpr("price * ? + ?", 2, 100)})

	assert.Nil(t, err)
	assert.Equal(t, "UPDATE product SET price = price * ? + ? WHERE id = ?", sql)
	assert.Equal(t, []any{2, 100, 1}, args)
}

func TestToDelete(t *testing.T) {
	sql, args, err := warpper(
		Table("user"),
		Where("id = ?", 1),
	).deleteSQL()

	assert.Nil(t, err)
	assert.Equal(t, "DELETE FROM user WHERE id = ?", sql)
	assert.Equal(t, []any{1}, args)

	sql, args, err = warpper(
		Table("user"),
		WhereIn("id IN (?)", []int{1, 2}),
	).deleteSQL()

	assert.Nil(t, err)
	assert.Equal(t, "DELETE FROM user WHERE id IN (?, ?)", sql)
	assert.Equal(t, []any{1, 2}, args)
}

func TestToTruncate(t *testing.T) {
	assert.Equal(t, "TRUNCATE user", warpper(Table("user")).truncateSQL())
}
