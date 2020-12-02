package common

import "github.com/bwmarrin/snowflake"

var node *snowflake.Node

func init() {
	tmp, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}
	node = tmp
}
func GetUUId() int64 {
	if node == nil {
		panic("id生成器为空")
	}
	id := node.Generate()
	return id.Int64()
}
