package http

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/unit"
)

func newController(
	serv model.Servicer,
) controller {
	conf := newContrConfig()
	return controller{conf, serv}
}

type controller struct {
	conf contrConfig
	serv model.Servicer
}

func (c *controller) handleAlbum() httprouter.Handle {
	type request struct {
		files [][]byte
	}
	type response struct {
		Album struct {
			Id string `json:"id"`
		} `json:"album"`
	}
	input := func(r *http.Request, ps httprouter.Params) (context.Context, request, error) {
		ctx := r.Context()
		err := r.ParseMultipartForm(32 * unit.MB)
		if err != nil {
			return nil, request{}, errors.Wrap(err)
		}
		fhs := r.MultipartForm.File["images"]
		if len(fhs) < 2 {
			return nil, request{}, errors.Wrap(model.ErrNotEnoughImages)
		}
		if len(fhs) > c.conf.maxNumberOfFiles {
			return nil, request{}, errors.Wrap(model.ErrTooManyImages)
		}
		req := request{files: make([][]byte, 0, len(fhs))}
		for _, fh := range fhs {
			if fh.Size > c.conf.maxFileSize {
				return nil, request{}, errors.Wrap(model.ErrImageTooBig)
			}
			f, err := fh.Open()
			if err != nil {
				return nil, request{}, errors.Wrap(err)
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				_ = f.Close()
				return nil, request{}, errors.Wrap(err)
			}
			typ := http.DetectContentType(b)
			if !strings.HasPrefix(typ, "image/") {
				_ = f.Close()
				return nil, request{}, errors.Wrap(model.ErrNotImage)
			}
			req.files = append(req.files, b)
			err = f.Close()
			if err != nil {
				return nil, request{}, errors.Wrap(err)
			}
		}
		return ctx, req, nil
	}
	process := func(ctx context.Context, req request) (response, error) {
		album, err := c.serv.Album(ctx, req.files)
		if err != nil {
			return response{}, errors.Wrap(err)
		}
		resp := response{}
		resp.Album.Id = album
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp response) error {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(201)
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	}
	return handleHttpRouterError(
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
			ctx, req, err := input(r, ps)
			if err != nil {
				return errors.Wrap(err)
			}
			resp, err := process(ctx, req)
			if err != nil {
				return errors.Wrap(err)
			}
			err = output(ctx, w, resp)
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		},
	)
}

func (c *controller) handlePair() httprouter.Handle {
	type request struct {
		album struct {
			id string
		}
	}
	type response struct {
		Img1 struct {
			Token string `json:"token"`
			Src   string `json:"src"`
		} `json:"img1"`
		Img2 struct {
			Token string `json:"token"`
			Src   string `json:"src"`
		} `json:"img2"`
	}
	input := func(r *http.Request, ps httprouter.Params) (context.Context, request, error) {
		ctx := r.Context()
		req := request{}
		req.album.id = ps.ByName("album")
		return ctx, req, nil
	}
	process := func(ctx context.Context, req request) (response, error) {
		img1, img2, err := c.serv.Pair(ctx, req.album.id)
		if err != nil {
			return response{}, errors.Wrap(err)
		}
		resp := response{}
		resp.Img1.Src = img1.Src
		resp.Img1.Token = img1.Token
		resp.Img2.Src = img2.Src
		resp.Img2.Token = img2.Token
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp response) error {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	}
	return handleHttpRouterError(
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
			ctx, req, err := input(r, ps)
			if err != nil {
				return errors.Wrap(err)
			}
			resp, err := process(ctx, req)
			if err != nil {
				return errors.Wrap(err)
			}
			err = output(ctx, w, resp)
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		},
	)
}

func (c *controller) handleVote() httprouter.Handle {
	type request struct {
		Album struct {
			id      string
			ImgFrom struct {
				Token string
			}
			ImgTo struct {
				Token string
			}
		}
	}
	type response struct {
	}
	input := func(r *http.Request, ps httprouter.Params) (context.Context, request, error) {
		ctx := r.Context()
		req := request{}
		req.Album.id = ps.ByName("album")
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return nil, request{}, errors.Wrap(err)
		}
		return ctx, req, nil
	}
	process := func(ctx context.Context, req request) (response, error) {
		err := c.serv.Vote(ctx, req.Album.id, req.Album.ImgFrom.Token, req.Album.ImgTo.Token)
		if err != nil {
			return response{}, errors.Wrap(err)
		}
		resp := response{}
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp response) error {
		return nil
	}
	return handleHttpRouterError(
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
			ctx, req, err := input(r, ps)
			if err != nil {
				return errors.Wrap(err)
			}
			resp, err := process(ctx, req)
			if err != nil {
				return errors.Wrap(err)
			}
			err = output(ctx, w, resp)
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		},
	)
}

func (c *controller) handleTop() httprouter.Handle {
	type request struct {
		album struct {
			id string
		}
	}
	type image struct {
		Src    string  `json:"src"`
		Rating float64 `json:"rating"`
	}
	type response struct {
		Images []image `json:"images"`
	}
	input := func(r *http.Request, ps httprouter.Params) (context.Context, request, error) {
		ctx := r.Context()
		req := request{}
		req.album.id = ps.ByName("album")
		return ctx, req, nil
	}
	process := func(ctx context.Context, req request) (response, error) {
		imgs, err := c.serv.Top(ctx, req.album.id)
		if err != nil {
			return response{}, errors.Wrap(err)
		}
		resp := response{Images: make([]image, 0, len(imgs))}
		for _, img := range imgs {
			image := image{img.Src, img.Rating}
			resp.Images = append(resp.Images, image)
		}
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp response) error {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err := json.NewEncoder(w).Encode(resp)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	}
	return handleHttpRouterError(
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
			ctx, req, err := input(r, ps)
			if err != nil {
				return errors.Wrap(err)
			}
			resp, err := process(ctx, req)
			if err != nil {
				return errors.Wrap(err)
			}
			err = output(ctx, w, resp)
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		},
	)
}
