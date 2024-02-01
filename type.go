package yiigo

import "fmt"

// X 类型别名
type X map[string]any

// DBDriver 数据库驱动
type DBDriver string

const (
	MySQL    DBDriver = "mysql"
	Postgres DBDriver = "pgx"
	SQLite   DBDriver = "sqlite3"
)

// Quantity 字节大小
type Quantity int64

// String 实现 Stringer 接口
func (q Quantity) String() string {
	if q >= TiB {
		return fmt.Sprintf("%.2fTB", float64(q)/float64(TiB))
	}

	if q >= GiB {
		return fmt.Sprintf("%.2fGB", float64(q)/float64(GiB))
	}

	if q >= MiB {
		return fmt.Sprintf("%.2fMB", float64(q)/float64(MiB))
	}

	if q >= KiB {
		return fmt.Sprintf("%.2fKB", float64(q)/float64(KiB))
	}

	return fmt.Sprintf("%dB", q)
}

const (
	// TiB - TebiByte size
	TiB Quantity = 1024 * 1024 * 1024 * 1024
	// GiB - GibiByte size
	GiB Quantity = 1024 * 1024 * 1024
	// MiB - MebiByte size
	MiB Quantity = 1024 * 1024
	// KiB - KibiByte size
	KiB Quantity = 1024
)
