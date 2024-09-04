package yiigo

// ILevelNode 层级树泛型约束
type ILevelNode interface {
	GetID() int64
	GetPid() int64
}

// LevelTree 菜单或分类层级树
type LevelTree[T ILevelNode] struct {
	Data     T
	Children []*LevelTree[T]
}

// BuildLevelTree 构建菜单或分类层级树（data=按pid归类后的数据, pid=树的起始ID）
func BuildLevelTree[T ILevelNode](data map[int64][]T, pid int64) []*LevelTree[T] {
	nodes := data[pid]
	count := len(nodes)
	root := make([]*LevelTree[T], 0, count)
	for i := 0; i < count; i++ {
		node := nodes[i]
		root = append(root, &LevelTree[T]{
			Data:     node,
			Children: BuildLevelTree[T](data, node.GetID()),
		})
	}
	return root
}
