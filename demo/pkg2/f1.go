package pkg2

import mp "github.com/jdxj/study-ast/demo/pkg1"

type People struct {
	Jump mp.Jump `json:"jump"` // 跳
	*mp.Sing
	Animal mp.Animal `json:"animal"` // 动物
	Age    int       `json:"age"`    // 年龄
	Rap    Rap       `json:"rap"`    // 低
	// *Rap
}

type Rap struct {
	Low int `json:"low"` // 低音
}
