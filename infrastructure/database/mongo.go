package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	optionsdb "go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/zitryss/aye-and-nay/domain/model"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/retry"
)

type imageDao struct {
	Album      string
	Id         string
	Src        string
	Rating     float64
	Compressed bool
}

type edgeDao struct {
	Album  string
	From   string
	To     string
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
	return &Mongo{conf, images, edges}, nil
}

type Mongo struct {
	conf   mongoConfig
	images *mongodb.Collection
	edges  *mongodb.Collection
}

func (m *Mongo) SaveAlbum(ctx context.Context, alb model.Album) error {
	filter := bson.D{{"album", alb.Id}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n > 0 {
		return errors.Wrap(model.ErrAlbumAlreadyExists)
	}
	imgsDao := make([]interface{}, 0, len(alb.Images))
	for _, img := range alb.Images {
		imgDao := imageDao{alb.Id, img.Id, img.Src, img.Rating, m.conf.compressed}
		imgsDao = append(imgsDao, imgDao)
	}
	_, err = m.images.InsertMany(ctx, imgsDao)
	if err != nil {
		return errors.Wrap(err)
	}
	for from, v := range alb.Edges {
		for to, rating := range v {
			edgDao := edgeDao{alb.Id, from, to, rating}
			_, err = m.edges.InsertOne(ctx, edgDao)
			if err != nil {
				return errors.Wrap(err)
			}
		}
	}
	return nil
}

func (m *Mongo) CountImages(ctx context.Context, album uint64) (int, error) {
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	if n == 0 {
		return 0, errors.Wrap(model.ErrAlbumNotFound)
	}
	return int(n), nil
}

func (m *Mongo) CountImagesCompressed(ctx context.Context, album uint64) (int, error) {
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	if n == 0 {
		return 0, errors.Wrap(model.ErrAlbumNotFound)
	}
	filter = bson.D{{"album", album}, {"compressed", true}}
	n, err = m.images.CountDocuments(ctx, filter)
	if err != nil {
		return 0, errors.Wrap(err)
	}
	return int(n), nil
}

func (m *Mongo) UpdateCompressionStatus(ctx context.Context, album uint64, image uint64) error {
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 0 {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	filter = bson.D{{"album", album}, {"id", image}}
	n, err = m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 0 {
		return errors.Wrap(model.ErrImageNotFound)
	}
	filter = bson.D{{"album", album}, {"id", image}}
	update := bson.D{{"$set", bson.D{{"compressed", true}}}}
	_, err = m.images.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (m *Mongo) GetImage(ctx context.Context, album uint64, image uint64) (model.Image, error) {
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return model.Image{}, errors.Wrap(err)
	}
	if n == 0 {
		return model.Image{}, errors.Wrap(model.ErrAlbumNotFound)
	}
	filter = bson.D{{"album", album}, {"id", image}}
	n, err = m.images.CountDocuments(ctx, filter)
	if err != nil {
		return model.Image{}, errors.Wrap(err)
	}
	if n == 0 {
		return model.Image{}, errors.Wrap(model.ErrImageNotFound)
	}
	filter = bson.D{{"album", album}, {"id", image}}
	img := model.Image{}
	err = m.images.FindOne(ctx, filter).Decode(&img)
	if err != nil {
		return model.Image{}, errors.Wrap(err)
	}
	return img, nil
}

func (m *Mongo) GetImages(ctx context.Context, album uint64) ([]uint64, error) {
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if n == 0 {
		return nil, errors.Wrap(model.ErrAlbumNotFound)
	}
	images := make([]string, 0, n)
	filter = bson.D{{"album", album}}
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
		images = append(images, imgDao.Id)
	}
	err = cursor.Err()
	if err != nil {
		return nil, errors.Wrap(err)
	}
	err = cursor.Close(ctx)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	return images, nil
}

func (m *Mongo) SaveVote(ctx context.Context, album uint64, imageFrom uint64, imageTo uint64) error {
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 0 {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	filter = bson.D{{"album", album}, {"from", imageFrom}, {"to", imageTo}}
	update := bson.D{{"$inc", bson.D{{"weight", 1}}}}
	opts := optionsdb.Update().SetUpsert(true)
	_, err = m.edges.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (m *Mongo) GetEdges(ctx context.Context, album uint64) (map[uint64]map[uint64]int, error) {
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if n == 0 {
		return nil, errors.Wrap(model.ErrAlbumNotFound)
	}
	edgs := make(map[string]map[string]int, n)
	filter = bson.D{{"album", album}}
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
		edgs[imgDao.Id] = make(map[string]int, n)
		filter := bson.D{{"album", album}, {"from", imgDao.Id}}
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
			edgs[edgDao.From][edgDao.To] = edgDao.Weight
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
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n == 0 {
		return errors.Wrap(model.ErrAlbumNotFound)
	}
	for id, rating := range vector {
		filter := bson.D{{"album", album}, {"id", id}}
		update := bson.D{{"$set", bson.D{{"rating", rating}}}}
		_, err := m.images.UpdateOne(ctx, filter, update)
		if err != nil {
			return errors.Wrap(err)
		}
	}
	return nil
}

func (m *Mongo) GetImagesOrdered(ctx context.Context, album uint64) ([]model.Image, error) {
	filter := bson.D{{"album", album}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err)
	}
	if n == 0 {
		return nil, errors.Wrap(model.ErrAlbumNotFound)
	}
	imgs := make([]model.Image, 0, n)
	filter = bson.D{{"album", album}}
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
		img := model.Image{Id: imgDao.Id, Src: imgDao.Src, Rating: imgDao.Rating}
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
	filter := bson.D{{"album", album}}
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
	return nil
}
