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

	"github.com/zitryss/aye-and-nay/domain/domain"
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
			_, err := io.Copy(w, Png())
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("Positive2", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			_, err := io.Copy(w, Png())
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("NegativeInvalidUrl1", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		mockserver1.Close()
		viper.Set("compressor.shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeTimeout1", func(t *testing.T) {
		d := viper.GetDuration("compressor.shortpixel.uploadTimeout") + viper.GetDuration("compressor.shortpixel.downloadTimeout")
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(2 * d)
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		viper.Set("compressor.shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeTimeout2", func(t *testing.T) {
		d := viper.GetDuration("compressor.shortpixel.uploadTimeout") + viper.GetDuration("compressor.shortpixel.downloadTimeout")
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeErrorHttpCode1", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}
		mockserver1 := httptest.NewServer(http.HandlerFunc(fn1))
		defer mockserver1.Close()
		viper.Set("compressor.shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrNotImage) {
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
		viper.Set("compressor.shortpixel.url", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrNotImage) {
			t.Error(err)
		}
	})
	t.Run("PositiveRecovery1", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			_, err := io.Copy(w, Png())
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
		viper.Set("compressor.shortpixel.url", mockserver3.URL)
		viper.Set("compressor.shortpixel.url2", mockserver2.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("PositiveRecovery2", func(t *testing.T) {
		fn1 := func(w http.ResponseWriter, r *http.Request) {
			_, err := io.Copy(w, Png())
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
		viper.Set("compressor.shortpixel.url", mockserver3.URL)
		viper.Set("compressor.shortpixel.url2", mockserver2.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		viper.Set("compressor.shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
	t.Run("NegativeRecoveryTimeout", func(t *testing.T) {
		d := viper.GetDuration("compressor.shortpixel.uploadTimeout") + viper.GetDuration("compressor.shortpixel.downloadTimeout")
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		viper.Set("compressor.shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		viper.Set("compressor.shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		viper.Set("compressor.shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		viper.Set("compressor.shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
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
		viper.Set("compressor.shortpixel.url", mockserver2.URL)
		viper.Set("compressor.shortpixel.url2", mockserver1.URL)
		sp := NewShortPixel()
		_, err := sp.Compress(context.Background(), Png())
		if !errors.Is(err, domain.ErrThirdPartyUnavailable) {
			t.Error(err)
		}
	})
}
