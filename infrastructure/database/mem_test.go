package database

import (
	"context"
	"reflect"
	"sort"
	"testing"

	"github.com/zitryss/aye-and-nay/domain/model"
	_ "github.com/zitryss/aye-and-nay/internal/config"
	. "github.com/zitryss/aye-and-nay/internal/testing"
	"github.com/zitryss/aye-and-nay/pkg/errors"
)

func TestMemAlbum(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumEmptyFactory("EMqPQEyhp5cPTnaV")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		edgs, err := mem.GetEdges(context.Background(), "EMqPQEyhp5cPTnaV")
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(edgs, alb.Edges) {
			t.Error("edgs != alb.GetEdges")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumFullFactory("6FsUPNGm8XT89Vjg")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		alb = AlbumFullFactory("6FsUPNGm8XT89Vjg")
		err = mem.SaveAlbum(context.Background(), alb)
		if !errors.Is(err, model.ErrAlbumAlreadyExists) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem()
		_, err := mem.GetImages(context.Background(), "bZBnH7G6zFDZ9WHm")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		mem := NewMem()
		_, err := mem.GetEdges(context.Background(), "qbkzA2HqgELCxB5P")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestMemCount(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumEmptyFactory("c86jMVAX5Qgs2MZy")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		n, err := mem.CountImages(context.Background(), "c86jMVAX5Qgs2MZy")
		if err != nil {
			t.Error(err)
		}
		if n != 5 {
			t.Error("n != 5")
		}
		n, err = mem.CountImagesCompressed(context.Background(), "c86jMVAX5Qgs2MZy")
		if err != nil {
			t.Error(err)
		}
		if n != 0 {
			t.Error("n != 0")
		}
		err = mem.UpdateCompressionStatus(context.Background(), "c86jMVAX5Qgs2MZy", "RcBj3m9vuYPbntAE")
		if err != nil {
			t.Error(err)
		}
		n, err = mem.CountImages(context.Background(), "c86jMVAX5Qgs2MZy")
		if err != nil {
			t.Error(err)
		}
		if n != 5 {
			t.Error("n != 5")
		}
		n, err = mem.CountImagesCompressed(context.Background(), "c86jMVAX5Qgs2MZy")
		if err != nil {
			t.Error(err)
		}
		if n != 1 {
			t.Error("n != 1")
		}
		err = mem.UpdateCompressionStatus(context.Background(), "c86jMVAX5Qgs2MZy", "Q3NafBGuDH9PAtS4")
		if err != nil {
			t.Error(err)
		}
		n, err = mem.CountImagesCompressed(context.Background(), "c86jMVAX5Qgs2MZy")
		if err != nil {
			t.Error(err)
		}
		if n != 2 {
			t.Error("n != 2")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumEmptyFactory("x8nqgfCUVsFL985w")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		err = mem.UpdateCompressionStatus(context.Background(), "x8nqgfCUVsFL985w", "RcBj3m9vuYPbntAE")
		if err != nil {
			t.Error(err)
		}
		err = mem.UpdateCompressionStatus(context.Background(), "x8nqgfCUVsFL985w", "RcBj3m9vuYPbntAE")
		if err != nil {
			t.Error(err)
		}
		n, err := mem.CountImagesCompressed(context.Background(), "x8nqgfCUVsFL985w")
		if err != nil {
			t.Error(err)
		}
		if n != 1 {
			t.Error("n != 1")
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem()
		_, err := mem.CountImages(context.Background(), "WPbkn8VTVTPd5WYJ")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative3", func(t *testing.T) {
		mem := NewMem()
		_, err := mem.CountImagesCompressed(context.Background(), "nLYW4zNnH3tt639m")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative4", func(t *testing.T) {
		mem := NewMem()
		err := mem.UpdateCompressionStatus(context.Background(), "FLwXJhs4D2kkpehK", "RcBj3m9vuYPbntAE")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative5", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumEmptyFactory("2drK8rREqpFS2WYp")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		err = mem.UpdateCompressionStatus(context.Background(), "2drK8rREqpFS2WYp", "EC5md2qhemwAZmGf")
		if !errors.Is(err, model.ErrImageNotFound) {
			t.Error(err)
		}
	})
}

func TestMemImage(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumEmptyFactory("k9YA7PJmcMcdqEcR")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		img1, err := mem.GetImage(context.Background(), "k9YA7PJmcMcdqEcR", "VYFczQcF45x7gLYH")
		if err != nil {
			t.Error(err)
		}
		img2 := model.Image{Id: "VYFczQcF45x7gLYH", Src: "/aye-and-nay/albums/k9YA7PJmcMcdqEcR/images/428PcLG7e7VZHyAJ"}
		if !reflect.DeepEqual(img1, img2) {
			t.Error("img1 != img2")
		}
	})
	t.Run("Negative1", func(t *testing.T) {
		mem := NewMem()
		_, err := mem.GetImage(context.Background(), "8856LWPRnuSckPCa", "VYFczQcF45x7gLYH")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative2", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumEmptyFactory("g3VSAWnwX5fDkjcr")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		_, err = mem.GetImage(context.Background(), "g3VSAWnwX5fDkjcr", "W3rdTdrbRN3jedHB")
		if !errors.Is(err, model.ErrImageNotFound) {
			t.Error(err)
		}
	})
}

func TestMemVote(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumFullFactory("nAUeQgkR82njjGjB")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		imageFrom := "442BbctbQhcQHrgH"
		imageTo := "qBmu5KGTqCdvfgTU"
		err = mem.SaveVote(context.Background(), "nAUeQgkR82njjGjB", imageFrom, imageTo)
		if err != nil {
			t.Error(err)
		}
		err = mem.SaveVote(context.Background(), "nAUeQgkR82njjGjB", imageFrom, imageTo)
		if err != nil {
			t.Error(err)
		}
		edgs, err := mem.GetEdges(context.Background(), "nAUeQgkR82njjGjB")
		if err != nil {
			t.Error(err)
		}
		if edgs["442BbctbQhcQHrgH"]["qBmu5KGTqCdvfgTU"] != 2 {
			t.Error("edgs[imageFrom][imageTo] != 2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		mem := NewMem()
		imageFrom := "hQXK3DTRrQ8AHCcd"
		imageTo := "gukYVmHFmnB6fg7Q"
		err := mem.SaveVote(context.Background(), "Xuz8ZqVt8k3mAC6d", imageFrom, imageTo)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestMemSort(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumFullFactory("Xr5qXyfQAgnSNTzM")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		imgs1, err := mem.GetImagesOrdered(context.Background(), "Xr5qXyfQAgnSNTzM")
		if err != nil {
			t.Error(err)
		}
		img1 := model.Image{Id: "VYFczQcF45x7gLYH", Src: "/aye-and-nay/albums/Xr5qXyfQAgnSNTzM/images/428PcLG7e7VZHyAJ", Rating: 0.77920413}
		img2 := model.Image{Id: "RcBj3m9vuYPbntAE", Src: "/aye-and-nay/albums/Xr5qXyfQAgnSNTzM/images/6sgsr8WwqudTDzhR", Rating: 0.48954984}
		img3 := model.Image{Id: "442BbctbQhcQHrgH", Src: "/aye-and-nay/albums/Xr5qXyfQAgnSNTzM/images/kUrtHH5hTLbcSJdu", Rating: 0.41218211}
		img4 := model.Image{Id: "Q3NafBGuDH9PAtS4", Src: "/aye-and-nay/albums/Xr5qXyfQAgnSNTzM/images/2H7NpJkPwBWUk6gL", Rating: 0.19186324}
		img5 := model.Image{Id: "qBmu5KGTqCdvfgTU", Src: "/aye-and-nay/albums/Xr5qXyfQAgnSNTzM/images/gXR6VrL9h7E3pFVY", Rating: 0.13278389}
		imgs2 := []model.Image{img1, img2, img3, img4, img5}
		if !reflect.DeepEqual(imgs1, imgs2) {
			t.Error("imgs1 != imgs2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		mem := NewMem()
		_, err := mem.GetImagesOrdered(context.Background(), "M6cMTehk3LfV5CBy")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestMemRatings(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumFullFactory("Tz6NXWHXFzvWpumP")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		img1 := model.Image{Id: "RcBj3m9vuYPbntAE", Src: "/aye-and-nay/albums/Tz6NXWHXFzvWpumP/images/6sgsr8WwqudTDzhR", Rating: 0.54412788}
		img2 := model.Image{Id: "Q3NafBGuDH9PAtS4", Src: "/aye-and-nay/albums/Tz6NXWHXFzvWpumP/images/2H7NpJkPwBWUk6gL", Rating: 0.32537162}
		img3 := model.Image{Id: "442BbctbQhcQHrgH", Src: "/aye-and-nay/albums/Tz6NXWHXFzvWpumP/images/kUrtHH5hTLbcSJdu", Rating: 0.43185491}
		img4 := model.Image{Id: "VYFczQcF45x7gLYH", Src: "/aye-and-nay/albums/Tz6NXWHXFzvWpumP/images/428PcLG7e7VZHyAJ", Rating: 0.57356209}
		img5 := model.Image{Id: "qBmu5KGTqCdvfgTU", Src: "/aye-and-nay/albums/Tz6NXWHXFzvWpumP/images/gXR6VrL9h7E3pFVY", Rating: 0.61438023}
		imgs1 := []model.Image{img1, img2, img3, img4, img5}
		vector := map[string]float64{}
		vector[img1.Id] = img1.Rating
		vector[img2.Id] = img2.Rating
		vector[img3.Id] = img3.Rating
		vector[img4.Id] = img4.Rating
		vector[img5.Id] = img5.Rating
		err = mem.UpdateRatings(context.Background(), "Tz6NXWHXFzvWpumP", vector)
		if err != nil {
			t.Error(err)
		}
		imgs2, err := mem.GetImagesOrdered(context.Background(), "Tz6NXWHXFzvWpumP")
		if err != nil {
			t.Error(err)
		}
		sort.Slice(imgs1, func(i, j int) bool { return imgs1[i].Rating > imgs1[j].Rating })
		if !reflect.DeepEqual(imgs1, imgs2) {
			t.Error("imgs1 != imgs2")
		}
	})
	t.Run("Negative", func(t *testing.T) {
		mem := NewMem()
		img1 := model.Image{Id: "RcBj3m9vuYPbntAE", Src: "/aye-and-nay/albums/PB6wujzcRKjGKVzd/images/6sgsr8WwqudTDzhR", Rating: 0.54412788}
		img2 := model.Image{Id: "Q3NafBGuDH9PAtS4", Src: "/aye-and-nay/albums/PB6wujzcRKjGKVzd/images/2H7NpJkPwBWUk6gL", Rating: 0.32537162}
		img3 := model.Image{Id: "442BbctbQhcQHrgH", Src: "/aye-and-nay/albums/PB6wujzcRKjGKVzd/images/kUrtHH5hTLbcSJdu", Rating: 0.43185491}
		img4 := model.Image{Id: "VYFczQcF45x7gLYH", Src: "/aye-and-nay/albums/PB6wujzcRKjGKVzd/images/428PcLG7e7VZHyAJ", Rating: 0.57356209}
		img5 := model.Image{Id: "qBmu5KGTqCdvfgTU", Src: "/aye-and-nay/albums/PB6wujzcRKjGKVzd/images/gXR6VrL9h7E3pFVY", Rating: 0.61438023}
		vector := map[string]float64{}
		vector[img1.Id] = img1.Rating
		vector[img2.Id] = img2.Rating
		vector[img3.Id] = img3.Rating
		vector[img4.Id] = img4.Rating
		vector[img5.Id] = img5.Rating
		err := mem.UpdateRatings(context.Background(), "PB6wujzcRKjGKVzd", vector)
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}

func TestMemDelete(t *testing.T) {
	t.Run("Positive", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumEmptyFactory("CsRxWcm7bjhjCjPH")
		_, err := mem.CountImages(context.Background(), "CsRxWcm7bjhjCjPH")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
		err = mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		n, err := mem.CountImages(context.Background(), "CsRxWcm7bjhjCjPH")
		if err != nil {
			t.Error(err)
		}
		if n != 5 {
			t.Error("n != 5")
		}
		err = mem.DeleteAlbum(context.Background(), "CsRxWcm7bjhjCjPH")
		if err != nil {
			t.Error(err)
		}
		_, err = mem.CountImages(context.Background(), "CsRxWcm7bjhjCjPH")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
	t.Run("Negative", func(t *testing.T) {
		mem := NewMem()
		alb := AlbumEmptyFactory("pXHbPK8WuWC9x8cp")
		err := mem.SaveAlbum(context.Background(), alb)
		if err != nil {
			t.Error(err)
		}
		err = mem.DeleteAlbum(context.Background(), "9JFs2DWEDmZWXSyy")
		if !errors.Is(err, model.ErrAlbumNotFound) {
			t.Error(err)
		}
	})
}
