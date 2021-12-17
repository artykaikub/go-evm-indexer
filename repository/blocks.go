package repository

import (
	"context"
	"errors"
	"go-evm-indexer/models"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type IBlocksRepository interface {
	FindLastestBlock(ctx context.Context) (*models.Block, error)
	FindBlockByHash(ctx context.Context, hash common.Hash) (*models.Block, error)
	FindBlockByNumber(ctx context.Context, number uint64) (*models.Block, error)
	FindBlockByRange(ctx context.Context, from, to uint64) ([]models.Block, error)
	FindIncompleteBlock(ctx context.Context) ([]models.Block, error)
	AddBlock(ctx context.Context, block *models.Block) error
	DeleteAllIncompleteBlocks(ctx context.Context) error
	UpdateToDone(ctx context.Context, number uint64) (*models.Block, error)
	CountBlocks(ctx context.Context) (uint64, error)
}

type BlocksRepository struct {
	collection *mongo.Collection
}

func NewBlocksRepository(db *mongo.Database) *BlocksRepository {
	repo := &BlocksRepository{
		collection: db.Collection("blocks"),
	}
	repo.createIndexes()

	return repo
}

func (b *BlocksRepository) createIndexes() {
	models := []mongo.IndexModel{
		{
			Keys:    bsonx.Doc{{Key: "hash", Value: bsonx.Int32(-1)}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bsonx.Doc{{Key: "number", Value: bsonx.Int32(-1)}},
			Options: options.Index().SetUnique(true),
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := b.collection.Indexes().CreateMany(context.Background(), models, opts)
	if err != nil {
		log.Fatalf("❌ failed to create indexes of blocks repository : %s\n", err.Error())
	}
}

func (b *BlocksRepository) FindLastestBlock(ctx context.Context) (*models.Block, error) {
	opts := options.FindOne()
	opts.SetSort(bson.M{
		"number": -1,
	})

	var out *models.Block
	if err := b.collection.FindOne(ctx, bson.M{}, opts).Decode(&out); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return out, nil
}

func (b *BlocksRepository) FindBlockByNumber(ctx context.Context, number uint64) (*models.Block, error) {
	var out *models.Block
	if err := b.collection.FindOne(ctx, bson.M{
		"number": number,
	}).Decode(&out); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return out, nil
}

func (b *BlocksRepository) FindBlockByHash(ctx context.Context, hash common.Hash) (*models.Block, error) {
	var out *models.Block
	if err := b.collection.FindOne(ctx, bson.M{
		"hash": hash.Hex(),
	}).Decode(&out); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	return out, nil
}

func (b *BlocksRepository) FindBlockByRange(ctx context.Context, from, to uint64) ([]models.Block, error) {
	opts := options.Find()
	opts.SetSort(bson.M{
		"number": 1,
	})

	cursor, err := b.collection.Find(ctx, bson.M{
		"number": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}, opts)
	if err != nil {
		return nil, err
	}

	var out []models.Block
	err = cursor.All(ctx, &out)
	return out, err
}

func (b *BlocksRepository) FindIncompleteBlock(ctx context.Context) ([]models.Block, error) {
	cursor, err := b.collection.Find(ctx, bson.M{
		"isDone": false,
	}, nil)
	if err != nil {
		return nil, err
	}

	var out []models.Block
	err = cursor.All(ctx, &out)
	return out, err
}

func (b *BlocksRepository) AddBlock(ctx context.Context, block *models.Block) error {
	payload, err := block.MarshalBson()
	if err != nil {
		return err
	}

	_, err = b.collection.InsertOne(ctx, payload)
	if err != nil {
		return err
	}

	return nil
}

func (b *BlocksRepository) DeleteAllIncompleteBlocks(ctx context.Context) error {
	_, err := b.collection.DeleteMany(ctx, bson.M{
		"isDone": false,
	})
	return err
}

func (b *BlocksRepository) UpdateToDone(ctx context.Context, number uint64) (*models.Block, error) {
	otps := options.FindOneAndUpdate()
	otps.SetReturnDocument(options.After)

	var out *models.Block
	err := b.collection.FindOneAndUpdate(ctx, bson.M{
		"number": number,
	}, bson.M{
		"$set": bson.M{
			"isDone": true,
		},
	}, otps).Decode(&out)

	return out, err
}

func (b *BlocksRepository) CountBlocks(ctx context.Context) (uint64, error) {
	count, err := b.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return 0, nil
		}
		return 0, err
	}

	return uint64(count), nil
}
