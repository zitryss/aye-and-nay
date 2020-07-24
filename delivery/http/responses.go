//go:generate easyjson -all responses.go

package http

type albumResponse struct {
	Album struct {
		Id string `json:"id"`
	} `json:"album"`
}

type readyResponse struct {
	Album struct {
		Progress float64 `json:"progress"`
	} `json:"album"`
}

type pairResponse struct {
	Img1 struct {
		Token string `json:"token"`
		Src   string `json:"src"`
	} `json:"img1"`
	Img2 struct {
		Token string `json:"token"`
		Src   string `json:"src"`
	} `json:"img2"`
}

type voteResponse struct {
}

type topResponse struct {
	Images []image `json:"images"`
}

type image struct {
	Src    string  `json:"src"`
	Rating float64 `json:"rating"`
}
