package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/wellitonscheer/ticket-helper/internal/config"
)

type MilvusClient struct {
	Client client.Client
	Ctx    context.Context
	Cancel context.CancelFunc
}

func NewMilvusConnection(conf *config.Config) *MilvusClient {
	fmt.Println("Connecting to milvus now.")

	milvusAddr := fmt.Sprintf("%s:%s", conf.Common.BaseUrl, conf.Milvus.MilvulPort)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	c, err := client.NewClient(ctx, client.Config{
		Address:        milvusAddr,
		RetryRateLimit: &client.RetryRateLimitOption{MaxRetry: 3, MaxBackoff: time.Second * 2},
	})
	if err != nil {
		cancel()
		panic(err)
	}

	return &MilvusClient{c, ctx, cancel}
}

func TestDb() error {
	milvus := getMilvusInstance()

	collectionName := `gosdk_basic_collection`

	collExists, err := (*milvus).client.HasCollection(milvus.ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection exists: %v", err.Error())
	}
	if collExists {
		err = milvus.client.DropCollection(milvus.ctx, collectionName)
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
	err = milvus.client.CreateCollection(milvus.ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err.Error())
	}

	collections, err := milvus.client.ListCollections(milvus.ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %v", err.Error())
	}
	for _, collection := range collections {
		log.Printf("Collection id: %d, name: %s\n", collection.ID, collection.Name)
	}

	partitions, err := milvus.client.ShowPartitions(milvus.ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to show partitions: %v", err.Error())
	}
	for _, partition := range partitions {
		log.Printf("partition id: %d, name: %s\n", partition.ID, partition.Name)
	}

	partitionName := "new_partition"
	err = milvus.client.CreatePartition(milvus.ctx, collectionName, partitionName)
	if err != nil {
		return fmt.Errorf("failed to create partition: %v", err.Error())
	}

	log.Println("After create partition")
	partitions, err = milvus.client.ShowPartitions(milvus.ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to show partitions: %v", err.Error())
	}
	for _, partition := range partitions {
		log.Printf("partition id: %d, name: %s\n", partition.ID, partition.Name)
	}

	_ = milvus.client.DropCollection(milvus.ctx, collectionName)
	milvus.client.Close()

	return nil
}
