package mongo

import (
	"errors"
	"fmt"

	"github.com/diabolusgx/guess-the-number/internal/config"
	"github.com/diabolusgx/guess-the-number/pkg/logger"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type BaseClient interface {
	GetCollection(collectionName string) *mongo.Collection
	StartSession(opts ...options.Lister[options.SessionOptions]) (*mongo.Session, error)
}

type Client struct {
	client *mongo.Client
	config *config.Configuration
}

func NewClient(config *config.Configuration, logger *logger.Logger) (BaseClient, error) {
	uri := config.Mongo.GetConnectionURI()
	if uri == "" {
		logger.Fatal("empty mongo connection uri")
		return nil, errors.New("empty mongo connection uri")
	}

	// Configure BSON options to handle ObjectIDs as hex strings
	bsonOpts := &options.BSONOptions{
		ObjectIDAsHexString: true,
	}

	clientOptions := options.Client().
		SetMaxPoolSize(config.Mongo.MaxPoolSize).
		SetMinPoolSize(config.Mongo.MinPoolSize).
		ApplyURI(uri).
		SetReadPreference(readpref.SecondaryPreferred()).
		SetBSONOptions(bsonOpts)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		logger.Fatalf("failed to connect to mongo: %w", err)
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	logger.Debugw("connected to mongodb",
		"host", config.Mongo.Host,
		"db", config.Mongo.Database,
	)

	return &Client{
		client: client,
		config: config,
	}, nil
}

func (c *Client) GetCollection(collectionName string) *mongo.Collection {
	return c.client.Database(c.config.Mongo.Database).Collection(collectionName)
}

func (c *Client) StartSession(opts ...options.Lister[options.SessionOptions]) (*mongo.Session, error) {
	return c.client.StartSession(opts...)
}
