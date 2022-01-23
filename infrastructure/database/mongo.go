package database

import (
	"context"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	optionsdb "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/zitryss/aye-and-nay/domain/domain"
	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

var (
	_ domain.Databaser = (*Mongo)(nil)
)

type albumLru map[uint64]string

type imageDao struct {
	Album      int64
	Id         int64
	Src        string
	Rating     float64
	Compressed bool
	Expires    time.Time
}

type edgeDao struct {
	Album  int64
	From   int64
	To     int64
	Weight int
}

func NewMongo(ctx context.Context, conf MongoConfig) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(ctx, conf.Timeout)
	defer cancel()
	opts := optionsdb.Client().ApplyURI("mongodb://" + conf.Host + ":" + conf.Port)
	client, err := mongodb.Connect(ctx, opts)
	if err != nil {
		return &Mongo{}, errors.Wrap(err)
	}
	db := client.Database("aye-and-nay")
	images := db.Collection("images")
	edges := db.Collection("edges")
	cache, err := lru.New(conf.LRU)
	if err != nil {
		return &Mongo{}, errors.Wrap(err)
	}
	m := &Mongo{conf, client, db, images, edges, cache}
	err = retry.Do(conf.RetryTimes, conf.RetryPause, func() error {
		_, err := m.Health(ctx)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return &Mongo{}, errors.Wrap(err)
	}
	return m, nil
}

type Mongo struct {
	conf   MongoConfig
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
		return errors.Wrap(domain.ErrAlbumAlreadyExists)
	}
	imgsDao := make([]interface{}, 0, len(alb.Images))
	albLru := make(albumLru, len(alb.Images))
	for _, img := range alb.Images {
		imgDao := imageDao{int64(alb.Id), int64(img.Id), img.Src, img.Rating, m.conf.Compressed, alb.Expires}
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
		return errors.Wrap(domain.ErrImageNotFound)
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
		return "", errors.Wrap(domain.ErrImageNotFound)
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
	filter := bson.D{{"album", int64(album)}}
	cursor, err := m.images.Find(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	imgsDao := make([]imageDao, 0, len(albLru))
	err = cursor.All(ctx, &imgsDao)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	edgs := make(map[uint64]map[uint64]int, len(albLru))
	for _, imgDao := range imgsDao {
		edgs[uint64(imgDao.Id)] = make(map[uint64]int, len(albLru))
		filter := bson.D{{"album", int64(album)}, {"from", imgDao.Id}}
		cursor, err := m.edges.Find(ctx, filter)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		edgsDao := []edgeDao(nil)
		err = cursor.All(ctx, &edgsDao)
		if err != nil {
			return nil, errors.Wrap(err)
		}
		for _, edgDao := range edgsDao {
			edgs[uint64(edgDao.From)][uint64(edgDao.To)] = edgDao.Weight
		}
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
	filter := bson.D{{"album", int64(album)}}
	opts := optionsdb.Find().SetSort(bson.D{{"rating", -1}})
	cursor, err := m.images.Find(ctx, filter, opts)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	imgsDao := make([]imageDao, 0, len(albLru))
	err = cursor.All(ctx, &imgsDao)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	imgs := make([]model.Image, 0, len(albLru))
	for _, imgDao := range imgsDao {
		img := model.Image{Id: uint64(imgDao.Id), Src: imgDao.Src, Rating: imgDao.Rating}
		imgs = append(imgs, img)
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
		return errors.Wrap(domain.ErrAlbumNotFound)
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

func (m *Mongo) AlbumsToBeDeleted(ctx context.Context) ([]model.Album, error) {
	match := bson.D{{"$match", bson.D{{"expires", bson.D{{"$ne", time.Time{}}}}}}}
	group := bson.D{{"$group", bson.D{{"_id", "$album"}, {"expires", bson.D{{"$first", "$expires"}}}}}}
	cursor, err := m.images.Aggregate(ctx, mongodb.Pipeline{match, group})
	if err != nil {
		return nil, errors.Wrap(err)
	}
	results := []bson.M(nil)
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	albs := make([]model.Album, 0, len(results))
	for _, r := range results {
		albs = append(albs, model.Album{Id: uint64(r["_id"].(int64)), Expires: r["expires"].(primitive.DateTime).Time()})
	}
	return albs, nil
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
			return nil, errors.Wrap(domain.ErrUnknown)
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
		return errors.Wrap(domain.ErrAlbumNotFound)
	}
	filter = bson.D{{"album", int64(album)}}
	cursor, err := m.images.Find(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	imgsDao := []imageDao(nil)
	err = cursor.All(ctx, &imgsDao)
	if err != nil {
		return errors.Wrap(err)
	}
	albLru := make(albumLru, n)
	for _, imgDao := range imgsDao {
		albLru[uint64(imgDao.Id)] = imgDao.Src
	}
	m.cache.Add(album, albLru)
	return nil
}

func (m *Mongo) Health(ctx context.Context) (bool, error) {
	err := m.client.Ping(ctx, readpref.Primary())
	if err != nil {
		return false, errors.Wrapf(domain.ErrBadHealthDatabase, "%s", err)
	}
	return true, err
}

func (m *Mongo) Close(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, m.conf.Timeout)
	defer cancel()
	err := m.client.Disconnect(ctx)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}
