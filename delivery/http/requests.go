//go:generate $GOPATH/bin/easyjson -all requests.go

package http

type albumRequest struct {
	files [][]byte
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
