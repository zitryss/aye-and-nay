package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
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
	input := func(r *http.Request, ps httprouter.Params) (context.Context, albumRequest, error) {
		ctx := r.Context()
		maxMemory := int64(c.conf.maxNumberOfFiles) * c.conf.maxFileSize
		err := r.ParseMultipartForm(maxMemory)
		if err != nil {
			return nil, albumRequest{}, errors.Wrap(err)
		}
		fhs := r.MultipartForm.File["images"]
		if len(fhs) < 2 {
			return nil, albumRequest{}, errors.Wrap(model.ErrNotEnoughImages)
		}
		if len(fhs) > c.conf.maxNumberOfFiles {
			return nil, albumRequest{}, errors.Wrap(model.ErrTooManyImages)
		}
		req := albumRequest{ff: make([]model.File, 0, len(fhs)), multi: r.MultipartForm}
		for _, fh := range fhs {
			if fh.Size > c.conf.maxFileSize {
				return nil, albumRequest{}, errors.Wrap(model.ErrImageTooBig)
			}
			f, err := fh.Open()
			if err != nil {
				return nil, albumRequest{}, errors.Wrap(err)
			}
			b := make([]byte, 512)
			_, err = f.Read(b)
			if err != nil {
				_ = f.Close()
				for _, f := range req.ff {
					_ = f.Reader.(io.Closer).Close()
				}
				_ = req.multi.RemoveAll()
				return nil, albumRequest{}, errors.Wrap(err)
			}
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				_ = f.Close()
				for _, f := range req.ff {
					_ = f.Reader.(io.Closer).Close()
				}
				_ = req.multi.RemoveAll()
				return nil, albumRequest{}, errors.Wrap(err)
			}
			typ := http.DetectContentType(b)
			if !strings.HasPrefix(typ, "image/") {
				_ = f.Close()
				for _, f := range req.ff {
					_ = f.Reader.(io.Closer).Close()
				}
				_ = req.multi.RemoveAll()
				return nil, albumRequest{}, errors.Wrap(model.ErrNotImage)
			}
			req.ff = append(req.ff, model.File{Reader: f, Size: fh.Size})
		}
		return ctx, req, nil
	}
	process := func(ctx context.Context, req albumRequest) (albumResponse, error) {
		defer func() {
			for _, f := range req.ff {
				_ = f.Reader.(io.Closer).Close()
			}
			_ = req.multi.RemoveAll()
		}()
		album, err := c.serv.Album(ctx, req.ff)
		if err != nil {
			return albumResponse{}, errors.Wrap(err)
		}
		resp := albumResponse{}
		resp.Album.Id = album
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp albumResponse) error {
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

func (c *controller) handleReady() httprouter.Handle {
	input := func(r *http.Request, ps httprouter.Params) (context.Context, readyRequest, error) {
		ctx := r.Context()
		req := readyRequest{}
		req.album.id = ps.ByName("album")
		return ctx, req, nil
	}
	process := func(ctx context.Context, req readyRequest) (readyResponse, error) {
		p, err := c.serv.Progress(ctx, req.album.id)
		if err != nil {
			return readyResponse{}, errors.Wrap(err)
		}
		resp := readyResponse{}
		resp.Album.Progress = p
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp readyResponse) error {
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

func (c *controller) handlePair() httprouter.Handle {
	input := func(r *http.Request, ps httprouter.Params) (context.Context, pairRequest, error) {
		ctx := r.Context()
		req := pairRequest{}
		req.album.id = ps.ByName("album")
		return ctx, req, nil
	}
	process := func(ctx context.Context, req pairRequest) (pairResponse, error) {
		img1, img2, err := c.serv.Pair(ctx, req.album.id)
		if err != nil {
			return pairResponse{}, errors.Wrap(err)
		}
		resp := pairResponse{}
		resp.Img1.Src = img1.Src
		resp.Img1.Token = img1.Token
		resp.Img2.Src = img2.Src
		resp.Img2.Token = img2.Token
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp pairResponse) error {
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
	input := func(r *http.Request, ps httprouter.Params) (context.Context, voteRequest, error) {
		ctx := r.Context()
		req := voteRequest{}
		req.Album.id = ps.ByName("album")
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return nil, voteRequest{}, errors.Wrap(err)
		}
		return ctx, req, nil
	}
	process := func(ctx context.Context, req voteRequest) (voteResponse, error) {
		err := c.serv.Vote(ctx, req.Album.id, req.Album.ImgFrom.Token, req.Album.ImgTo.Token)
		if err != nil {
			return voteResponse{}, errors.Wrap(err)
		}
		resp := voteResponse{}
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp voteResponse) error {
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
	input := func(r *http.Request, ps httprouter.Params) (context.Context, topRequest, error) {
		ctx := r.Context()
		req := topRequest{}
		req.album.id = ps.ByName("album")
		return ctx, req, nil
	}
	process := func(ctx context.Context, req topRequest) (topResponse, error) {
		imgs, err := c.serv.Top(ctx, req.album.id)
		if err != nil {
			return topResponse{}, errors.Wrap(err)
		}
		resp := topResponse{Images: make([]image, 0, len(imgs))}
		for _, img := range imgs {
			image := image{img.Src, img.Rating}
			resp.Images = append(resp.Images, image)
		}
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp topResponse) error {
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
