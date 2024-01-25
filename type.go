package yiigo

// X 类型别名
type X map[string]any

// DBDriver 数据库驱动
type DBDriver string

const (
	MySQL    DBDriver = "mysql"
	Postgres DBDriver = "pgx"
	SQLite   DBDriver = "sqlite3"
)
