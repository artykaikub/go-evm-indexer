package repository

import (
	"context"
	"go-evm-indexer/models"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type ITransactionsRepository interface {
	FindTransactionsByBlockHash(ctx context.Context, blockHash common.Hash) ([]models.Transaction, error)
	FindTransactionByHash(ctx context.Context, hash common.Hash) (*models.Transaction, error)
	AddTransaction(ctx context.Context, tx *models.Transaction) error
	DeleteAllTransactionsByBlockHash(ctx context.Context, blockHash common.Hash) error
}

type TransactionsRepository struct {
	collection *mongo.Collection
}

func NewTransactionsRepository(db *mongo.Database) *TransactionsRepository {
	repo := &TransactionsRepository{
		collection: db.Collection("transactions"),
	}
	repo.createIndexes()

	return repo
}

func (t *TransactionsRepository) createIndexes() {
	models := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "hash", Value: bsonx.Int32(-1)}},
			Options: options.Index().SetUnique(true),
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := t.collection.Indexes().CreateMany(context.Background(), models, opts)
	if err != nil {
		log.Fatalf("❌ failed to create indexes of transactions repository : %s\n", err.Error())
	}
}

func (t *TransactionsRepository) FindTransactionsByBlockHash(ctx context.Context, blockHash common.Hash) ([]models.Transaction, error) {
	cursor, err := t.collection.Find(ctx, bson.M{
		"blockHash": blockHash.Hex(),
	}, nil)
	if err != nil {
		return nil, err
	}

	var out []models.Transaction
	err = cursor.All(ctx, &out)
	return out, err
}

func (t *TransactionsRepository) FindTransactionByHash(ctx context.Context, hash common.Hash) (*models.Transaction, error) {
	var out *models.Transaction
	if err := t.collection.FindOne(ctx, bson.M{
		"hash": hash.Hex(),
	}).Decode(&out); err != nil {
		return nil, err
	}

	return out, nil
}

func (t *TransactionsRepository) AddTransaction(ctx context.Context, tx *models.Transaction) error {
	payload, err := tx.MarshalBson()
	if err != nil {
		return err
	}

	_, err = t.collection.InsertOne(ctx, payload)
	if err != nil {
		return err
	}

	return nil
}

func (t *TransactionsRepository) DeleteAllTransactionsByBlockHash(ctx context.Context, blockHash common.Hash) error {
	_, err := t.collection.DeleteMany(ctx, bson.M{
		"blockHash": blockHash.Hex(),
	})

	return err
}
