package http

import (
	"html/template"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func newHtml(
	serv model.Servicer,
) (html, error) {
	conf := newHtmlConfig()
	tmpl, err := template.ParseGlob(conf.templatesDirPath)
	if err != nil {
		return html{}, errors.Wrap(err)
	}
	return html{conf, tmpl, serv}, nil
}

type html struct {
	conf htmlConfig
	tmpl *template.Template
	serv model.Servicer
}

func (h *html) handleAlbum() httprouter.Handle {
	return handleHttpRouterError(
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
			err := h.tmpl.ExecuteTemplate(w, "index.gohtml", nil)
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		},
	)
}

func (h *html) handlePair() httprouter.Handle {
	return handleHttpRouterError(
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
			ctx := r.Context()
			album := ps.ByName("album")
			found, err := h.serv.Exists(ctx, album)
			if err != nil {
				return errors.Wrap(err)
			}
			if !found {
				return errors.Wrap(model.ErrPageNotFound)
			}
			err = h.tmpl.ExecuteTemplate(w, "pair.gohtml", nil)
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		},
	)
}

func (h *html) handleTop() httprouter.Handle {
	return handleHttpRouterError(
		func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) error {
			ctx := r.Context()
			album := ps.ByName("album")
			found, err := h.serv.Exists(ctx, album)
			if err != nil {
				return errors.Wrap(err)
			}
			if !found {
				return errors.Wrap(model.ErrPageNotFound)
			}
			err = h.tmpl.ExecuteTemplate(w, "top.gohtml", nil)
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		},
	)
}
