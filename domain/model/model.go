package model

type Album struct {
	Id     string
	Images []Image
	Edges  map[string]map[string]int
}

type Image struct {
	Id     string
	B      []byte
	Src    string
	Token  string
	Rating float64
}
