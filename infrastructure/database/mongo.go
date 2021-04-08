package database

import (
	"context"

	lru "github.com/hashicorp/golang-lru"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	optionsdb "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

type albumLru map[uint64]string

type imageDao struct {
	Album      int64
	Id         int64
	Src        string
	Rating     float64
	Compressed bool
}

type edgeDao struct {
	Album  int64
	From   int64
	To     int64
	Weight int
}

func NewMongo() (*Mongo, error) {
	conf := newMongoConfig()
	ctx, cancel := context.WithTimeout(context.Background(), conf.timeout)
	defer cancel()
	opts := optionsdb.Client().ApplyURI("mongodb://" + conf.host + ":" + conf.port)
	client, err := mongodb.Connect(ctx, opts)
	if err != nil {
		return &Mongo{}, errors.Wrap(err)
	}
	err = retry.Do(conf.times, conf.pause, func() error {
		err := client.Ping(ctx, readpref.Primary())
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return &Mongo{}, errors.Wrap(err)
	}
	db := client.Database("aye-and-nay")
	images := db.Collection("images")
	edges := db.Collection("edges")
	cache, err := lru.New(conf.lru)
	if err != nil {
		return &Mongo{}, errors.Wrap(err)
	}
	return &Mongo{conf, client, db, images, edges, cache}, nil
}

type Mongo struct {
	conf   mongoConfig
	client *mongodb.Client
	db     *mongodb.Database
	images *mongodb.Collection
	edges  *mongodb.Collection
	cache  *lru.Cache
}

func (m *Mongo) SaveAlbum(ctx context.Context, alb model.Album) error {
	filter := bson.D{{"album", int64(alb.Id)}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n > 0 {
		return errors.Wrap(model.ErrAlbumAlreadyExists)
	}
	imgsDao := make([]interface{}, 0, len(alb.Images))
	albLru := make(albumLru, len(alb.Images))
	for _, img := range alb.Images {
		imgDao := imageDao{int64(alb.Id), int64(img.Id), img.Src, img.Rating, m.conf.compressed}
		imgsDao = append(imgsDao, imgDao)
		albLru[img.Id] = img.Src
	}
	_, err = m.images.InsertMany(ctx, imgsDao)
	if err != nil {
		return errors.Wrap(err)
	}
	for from, v := range alb.Edges {
		for to, rating := range v {
			edgDao := edgeDao{int64(alb.Id), int64(from), int64(to), rating}
			_, err = m.edges.InsertOne(ctx, edgDao)
			if err != nil {
				return errors.Wrap(err)
			}
		}
	}
	m.cache.Add(alb.Id, albLru)
	return nil
}

func (m *Mongo) CountImages(ctx context.Context, album uint64) (int, error) {
	albLru, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return len(albLru), nil
}

func (m *Mongo) CountImagesCompressed(ctx context.Context, album uint64) (int, error) {
	_, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	filter := bson.D{{"album", int64(album)}, {"compressed", true}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return int(n), nil
}

func (m *Mongo) UpdateCompressionStatus(ctx context.Context, album uint64, image uint64) error {
	albLru, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return errors.Wrap(err)
	}
	_, ok := albLru[image]
	if !ok {
		return errors.Wrap(model.ErrImageNotFound)
	}
	filter := bson.D{{"album", int64(album)}, {"id", int64(image)}}
	update := bson.D{{"$set", bson.D{{"compressed", true}}}}
	_, err = m.images.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (m *Mongo) GetImageSrc(ctx context.Context, album uint64, image uint64) (string, error) {
	albLru, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return "", errors.Wrap(err)
	}
	src, ok := albLru[image]
	if !ok {
		return "", errors.Wrap(model.ErrImageNotFound)
	}
	return src, nil
}

func (m *Mongo) GetImagesIds(ctx context.Context, album uint64) ([]uint64, error) {
	albLru, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	images := make([]uint64, 0, len(albLru))
	for image := range albLru {
		images = append(images, image)
	}
	return images, nil
}

func (m *Mongo) SaveVote(ctx context.Context, album uint64, imageFrom uint64, imageTo uint64) error {
	_, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return errors.Wrap(err)
	}
	filter := bson.D{{"album", int64(album)}, {"from", int64(imageFrom)}, {"to", int64(imageTo)}}
	update := bson.D{{"$inc", bson.D{{"weight", 1}}}}
	opts := optionsdb.Update().SetUpsert(true)
	_, err = m.edges.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (m *Mongo) GetEdges(ctx context.Context, album uint64) (map[uint64]map[uint64]int, error) {
	albLru, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	edgs := make(map[uint64]map[uint64]int, len(albLru))
	filter := bson.D{{"album", int64(album)}}
	cursor, err := m.images.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		imgDao := imageDao{}
		err := cursor.Decode(&imgDao)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		edgs[uint64(imgDao.Id)] = make(map[uint64]int, len(albLru))
		filter := bson.D{{"album", int64(album)}, {"from", imgDao.Id}}
		cursor, err := m.edges.Find(ctx, filter)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		for cursor.Next(ctx) {
			edgDao := edgeDao{}
			err := cursor.Decode(&edgDao)
			if err != nil {
				_ = cursor.Close(ctx)
				return nil, errors.Wrap(err)
			}
			edgs[uint64(edgDao.From)][uint64(edgDao.To)] = edgDao.Weight
		}
		err = cursor.Err()
		if err != nil {
			_ = cursor.Close(ctx)
			return nil, errors.Wrap(err)
		}
		err = cursor.Close(ctx)
		if err != nil {
			return nil, errors.Wrap(err)
		}
	}
	err = cursor.Err()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	err = cursor.Close(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return edgs, nil
}

func (m *Mongo) UpdateRatings(ctx context.Context, album uint64, vector map[uint64]float64) error {
	_, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return errors.Wrap(err)
	}
	for id, rating := range vector {
		filter := bson.D{{"album", int64(album)}, {"id", int64(id)}}
		update := bson.D{{"$set", bson.D{{"rating", rating}}}}
		_, err := m.images.UpdateOne(ctx, filter, update)
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

func (m *Mongo) GetImagesOrdered(ctx context.Context, album uint64) ([]model.Image, error) {
	albLru, err := m.lruGetOrAddAndGet(ctx, album)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	imgs := make([]model.Image, 0, len(albLru))
	filter := bson.D{{"album", int64(album)}}
	opts := optionsdb.Find().SetSort(bson.D{{"rating", -1}})
	cursor, err := m.images.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		imgDao := imageDao{}
		err := cursor.Decode(&imgDao)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		img := model.Image{Id: uint64(imgDao.Id), Src: imgDao.Src, Rating: imgDao.Rating}
		imgs = append(imgs, img)
	}
	err = cursor.Err()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	err = cursor.Close(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return imgs, nil
}

func (m *Mongo) DeleteAlbum(ctx context.Context, album uint64) error {
	filter := bson.D{{"album", int64(album)}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 0 {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	_, err = m.images.DeleteMany(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	_, err = m.edges.DeleteMany(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	m.cache.Remove(album)
	return nil
}

func (m *Mongo) lruGetOrAddAndGet(ctx context.Context, album uint64) (albumLru, error) {
	a, ok := m.cache.Get(album)
	if !ok {
		err := m.lruAdd(ctx, album)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		a, ok = m.cache.Get(album)
		if !ok {
			return nil, errors.Wrap(model.ErrUnknown)
		}
	}
	return a.(albumLru), nil
}

func (m *Mongo) lruAdd(ctx context.Context, album uint64) error {
	filter := bson.D{{"album", int64(album)}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 0 {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	albLru := make(albumLru, n)
	filter = bson.D{{"album", int64(album)}}
	cursor, err := m.images.Find(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		imgDao := imageDao{}
		err := cursor.Decode(&imgDao)
		if err != nil {
			return errors.Wrap(err)
		}
		albLru[uint64(imgDao.Id)] = imgDao.Src
	}
	err = cursor.Err()
	if err != nil {
		return errors.Wrap(err)
	}
	err = cursor.Close(ctx)
	if err != nil {
		return errors.Wrap(err)
	}
	m.cache.Add(album, albLru)
	return nil
}

func (m *Mongo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), m.conf.timeout)
	defer cancel()
	err := m.client.Disconnect(ctx)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
