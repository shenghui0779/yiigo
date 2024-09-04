package yiigo

import (
	"encoding/json"
	"testing"
)

type Demo struct {
	ID   int64  `json:"id"`
	Pid  int64  `json:"pid"`
	Name string `json:"name"`
}

func (d *Demo) GetID() int64 {
	return d.ID
}

func (d *Demo) GetPid() int64 {
	return d.Pid
}

func TestTree(t *testing.T) {
	var data = map[int64][]*Demo{
		0: {
			{
				ID:   1,
				Pid:  0,
				Name: "1",
			},
			{
				ID:   2,
				Pid:  0,
				Name: "2",
			},
		},
		1: {
			{
				ID:   3,
				Pid:  1,
				Name: "3",
			},
			{
				ID:   4,
				Pid:  1,
				Name: "4",
			},
		},
		2: {
			{
				ID:   5,
				Pid:  2,
				Name: "5",
			},
			{
				ID:   6,
				Pid:  2,
				Name: "6",
			},
		},
		3: {
			{
				ID:   7,
				Pid:  3,
				Name: "7",
			},
			{
				ID:   8,
				Pid:  3,
				Name: "8",
			},
		},
		4: {
			{
				ID:   9,
				Pid:  4,
				Name: "9",
			},
			{
				ID:   10,
				Pid:  4,
				Name: "10",
			},
		},
	}

	tree := BuildLevelTree(data, 0)
	b, _ := json.MarshalIndent(tree, "", "  ")
	t.Log(string(b))
}
