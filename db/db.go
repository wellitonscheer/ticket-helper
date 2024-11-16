package db

import (
	"context"
	"fmt"
	"log"
	"sync"

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

func getMilvusInstance() *MilvusClient {
	if milvusInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if milvusInstance == nil {
			fmt.Println("Creating single instance now.")

			milvusAddr := `localhost:19530`

			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)

			c, err := client.NewClient(ctx, client.Config{
				Address: milvusAddr,
			})
			if err != nil {
				log.Fatal("failed to connect to milvus:", err.Error())
			}

			milvusInstance = &MilvusClient{c, ctx, cancel}
		} else {
			fmt.Println("Single instance already created.")
		}
	} else {
		fmt.Println("Single instance already created.")
	}

	return milvusInstance
}

func TestDb() {
	milvus := getMilvusInstance()

	collectionName := `gosdk_basic_collection`

	// first, lets check the collection exists
	collExists, err := (*milvus).c.HasCollection(milvus.ctx, collectionName)
	if err != nil {
		log.Fatal("failed to check collection exists:", err.Error())
	}
	if collExists {
		// let's say the example collection is only for sampling the API
		// drop old one in case early crash or something
		_ = milvus.c.DropCollection(milvus.ctx, collectionName)
	}

	// define collection schema
	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:    "this is the basic example collection",
		AutoID:         true,
		Fields: []*entity.Field{
			// currently primary key field is compulsory, and only int64 is allowd
			{
				Name:       "int64",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			// also the vector field is needed
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{ // the vector dim may changed def method in release
					entity.TypeParamDim: "128",
				},
			},
		},
	}
	err = milvus.c.CreateCollection(milvus.ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		log.Fatal("failed to create collection:", err.Error())
	}

	collections, err := milvus.c.ListCollections(milvus.ctx)
	if err != nil {
		log.Fatal("failed to list collections:", err.Error())
	}
	for _, collection := range collections {
		// print all the collections, id & name
		log.Printf("Collection id: %d, name: %s\n", collection.ID, collection.Name)
	}

	// show collection partitions
	partitions, err := milvus.c.ShowPartitions(milvus.ctx, collectionName)
	if err != nil {
		log.Fatal("failed to show partitions:", err.Error())
	}
	for _, partition := range partitions {
		// print partition info, the shall be a default partition for out collection
		log.Printf("partition id: %d, name: %s\n", partition.ID, partition.Name)
	}

	partitionName := "new_partition"
	// now let's try to create a partition
	err = milvus.c.CreatePartition(milvus.ctx, collectionName, partitionName)
	if err != nil {
		log.Fatal("failed to create partition:", err.Error())
	}

	log.Println("After create partition")
	// show collection partitions, check creation
	partitions, err = milvus.c.ShowPartitions(milvus.ctx, collectionName)
	if err != nil {
		log.Fatal("failed to show partitions:", err.Error())
	}
	for _, partition := range partitions {
		log.Printf("partition id: %d, name: %s\n", partition.ID, partition.Name)
	}

	// clean up our mess
	_ = milvus.c.DropCollection(milvus.ctx, collectionName)
	milvus.c.Close()
}
