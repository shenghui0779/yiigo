package yiigo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToQuery(t *testing.T) {
	ctx := context.TODO()

	query, binds := builder.Wrap(
		Table("user"),
		Where("id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		Where("name = ? AND age > ?", "yiigo", 20),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user WHERE name = ? AND age > ?", query)
	assert.Equal(t, []interface{}{"yiigo", 20}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		WhereIn("age IN (?)", []int{20, 30}),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user WHERE age IN (?, ?)", query)
	assert.Equal(t, []interface{}{20, 30}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		Select("id", "name", "age"),
		Where("id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT id, name, age FROM user WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		Distinct("name"),
		Where("id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT DISTINCT name FROM user WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		Join("address", "user.id = address.user_id"),
		Where("user.id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user INNER JOIN address ON user.id = address.user_id WHERE user.id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		LeftJoin("address", "user.id = address.user_id"),
		Where("user.id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user LEFT JOIN address ON user.id = address.user_id WHERE user.id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		RightJoin("address", "user.id = address.user_id"),
		Where("user.id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user RIGHT JOIN address ON user.id = address.user_id WHERE user.id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		FullJoin("address", "user.id = address.user_id"),
		Where("user.id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user FULL JOIN address ON user.id = address.user_id WHERE user.id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, _ = builder.Wrap(
		Table("sizes"),
		CrossJoin("colors"),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM sizes CROSS JOIN colors", query)

	query, binds = builder.Wrap(
		Table("user"),
		LeftJoin("address", "user.id = address.user_id"),
		RightJoin("company", "user.id = company.user_id"),
		Where("user.id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user LEFT JOIN address ON user.id = address.user_id RIGHT JOIN company ON user.id = company.user_id WHERE user.id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("address"),
		Select("user_id", "COUNT(*) AS total"),
		GroupBy("user_id"),
		Having("user_id = ?", 1),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT user_id, COUNT(*) AS total FROM address GROUP BY user_id HAVING user_id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		Where("age > ?", 20),
		OrderBy("age ASC", "id DESC"),
		Offset(5),
		Limit(10),
	).ToQuery(ctx)

	assert.Equal(t, "SELECT * FROM user WHERE age > ? ORDER BY age ASC, id DESC LIMIT ? OFFSET ?", query)
	assert.Equal(t, []interface{}{20, 10, 5}, binds)

	query, binds = builder.Wrap(
		Table("user_0"),
		Where("id = ?", 1),
		Union(builder.Wrap(Table("user_1"), Where("id = ?", 2))),
	).ToQuery(ctx)

	assert.Equal(t, "(SELECT * FROM user_0 WHERE id = ?) UNION (SELECT * FROM user_1 WHERE id = ?)", query)
	assert.Equal(t, []interface{}{1, 2}, binds)

	query, binds = builder.Wrap(
		Table("user_0"),
		Where("id = ?", 1),
		UnionAll(builder.Wrap(Table("user_1"), Where("id = ?", 2))),
	).ToQuery(ctx)

	assert.Equal(t, "(SELECT * FROM user_0 WHERE id = ?) UNION ALL (SELECT * FROM user_1 WHERE id = ?)", query)
	assert.Equal(t, []interface{}{1, 2}, binds)

	query, binds = builder.Wrap(
		Table("user_0"),
		WhereIn("age IN (?)", []int{10, 20}),
		Limit(5),
		Union(
			builder.Wrap(
				Table("user_1"),
				WhereIn("age IN (?)", []int{30, 40}),
				Limit(5),
			),
		),
	).ToQuery(ctx)

	assert.Equal(t, "(SELECT * FROM user_0 WHERE age IN (?, ?) LIMIT ?) UNION (SELECT * FROM user_1 WHERE age IN (?, ?) LIMIT ?)", query)
	assert.Equal(t, []interface{}{10, 20, 5, 30, 40, 5}, binds)

	query, binds = builder.Wrap(
		Table("user_0"),
		Where("id = ?", 1),
		Union(builder.Wrap(Table("user_1"), Where("id = ?", 2))),
		UnionAll(builder.Wrap(Table("user_2"), Where("id = ?", 3))),
	).ToQuery(ctx)

	assert.Equal(t, "(SELECT * FROM user_0 WHERE id = ?) UNION (SELECT * FROM user_1 WHERE id = ?) UNION ALL (SELECT * FROM user_2 WHERE id = ?)", query)
	assert.Equal(t, []interface{}{1, 2, 3}, binds)
}

func TestToInsert(t *testing.T) {
	ctx := context.TODO()

	type User struct {
		ID     int    `db:"-"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
		Phone  string `db:"phone,omitempty"`
	}

	query, binds := builder.Wrap(Table("user")).ToInsert(ctx, &User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
	})

	assert.Equal(t, "INSERT INTO user (name, gender, age) VALUES (?, ?, ?)", query)
	assert.Equal(t, []interface{}{"yiigo", "M", 29}, binds)

	query, binds = builder.Wrap(Table("user")).ToInsert(ctx, &User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
		Phone:  "13605109425",
	})

	assert.Equal(t, "INSERT INTO user (name, gender, age, phone) VALUES (?, ?, ?, ?)", query)
	assert.Equal(t, []interface{}{"yiigo", "M", 29, "13605109425"}, binds)

	// map 字段顺序不一定
	// query, binds = builder.Wrap(Table("user")).ToInsert(X{
	// 	"age":    29,
	// 	"gender": "M",
	// 	"name":   "yiigo",
	// })
	//
	// assert.Equal(t, "INSERT INTO user (age, gender, name) VALUES (?, ?, ?)", query)
	// assert.Equal(t, []interface{}{29, "M", "yiigo"}, binds)
}

func TestToBatchInsert(t *testing.T) {
	ctx := context.TODO()

	type User struct {
		ID     int    `db:"-"`
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
		Phone  string `db:"phone,omitempty"`
	}

	query, binds := builder.Wrap(Table("user")).ToBatchInsert(ctx, []*User{
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

	assert.Equal(t, "INSERT INTO user (name, gender, age) VALUES (?, ?, ?), (?, ?, ?)", query)
	assert.Equal(t, []interface{}{"yiigo", "M", 29, "test", "W", 20}, binds)

	query, binds = builder.Wrap(Table("user")).ToBatchInsert(ctx, []*User{
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

	assert.Equal(t, "INSERT INTO user (name, gender, age, phone) VALUES (?, ?, ?, ?), (?, ?, ?, ?)", query)
	assert.Equal(t, []interface{}{"yiigo", "M", 29, "13605109425", "test", "W", 20, "13605105471"}, binds)

	// map 字段顺序不一定
	// query, binds = builder.Wrap(Table("user")).ToBatchInsert([]X{
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
	// assert.Equal(t, "INSERT INTO user (age, gender, name) VALUES (?, ?, ?), (?, ?, ?)", query)
	// assert.Equal(t, []interface{}{29, "M", "yiigo", 20, "W", "test"}, binds)
}

func TestToUpdate(t *testing.T) {
	ctx := context.TODO()

	type User struct {
		Name   string `db:"name"`
		Gender string `db:"gender"`
		Age    int    `db:"age"`
		Phone  string `db:"phone,omitempty"`
	}

	query, binds := builder.Wrap(
		Table("user"),
		Where("id = ?", 1),
	).ToUpdate(ctx, &User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
	})

	assert.Equal(t, "UPDATE user SET name = ?, gender = ?, age = ? WHERE id = ?", query)
	assert.Equal(t, []interface{}{"yiigo", "M", 29, 1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		Where("id = ?", 1),
	).ToUpdate(ctx, &User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
		Phone:  "13605109425",
	})

	assert.Equal(t, "UPDATE user SET name = ?, gender = ?, age = ?, phone = ? WHERE id = ?", query)
	assert.Equal(t, []interface{}{"yiigo", "M", 29, "13605109425", 1}, binds)

	// map 字段顺序不一定
	// query, binds = builder.Wrap(
	// 	Table("user"),
	// 	Where("id = ?", 1),
	// ).ToUpdate(X{
	// 	"age":    29,
	// 	"gender": "M",
	// 	"name":   "yiigo",
	// })
	//
	// assert.Equal(t, "UPDATE user SET age = ?, gender = ?, name = ? WHERE id = ?", query)
	// assert.Equal(t, []interface{}{29, "M", "yiigo", 1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		WhereIn("id IN (?)", []int{1, 2}),
	).ToUpdate(ctx, &User{
		Name:   "yiigo",
		Gender: "M",
		Age:    29,
	})

	assert.Equal(t, "UPDATE user SET name = ?, gender = ?, age = ? WHERE id IN (?, ?)", query)
	assert.Equal(t, []interface{}{"yiigo", "M", 29, 1, 2}, binds)

	query, binds = builder.Wrap(
		Table("product"),
		Where("id = ?", 1),
	).ToUpdate(ctx, X{"price": Clause("price * ? + ?", 2, 100)})

	assert.Equal(t, "UPDATE product SET price = price * ? + ? WHERE id = ?", query)
	assert.Equal(t, []interface{}{2, 100, 1}, binds)
}

func TestToDelete(t *testing.T) {
	ctx := context.TODO()

	query, binds := builder.Wrap(
		Table("user"),
		Where("id = ?", 1),
	).ToDelete(ctx)

	assert.Equal(t, "DELETE FROM user WHERE id = ?", query)
	assert.Equal(t, []interface{}{1}, binds)

	query, binds = builder.Wrap(
		Table("user"),
		WhereIn("id IN (?)", []int{1, 2}),
	).ToDelete(ctx)

	assert.Equal(t, "DELETE FROM user WHERE id IN (?, ?)", query)
	assert.Equal(t, []interface{}{1, 2}, binds)
}

func TestToTruncate(t *testing.T) {
	assert.Equal(t, "TRUNCATE user", builder.Wrap(Table("user")).ToTruncate(context.TODO()))
}
