package pkg1

import "github.com/jdxj/study-ast/demo/pkg3"

type Animal struct {
	Name string `json:"name"` // 名字
}

type Sing struct {
	Title string      `json:"title"` // 标题
	Id    pkg3.Object `json:"id"`    // id
}

type Jump struct {
	High int `json:"high"` // 跳高
}
