package compressor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/viper"

	"github.com/zitryss/aye-and-nay/domain/model"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

type response []struct {
	Status struct {
		Code    interface{}
		Message string
	}
	OriginalURL        string `json:"OriginalURL,omitempty"`
	LosslessURL        string `json:"LosslessURL,omitempty"`
	LossyURL           string `json:"LossyURL,omitempty"`
	WebPLosslessURL    string `json:"WebPLosslessURL,omitempty"`
	WebPLossyURL       string `json:"WebPLossyURL,omitempty"`
	OriginalSize       string `json:"OriginalSize,omitempty"`
	LosslessSize       string `json:"LosslessSize,omitempty"`
	LoselessSize       string `json:"LoselessSize,omitempty"`
	LossySize          string `json:"LossySize,omitempty"`
	WebPLosslessSize   string `json:"WebPLosslessSize,omitempty"`
	WebPLoselessSize   string `json:"WebPLoselessSize,omitempty"`
	WebPLossySize      string `json:"WebPLossySize,omitempty"`
	TimeStamp          string `json:"TimeStamp,omitempty"`
	PercentImprovement int    `json:"PercentImprovement,omitempty"`
	Key                string `json:"Key,omitempty"`
	LocalPath          string `json:"LocalPath,omitempty"`
}

func TestShortPixel(t *testing.T) {
	t.Run("Positive1", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(Png())
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "2",
						Message: "Success",
					},
					OriginalURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LosslessURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LossyURL:           mockserver1.URL,
					WebPLosslessURL:    "NA",
					WebPLossyURL:       "NA",
					OriginalSize:       "67",
					LosslessSize:       "67",
					LoselessSize:       "67",
					LossySize:          "67",
					WebPLosslessSize:   "NA",
					WebPLoselessSize:   "NA",
					WebPLossySize:      "NA",
					TimeStamp:          "2019-12-30 12:15:01",
					PercentImprovement: 0,
					Key:                "file",
					LocalPath:          "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/149/3e34ec6bc4248510450f08ee7c7711fb.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "XmktcS25JRvCKNUK", B: Png()}, {Id: "WhdU3GFF4SunrjmU", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Positive2", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(Png())
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := `[
				{
					"Status": {
						"Code": "2",
						"Message": "Success"
					},
					"OriginalURL": "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					"LosslessURL": "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					"LossyURL": "` + mockserver1.URL + `",
					"WebPLosslessURL": "NA",
					"WebPLossyURL": "NA",
					"OriginalSize": "67",
					"LosslessSize": "67",
					"LoselessSize": "67",
					"LossySize": "67",
					"WebPLosslessSize": "NA",
					"WebPLoselessSize": "NA",
					"WebPLossySize": "NA",
					"TimeStamp": "2019-12-30 12:15:01",
					"PercentImprovement": 0,
					"Key": "file",
					"localPath": "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/149/3e34ec6bc4248510450f08ee7c7711fb."
				}
			]`
			_, err := io.WriteString(w, resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "v6hJ87B9aRdbFadY", B: Png()}, {Id: "FzMjvC7RG2ZkK4Zx", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("NegativeInvalidUrl1", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		mockserver1.Close()
		viper.Set("shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "vhFf9ADJG529jyZT", B: Png()}, {Id: "rwmTG5SEBg6pLpCW", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeInvalidUrl2", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "2",
						Message: "Success",
					},
					OriginalURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LosslessURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LossyURL:           mockserver1.URL,
					WebPLosslessURL:    "NA",
					WebPLossyURL:       "NA",
					OriginalSize:       "67",
					LosslessSize:       "67",
					LoselessSize:       "67",
					LossySize:          "67",
					WebPLosslessSize:   "NA",
					WebPLoselessSize:   "NA",
					WebPLossySize:      "NA",
					TimeStamp:          "2019-12-30 12:15:01",
					PercentImprovement: 0,
					Key:                "file",
					LocalPath:          "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/149/3e34ec6bc4248510450f08ee7c7711fb.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "Pe9gQ3whftqapwv4", B: Png()}, {Id: "Ravf2deNE283RCj3", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeTimeout1", func(t *testing.T) {
		d := viper.GetDuration("shortpixel.uploadTimeout") + viper.GetDuration("shortpixel.downloadTimeout")
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * d)
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		viper.Set("shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "X6GazrWM87k6aHQX", B: Png()}, {Id: "5EG4fvaE2hb4cHV5", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeTimeout2", func(t *testing.T) {
		d := viper.GetDuration("shortpixel.uploadTimeout") + viper.GetDuration("shortpixel.downloadTimeout")
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * d)
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "2",
						Message: "Success",
					},
					OriginalURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LosslessURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LossyURL:           mockserver1.URL,
					WebPLosslessURL:    "NA",
					WebPLossyURL:       "NA",
					OriginalSize:       "67",
					LosslessSize:       "67",
					LoselessSize:       "67",
					LossySize:          "67",
					WebPLosslessSize:   "NA",
					WebPLoselessSize:   "NA",
					WebPLossySize:      "NA",
					TimeStamp:          "2019-12-30 12:15:01",
					PercentImprovement: 0,
					Key:                "file",
					LocalPath:          "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/149/3e34ec6bc4248510450f08ee7c7711fb.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "6QrQxj466YqSDJsu", B: Png()}, {Id: "HBbpvG2CnMRqWTN5", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeErrorHttpCode1", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		viper.Set("shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "Zec88XcSn37CqDXd", B: Png()}, {Id: "fV22B78ae4tJXanu", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeErrorHttpCode2", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "2",
						Message: "Success",
					},
					OriginalURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LosslessURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LossyURL:           mockserver1.URL,
					WebPLosslessURL:    "NA",
					WebPLossyURL:       "NA",
					OriginalSize:       "67",
					LosslessSize:       "67",
					LoselessSize:       "67",
					LossySize:          "67",
					WebPLosslessSize:   "NA",
					WebPLoselessSize:   "NA",
					WebPLossySize:      "NA",
					TimeStamp:          "2019-12-30 12:15:01",
					PercentImprovement: 0,
					Key:                "file",
					LocalPath:          "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/149/3e34ec6bc4248510450f08ee7c7711fb.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "Q8eFH8wk8kqadrMa", B: Png()}, {Id: "XGQW6suBDrEjw9TJ", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeInvalidJson", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			resp := `
				{
					"menu": {
						"id": "file",
						"popup": {
							"menuitem": [
								{
									"onclick": "CreateNewDoc()",
									"value": "New"
								},
								{
									"onclick": "OpenDoc()",
									"value": "Open"
								},
								{
									"onclick": "CloseDoc()",
									"value": "Close"
								}
							]
						},
						"value": "File"
					}
			`
			_, err := io.WriteString(w, resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		viper.Set("shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "n6MEpgQUqLRakKZg", B: Png()}, {Id: "XFQ3WKAa44T6hQyE", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeErrorStatusCode1", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    -110,
						Message: "Upload error.(ERR_CODE: 4)",
					},
				},
			}
			err := json.NewEncoder(w).Encode(resp[0])
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		viper.Set("shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "9JJBwxDguWSKFgg5", B: Png()}, {Id: "D4wwPBtvHPrcJUSD", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeErrorStatusCode2", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    -201,
						Message: "Invalid image format",
					},
				},
			}
			err := json.NewEncoder(w).Encode(resp[0])
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		viper.Set("shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "xGDSE45qnqTtvDSa", B: Png()}, {Id: "mq9mUaYaua7jrhe7", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrNotImage) {
			t.Error(err)
		}
	})
	t.Run("NegativeErrorStatusCode3", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    -202,
						Message: "Invalid image or unsupported image format",
					},
				},
			}
			err := json.NewEncoder(w).Encode(resp[0])
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		viper.Set("shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "jCRRfuCY7Te2FFaD", B: Png()}, {Id: "3SHSbL8fDu7VFRJE", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrNotImage) {
			t.Error(err)
		}
	})
	t.Run("PositiveRecovery1", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(Png())
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "2",
						Message: "Success",
					},
					OriginalURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LosslessURL:        "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					LossyURL:           mockserver1.URL,
					WebPLosslessURL:    "NA",
					WebPLossyURL:       "NA",
					OriginalSize:       "67",
					LosslessSize:       "67",
					LoselessSize:       "67",
					LossySize:          "67",
					WebPLosslessSize:   "NA",
					WebPLoselessSize:   "NA",
					WebPLossySize:      "NA",
					TimeStamp:          "2019-12-30 12:15:01",
					PercentImprovement: 0,
					Key:                "file",
					LocalPath:          "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/149/3e34ec6bc4248510450f08ee7c7711fb.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		fn3 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "1",
						Message: "Image scheduled for processing.",
					},
					OriginalURL: "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					Key:         "file",
					LocalPath:   "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver3 := httptest.NewServer(http.HandlerFunc(fn3))
		defer mockserver3.Close()
		viper.Set("shortpixel.url", mockserver3.URL)
		viper.Set("shortpixel.url2", mockserver2.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "rvbc6pCQjAvxvWah", B: Png()}, {Id: "XmezWbBEd4vG679X", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("PositiveRecovery2", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(Png())
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := `[
				{
					"Status": {
						"Code": "2",
						"Message": "Success"
					},
					"OriginalURL": "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					"LosslessURL": "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/3e34ec6bc4248510450f08ee7c7711fb.",
					"LossyURL": "` + mockserver1.URL + `",
					"WebPLosslessURL": "NA",
					"WebPLossyURL": "NA",
					"OriginalSize": "67",
					"LosslessSize": "67",
					"LoselessSize": "67",
					"LossySize": "67",
					"WebPLosslessSize": "NA",
					"WebPLoselessSize": "NA",
					"WebPLossySize": "NA",
					"TimeStamp": "2019-12-30 12:15:01",
					"PercentImprovement": 0,
					"Key": "file",
					"localPath": "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/149/3e34ec6bc4248510450f08ee7c7711fb."
				}
			]`
			_, err := io.WriteString(w, resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		fn3 := func(w http.ResponseWriter, r *http.Request) {
			resp := `[
				{
					"Status": {
						"Code": "1",
						"Message": "Image scheduled for processing."
					},
					"OriginalURL": "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					"Key": "file",
					"localPath": "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5."
				}
			]`
			_, err := io.WriteString(w, resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver3 := httptest.NewServer(http.HandlerFunc(fn3))
		defer mockserver3.Close()
		viper.Set("shortpixel.url", mockserver3.URL)
		viper.Set("shortpixel.url2", mockserver2.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "LkDFqwCR7457kfMr", B: Png()}, {Id: "k63K7Mtty4htFsB8", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("NegativeRecoveryInvalidUrl", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "1",
						Message: "Image scheduled for processing.",
					},
					OriginalURL: "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					Key:         "file",
					LocalPath:   "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		viper.Set("shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "MVmd6b52gyaVh37c", B: Png()}, {Id: "YHj45hpKZerNEr4Z", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeRecoveryTimeout", func(t *testing.T) {
		d := viper.GetDuration("shortpixel.uploadTimeout") + viper.GetDuration("shortpixel.downloadTimeout")
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * d)
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "1",
						Message: "Image scheduled for processing.",
					},
					OriginalURL: "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					Key:         "file",
					LocalPath:   "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		viper.Set("shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "x4SKYN5ZCRRYqx8s", B: Png()}, {Id: "3PYTWeCHuFj4Pbam", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeRecoveryErrorHttpCode", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "1",
						Message: "Image scheduled for processing.",
					},
					OriginalURL: "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					Key:         "file",
					LocalPath:   "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		viper.Set("shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "dZT3FzQBGf8Uxady", B: Png()}, {Id: "3Rp2BwHqTHzqbBzU", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeRecoveryInvalidJson", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			resp := `
				{
					"menu": {
						"id": "file",
						"popup": {
							"menuitem": [
								{
									"onclick": "CreateNewDoc()",
									"value": "New"
								},
								{
									"onclick": "OpenDoc()",
									"value": "Open"
								},
								{
									"onclick": "CloseDoc()",
									"value": "Close"
								}
							]
						},
						"value": "File"
					}
			`
			_, err := io.WriteString(w, resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "1",
						Message: "Image scheduled for processing.",
					},
					OriginalURL: "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					Key:         "file",
					LocalPath:   "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		viper.Set("shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "UGMhksnt86ZKzRSe", B: Png()}, {Id: "zMZTT2XDBQGpgwuc", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeRecoveryErrorStatusCode", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    -110,
						Message: "Upload error.(ERR_CODE: 4)",
					},
				},
			}
			err := json.NewEncoder(w).Encode(resp[0])
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "1",
						Message: "Image scheduled for processing.",
					},
					OriginalURL: "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					Key:         "file",
					LocalPath:   "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		viper.Set("shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "tmEDMuYysNR8FnDn", B: Png()}, {Id: "3mE2rkPk6h2NkqFr", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeRecoveryProcessingStatusCode", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "1",
						Message: "Image scheduled for processing.",
					},
					OriginalURL: "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					Key:         "file",
					LocalPath:   "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5.",
				},
			}
			err := json.NewEncoder(w).Encode(resp[0])
			if err != nil {
				t.Error(err)
			}
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		fn2 := func(w http.ResponseWriter, r *http.Request) {
			resp := response{
				{
					Status: struct {
						Code    interface{}
						Message string
					}{
						Code:    "1",
						Message: "Image scheduled for processing.",
					},
					OriginalURL: "http://api.shortpixel.com/u/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/30b9135837abeb3fc93a23a8db336cd5.",
					Key:         "file",
					LocalPath:   "/usr/local/ssd-drive/shortpixel/69PUiNOX9KapCxdbYXRvlJ0hGECybj3EzOvRruTtys/354/30b9135837abeb3fc93a23a8db336cd5.",
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			if err != nil {
				t.Error(err)
			}
		}
		mockserver2 := httptest.NewServer(http.HandlerFunc(fn2))
		defer mockserver2.Close()
		viper.Set("shortpixel.url", mockserver2.URL)
		viper.Set("shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		imgs := []model.Image{{Id: "Q8cVnqcJPu527Zu3", B: Png()}, {Id: "tWn9HyTguKvBnULD", B: Png()}}
		err := sp.Compress(context.Background(), imgs)
		if !errors.Is(err, model.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
}
