package pkg3

type Object struct {
	Id    int `json:"id"` // 唯一
	Array []struct {
		Hello struct {
			World string `json:"world"` // ww
		} `json:"hello"` // hh
	} `json:"array"` // aa
}
