//go:generate $GOPATH/bin/easyjson requests.go

package http

import (
	"mime/multipart"
	"time"

	"github.com/zitryss/aye-and-nay/domain/model"
)

type albumRequest struct {
	ff    []model.File
	multi *multipart.Form
	dur   time.Duration
}

type readyRequest struct {
	album struct {
		id string
	}
}

type pairRequest struct {
	album struct {
		id string
	}
}

//easyjson:json
type voteRequest struct {
	Album struct {
		id      string
		ImgFrom struct {
			Token string `json:"token"`
		} `json:"imgFrom"`
		ImgTo struct {
			Token string `json:"token"`
		} `json:"imgTo"`
	} `json:"album"`
}

type topRequest struct {
	album struct {
		id string
	}
}
