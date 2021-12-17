package repository

import (
	"context"
	"go-evm-indexer/models"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type IEventsRepository interface {
	FindEventsByBlockHash(ctx context.Context, blockHash common.Hash) ([]models.Event, error)
	AddEvent(ctx context.Context, event *models.Event) error
	DeleteAllEventsByBlockHash(ctx context.Context, blockHash common.Hash) error
}

type EventsRepository struct {
	collection *mongo.Collection
}

func NewEventsRepository(db *mongo.Database) *EventsRepository {
	repo := &EventsRepository{
		collection: db.Collection("events"),
	}

	return repo
}

func (e *EventsRepository) FindEventsByBlockHash(ctx context.Context, blockHash common.Hash) ([]models.Event, error) {
	cursor, err := e.collection.Find(ctx, bson.M{
		"blockHash": blockHash.Hex(),
	}, nil)
	if err != nil {
		return nil, err
	}

	var out []models.Event
	err = cursor.All(ctx, &out)
	return out, err
}

func (e *EventsRepository) AddEvent(ctx context.Context, event *models.Event) error {
	payload, err := event.MarshalBson()
	if err != nil {
		return err
	}

	_, err = e.collection.InsertOne(ctx, payload)
	if err != nil {
		return err
	}

	return nil
}

func (e *EventsRepository) DeleteAllEventsByBlockHash(ctx context.Context, blockHash common.Hash) error {
	_, err := e.collection.DeleteMany(ctx, bson.M{
		"blockHash": blockHash.Hex(),
	})

	return err
}
