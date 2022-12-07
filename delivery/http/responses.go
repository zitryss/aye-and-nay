//go:generate $GOPATH/bin/easyjson responses.go

package http

import (
	"github.com/zitryss/aye-and-nay/domain/model"
)

//easyjson:json
type albumResponse struct {
	Album struct {
		Id string `json:"id"`
	} `json:"album"`
}

//easyjson:json
type statusResponse struct {
	Album struct {
		Compression struct {
			Progress float64 `json:"progress"`
		} `json:"compression"`
	} `json:"album"`
}

//easyjson:json
type pairResponse struct {
	Album struct {
		Img1 struct {
			Token string `json:"token"`
			Src   string `json:"src"`
		} `json:"img1"`
		Img2 struct {
			Token string `json:"token"`
			Src   string `json:"src"`
		} `json:"img2"`
	} `json:"album"`
}

type imageResponse struct {
	f model.File
}

type voteResponse struct {
}

//easyjson:json
type topResponse struct {
	Album struct {
		Images []image `json:"images"`
	} `json:"album"`
}

//easyjson:json
type image struct {
	Src    string  `json:"src"`
	Rating float64 `json:"rating"`
}

//easyjson:json
type errorResponse struct {
	Error struct {
		statusCode int
		AppCode    int    `json:"code"`
		UserMsg    string `json:"msg"`
	} `json:"error"`
}
