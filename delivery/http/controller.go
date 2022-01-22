package http

import (
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/base64"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func newController(
	conf ControllerConfig,
	serv domain.Servicer,
) controller {
	return controller{conf, serv}
}

type controller struct {
	conf ControllerConfig
	serv domain.Servicer
}

func (c *controller) handleAlbum() httprouter.Handle {
	input := func(r *http.Request, ps httprouter.Params) (context.Context, albumRequest, error) {
		ctx := r.Context()
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "multipart/form-data") {
			return nil, albumRequest{}, errors.Wrap(domain.ErrWrongContentType)
		}
		maxBodySize := int64(c.conf.MaxNumberOfFiles) * c.conf.MaxFileSize
		if r.ContentLength > maxBodySize {
			return nil, albumRequest{}, errors.Wrap(domain.ErrBodyTooLarge)
		}
		err := r.ParseMultipartForm(r.ContentLength)
		if err != nil {
			return nil, albumRequest{}, errors.Wrap(err)
		}
		fhs := r.MultipartForm.File["images"]
		if len(fhs) < 2 {
			return nil, albumRequest{}, errors.Wrap(domain.ErrNotEnoughImages)
		}
		if len(fhs) > c.conf.MaxNumberOfFiles {
			return nil, albumRequest{}, errors.Wrap(domain.ErrTooManyImages)
		}
		req := albumRequest{ff: make([]model.File, 0, len(fhs)), multi: r.MultipartForm}
		defer func() {
			for _, f := range req.ff {
				_ = f.Close()
			}
			_ = req.multi.RemoveAll()
		}()
		for _, fh := range fhs {
			if fh.Size > c.conf.MaxFileSize {
				return nil, albumRequest{}, errors.Wrap(domain.ErrImageTooLarge)
			}
			f, err := fh.Open()
			if err != nil {
				return nil, albumRequest{}, errors.Wrap(err)
			}
			b := make([]byte, 512)
			_, err = f.Read(b)
			if err != nil {
				_ = f.Close()
				return nil, albumRequest{}, errors.Wrap(err)
			}
			_, err = f.Seek(0, io.SeekStart)
			if err != nil {
				_ = f.Close()
				return nil, albumRequest{}, errors.Wrap(err)
			}
			typ := http.DetectContentType(b)
			if !strings.HasPrefix(typ, "image/") {
				_ = f.Close()
				return nil, albumRequest{}, errors.Wrap(domain.ErrNotImage)
			}
			closeFn := func() error {
				switch v := f.(type) {
				case *os.File:
					_ = v.Close()
					_ = os.Remove(v.Name())
				case multipart.File:
					_ = v.Close()
				default:
					panic(errors.Wrap(domain.ErrUnknown))
				}
				return nil
			}
			req.ff = append(req.ff, model.NewFile(f, closeFn, fh.Size))
		}
		vals := r.MultipartForm.Value["duration"]
		if len(vals) == 0 {
			return nil, albumRequest{}, errors.Wrap(domain.ErrDurationNotSet)
		}
		dur, err := time.ParseDuration(vals[0])
		if err != nil {
			return nil, albumRequest{}, errors.Wrap(domain.ErrDurationInvalid)
		}
		req.dur = dur
		return ctx, req, nil
	}
	process := func(ctx context.Context, req albumRequest) (albumResponse, error) {
		defer func() {
			for _, f := range req.ff {
				_ = f.Close()
			}
			_ = req.multi.RemoveAll()
		}()
		album, err := c.serv.Album(ctx, req.ff, req.dur)
		if err != nil {
			return albumResponse{}, errors.Wrap(err)
		}
		resp := albumResponse{}
		albumB64 := base64.FromUint64(album)
		resp.Album.Id = albumB64
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

func (c *controller) handleStatus() httprouter.Handle {
	input := func(r *http.Request, ps httprouter.Params) (context.Context, statusRequest, error) {
		ctx := r.Context()
		req := statusRequest{}
		req.album.id = ps.ByName("album")
		return ctx, req, nil
	}
	process := func(ctx context.Context, req statusRequest) (statusResponse, error) {
		album, err := base64.ToUint64(req.album.id)
		if err != nil {
			return statusResponse{}, errors.Wrap(err)
		}
		p, err := c.serv.Progress(ctx, album)
		if err != nil {
			return statusResponse{}, errors.Wrap(err)
		}
		resp := statusResponse{}
		resp.Album.Compression.Progress = p
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp statusResponse) error {
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
		album, err := base64.ToUint64(req.album.id)
		if err != nil {
			return pairResponse{}, errors.Wrap(err)
		}
		img1, img2, err := c.serv.Pair(ctx, album)
		if err != nil {
			return pairResponse{}, errors.Wrap(err)
		}
		resp := pairResponse{}
		resp.Album.Img1.Src = img1.Src
		img1TokenB64 := base64.FromUint64(img1.Token)
		resp.Album.Img1.Token = img1TokenB64
		resp.Album.Img2.Src = img2.Src
		img2TokenB64 := base64.FromUint64(img2.Token)
		resp.Album.Img2.Token = img2TokenB64
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

func (c *controller) handleImage() httprouter.Handle {
	input := func(r *http.Request, ps httprouter.Params) (context.Context, imageRequest, error) {
		ctx := r.Context()
		req := imageRequest{}
		req.image.token = ps.ByName("token")
		return ctx, req, nil
	}
	process := func(ctx context.Context, req imageRequest) (imageResponse, error) {
		token, err := base64.ToUint64(req.image.token)
		if err != nil {
			return imageResponse{}, errors.Wrap(err)
		}
		f, err := c.serv.Image(ctx, token)
		if err != nil {
			return imageResponse{}, errors.Wrap(err)
		}
		resp := imageResponse{f}
		return resp, nil
	}
	output := func(ctx context.Context, w http.ResponseWriter, resp imageResponse) error {
		defer resp.f.Close()
		_, err := io.Copy(w, resp.f.Reader)
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
		ct := r.Header.Get("Content-Type")
		if !strings.HasPrefix(ct, "application/json") {
			return nil, voteRequest{}, errors.Wrap(domain.ErrWrongContentType)
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			return nil, voteRequest{}, errors.Wrap(err)
		}
		return ctx, req, nil
	}
	process := func(ctx context.Context, req voteRequest) (voteResponse, error) {
		album, err := base64.ToUint64(req.Album.id)
		if err != nil {
			return voteResponse{}, errors.Wrap(err)
		}
		imgFromToken, err := base64.ToUint64(req.Album.ImgFrom.Token)
		if err != nil {
			return voteResponse{}, errors.Wrap(err)
		}
		imgToToken, err := base64.ToUint64(req.Album.ImgTo.Token)
		if err != nil {
			return voteResponse{}, errors.Wrap(err)
		}
		err = c.serv.Vote(ctx, album, imgFromToken, imgToToken)
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
		album, err := base64.ToUint64(req.album.id)
		if err != nil {
			return topResponse{}, errors.Wrap(err)
		}
		imgs, err := c.serv.Top(ctx, album)
		if err != nil {
			return topResponse{}, errors.Wrap(err)
		}
		resp := topResponse{}
		resp.Album.Images = make([]image, 0, len(imgs))
		for _, img := range imgs {
			image := image{img.Src, img.Rating}
			resp.Album.Images = append(resp.Album.Images, image)
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

func (c *controller) handleHealth() httprouter.Handle {
	return handleHttpRouterError(
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
			_, err := c.serv.Health(r.Context())
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		},
	)
}
