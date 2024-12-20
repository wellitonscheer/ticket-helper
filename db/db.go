package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

var lock = &sync.Mutex{}

type MilvusClient struct {
	c      client.Client
	ctx    context.Context
	cancel context.CancelFunc
}

var milvusInstance *MilvusClient

func getMilvusInstance() (*MilvusClient, error) {
	if milvusInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if milvusInstance == nil {
			fmt.Println("Creating single instance now.")
			err := godotenv.Load()
			if err != nil {
				return nil, fmt.Errorf("error loading .env file: %v", err.Error())
			}

			baseURL := os.Getenv("BASE_URL")
			milvusPort := os.Getenv("MILVUS_PORT")

			milvusAddr := fmt.Sprintf("%s:%s", baseURL, milvusPort)

			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)

			c, err := client.NewClient(ctx, client.Config{
				Address:        milvusAddr,
				RetryRateLimit: &client.RetryRateLimitOption{MaxRetry: 3, MaxBackoff: time.Second * 2},
			})
			if err != nil {
				cancel()
				return nil, fmt.Errorf("failed to connect to milvus: %v", err.Error())
			}

			milvusInstance = &MilvusClient{c, ctx, cancel}
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}

	return milvusInstance, nil
}

func TestDb() error {
	milvus, err := getMilvusInstance()
	if err != nil {
		return fmt.Errorf("failed connecting to milvus: %v", err.Error())
	}

	collectionName := `gosdk_basic_collection`

	collExists, err := (*milvus).c.HasCollection(milvus.ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection exists: %v", err.Error())
	}
	if collExists {
		err = milvus.c.DropCollection(milvus.ctx, collectionName)
		return fmt.Errorf("failed to drop collection: %v", err.Error())
	}

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "this is the basic example collection",
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       "int64",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: "128",
				},
			},
		},
	}
	err = milvus.c.CreateCollection(milvus.ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err.Error())
	}

	collections, err := milvus.c.ListCollections(milvus.ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %v", err.Error())
	}
	for _, collection := range collections {
		log.Printf("Collection id: %d, name: %s\n", collection.ID, collection.Name)
	}

	partitions, err := milvus.c.ShowPartitions(milvus.ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to show partitions: %v", err.Error())
	}
	for _, partition := range partitions {
		log.Printf("partition id: %d, name: %s\n", partition.ID, partition.Name)
	}

	partitionName := "new_partition"
	err = milvus.c.CreatePartition(milvus.ctx, collectionName, partitionName)
	if err != nil {
		return fmt.Errorf("failed to create partition: %v", err.Error())
	}

	log.Println("After create partition")
	partitions, err = milvus.c.ShowPartitions(milvus.ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to show partitions: %v", err.Error())
	}
	for _, partition := range partitions {
		log.Printf("partition id: %d, name: %s\n", partition.ID, partition.Name)
	}

	_ = milvus.c.DropCollection(milvus.ctx, collectionName)
	milvus.c.Close()

	return nil
}
