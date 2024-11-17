package db

import (
	"log"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type UserService struct {
	Save func()
}

var User = &UserService{save}

func save() {
	milvus := getMilvusInstance()

	collectionName := "user"

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
					entity.TypeParamDim: "1024",
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
}
