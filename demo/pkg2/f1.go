package pkg2

import (
	mp "github.com/jdxj/study-ast/demo/pkg1"
)

type People struct {
	Animal *mp.Animal `json:"animal"` // 动物
	Age    int        `json:"age"`    // 年龄
	mp.Sing
}
