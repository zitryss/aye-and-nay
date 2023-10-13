package database

import (
	"context"
	"encoding/binary"
	"encoding/gob"
	"runtime"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/exp/slices"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/pool"
)

var (
	_ domain.Databaser = (*Badger)(nil)
)

func NewBadger(conf BadgerConfig) (*Badger, error) {
	_ = runtime.GOMAXPROCS(128)
	path := "./badger"
	if conf.InMemory {
		path = ""
	}
	opts := badger.DefaultOptions(path).WithCompression(options.None).WithLogger(nil).WithInMemory(conf.InMemory)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	cache, err := lru.New(conf.LRU)
	if err != nil {
		return &Badger{}, errors.Wrap(err)
	}
	return &Badger{conf, db, cache}, nil
}

type Badger struct {
	conf  BadgerConfig
	db    *badger.DB
	cache *lru.Cache
}

func (b *Badger) Monitor(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			_ = b.db.RunValueLogGC(b.conf.GcRatio)
			time.Sleep(b.conf.CleanupInterval)
		}
	}()
}

func (b *Badger) SaveAlbum(_ context.Context, alb model.Album) error {
	_, err := b.get(alb.Id)
	if err == nil {
		return errors.Wrap(domain.ErrAlbumAlreadyExists)
	}
	edgs := make(map[uint64]map[uint64]int, len(alb.Images))
	albLru := make(albumLru, len(alb.Images))
	for i := range alb.Images {
		img := &alb.Images[i]
		img.Compressed = b.conf.Compressed
		edgs[img.Id] = make(map[uint64]int, len(alb.Images))
		albLru[img.Id] = img.Src
	}
	alb.Edges = edgs
	err = b.set(alb)
	if err != nil {
		return errors.Wrap(err)
	}
	b.cache.Add(alb.Id, albLru)
	return nil
}

func (b *Badger) CountImages(_ context.Context, album uint64) (int, error) {
	albLru, err := b.lruGetOrAddAndGet(album)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return len(albLru), nil
}

func (b *Badger) CountImagesCompressed(_ context.Context, album uint64) (int, error) {
	alb, err := b.get(album)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return 0, errors.Wrap(domain.ErrAlbumNotFound)
	}
	n := 0
	for _, img := range alb.Images {
		if img.Compressed {
			n++
		}
	}
	return n, nil
}

func (b *Badger) UpdateCompressionStatus(_ context.Context, album uint64, image uint64) error {
	alb, err := b.get(album)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return errors.Wrap(domain.ErrAlbumNotFound)
	}
	found := false
	for i := range alb.Images {
		img := &alb.Images[i]
		if img.Id == image {
			img.Compressed = true
			found = true
			break
		}
	}
	if !found {
		return errors.Wrap(domain.ErrImageNotFound)
	}
	err = b.set(alb)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (b *Badger) GetImageSrc(_ context.Context, album uint64, image uint64) (string, error) {
	albLru, err := b.lruGetOrAddAndGet(album)
	if err != nil {
		return "", errors.Wrap(err)
	}
	src, ok := albLru[image]
	if !ok {
		return "", errors.Wrap(domain.ErrImageNotFound)
	}
	return src, nil
}

func (b *Badger) GetImagesIds(_ context.Context, album uint64) ([]uint64, error) {
	albLru, err := b.lruGetOrAddAndGet(album)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	images := make([]uint64, 0, len(albLru))
	for image := range albLru {
		images = append(images, image)
	}
	return images, nil
}

func (b *Badger) SaveVote(_ context.Context, album uint64, imageFrom uint64, imageTo uint64) error {
	alb, err := b.get(album)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return errors.Wrap(domain.ErrAlbumNotFound)
	}
	alb.Edges[imageFrom][imageTo]++
	err = b.set(alb)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (b *Badger) GetEdges(_ context.Context, album uint64) (map[uint64]map[uint64]int, error) {
	alb, err := b.get(album)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, errors.Wrap(domain.ErrAlbumNotFound)
	}
	return alb.Edges, nil
}

func (b *Badger) UpdateRatings(_ context.Context, album uint64, vector map[uint64]float64) error {
	alb, err := b.get(album)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return errors.Wrap(domain.ErrAlbumNotFound)
	}
	for id, rating := range vector {
		for i := range alb.Images {
			img := &alb.Images[i]
			if img.Id == id {
				img.Rating = rating
				break
			}
		}
	}
	err = b.set(alb)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (b *Badger) GetImagesOrdered(_ context.Context, album uint64) ([]model.Image, error) {
	alb, err := b.get(album)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, errors.Wrap(domain.ErrAlbumNotFound)
	}
	slices.SortFunc(alb.Images, func(a, b model.Image) bool { return a.Rating > b.Rating })
	return alb.Images, nil
}

func (b *Badger) DeleteAlbum(_ context.Context, album uint64) error {
	_, err := b.get(album)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return errors.Wrap(domain.ErrAlbumNotFound)
	}
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, album)
	err = b.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete(key)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err)
	}
	b.cache.Remove(album)
	return nil
}

func (b *Badger) AlbumsToBeDeleted(_ context.Context) ([]model.Album, error) {
	keys, err := b.keys()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	albs := []model.Album(nil)
	for _, key := range keys {
		alb, err := b.get(key)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		if !alb.Expires.IsZero() {
			albs = append(albs, alb)
		}
	}
	return albs, nil
}

func (b *Badger) get(album uint64) (model.Album, error) {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, album)
	buf := pool.GetBuffer()
	defer pool.PutBuffer(buf)
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return errors.Wrap(err)
		}
		err = item.Value(func(val []byte) error {
			_, err = buf.Write(val)
			if err != nil {
				return errors.Wrap(err)
			}
			return nil
		})
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return model.Album{}, errors.Wrap(err)
	}
	alb := model.Album{}
	err = gob.NewDecoder(buf).Decode(&alb)
	if err != nil {
		return model.Album{}, errors.Wrap(err)
	}
	return alb, nil
}

func (b *Badger) set(alb model.Album) error {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, alb.Id)
	buf := pool.GetBuffer()
	defer pool.PutBuffer(buf)
	err := gob.NewEncoder(buf).Encode(alb)
	if err != nil {
		return errors.Wrap(err)
	}
	err = b.db.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, buf.Bytes())
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (b *Badger) keys() ([]uint64, error) {
	keys := []uint64(nil)
	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			key := binary.LittleEndian.Uint64(k)
			keys = append(keys, key)
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return keys, nil
}

func (b *Badger) lruGetOrAddAndGet(album uint64) (albumLru, error) {
	a, ok := b.cache.Get(album)
	if !ok {
		err := b.lruAdd(album)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		a, ok = b.cache.Get(album)
		if !ok {
			return nil, errors.Wrap(domain.ErrUnknown)
		}
	}
	return a.(albumLru), nil
}

func (b *Badger) lruAdd(album uint64) error {
	alb, err := b.get(album)
	if errors.Is(err, badger.ErrKeyNotFound) {
		return errors.Wrap(domain.ErrAlbumNotFound)
	}
	albLru := make(albumLru, len(alb.Images))
	for _, img := range alb.Images {
		albLru[img.Id] = img.Src
	}
	b.cache.Add(album, albLru)
	return nil
}

func (b *Badger) Health(_ context.Context) (bool, error) {
	return true, nil
}

func (b *Badger) Close(_ context.Context) error {
	err := b.db.Close()
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (b *Badger) Reset() error {
	err := b.db.DropAll()
	if err != nil {
		return errors.Wrap(err)
	}
	b.cache.Purge()
	return nil
}
