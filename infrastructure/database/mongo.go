package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func NewMongo() (mongo, error) {
	conf := newMongoConfig()
	ctx, _ := context.WithTimeout(context.Background(), conf.timeout)
	opts := options.Client().ApplyURI("mongodb://" + conf.host + ":" + conf.port)
	client, err := mongodb.Connect(ctx, opts)
	if err != nil {
		return mongo{}, errors.Wrap(err)
	}
	err = retry.Do(conf.times, conf.pause, func() error {
		ctx, _ := context.WithTimeout(context.Background(), conf.timeout)
		err := client.Ping(ctx, readpref.Primary())
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return mongo{}, errors.Wrap(err)
	}
	db := client.Database("aye-and-nay")
	images := db.Collection("images")
	edges := db.Collection("edges")
	return mongo{conf, images, edges}, nil
}

type mongo struct {
	conf   mongoConfig
	images *mongodb.Collection
	edges  *mongodb.Collection
}

func (m *mongo) SaveAlbum(ctx context.Context, alb model.Album) error {
	filter := bson.D{{"album", alb.Id}}
	n, err := m.images.CountDocuments(ctx, filter)
	if err != nil {
		return errors.Wrap(err)
	}
	if n > 0 {
		return errors.Wrap(model.ErrAblumAlreadyExists)
	}
	imgsDao := make([]interface{}, 0, len(alb.Images))
	for _, img := range alb.Images {
		compressed := false
		imgDao := imageDao{alb.Id, img.Id, img.Src, img.Rating, compressed}
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

func (m *mongo) CountImages(ctx context.Context, album string) (int, error) {
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

func (m *mongo) CountImagesCompressed(ctx context.Context, album string) (int, error) {
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

func (m *mongo) UpdateCompressionStatus(ctx context.Context, album string, image string) error {
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

func (m *mongo) GetImage(ctx context.Context, album string, image string) (model.Image, error) {
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

func (m *mongo) GetImages(ctx context.Context, album string) ([]string, error) {
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

func (m *mongo) SaveVote(ctx context.Context, album string, imageFrom string, imageTo string) error {
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
	opts := options.Update().SetUpsert(true)
	_, err = m.edges.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return errors.Wrap(err)
	}
	return nil
}

func (m *mongo) GetEdges(ctx context.Context, album string) (map[string]map[string]int, error) {
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

func (m *mongo) UpdateRatings(ctx context.Context, album string, vector map[string]float64) error {
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

func (m *mongo) GetImagesOrdered(ctx context.Context, album string) ([]model.Image, error) {
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
	opts := options.Find().SetSort(bson.D{{"rating", -1}})
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
