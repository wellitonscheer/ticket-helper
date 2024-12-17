package db

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type UserService struct {
	Save func(userName *string) error
}

var User = &UserService{save}

var userCollName string = "user"

func save(userName *string) error {
	if userName == nil {
		return errors.New("invalid user name")
	}

	milvus, err := getMilvusInstance()
	if err != nil {
		return fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	collExists, err := milvus.c.HasCollection(milvus.ctx, userCollName)
	if err != nil {
		return fmt.Errorf("failed to check collection exists: %v", err.Error())
	}
	// if collExists {
	// 	_ = milvus.c.DropCollection(milvus.ctx, userCollName)
	// }
	if !collExists {
		err = createUserCollection()
		if err != nil {
			return fmt.Errorf("failed to create collection '%s': %v", userCollName, err.Error())
		}
	}

	collections, err := milvus.c.ListCollections(milvus.ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %v", err.Error())
	}
	for _, collection := range collections {
		log.Printf("Collection id: %d, name: %s\n", collection.ID, collection.Name)
	}

	data := map[string][]string{
		"inputs": {*userName},
	}
	requestBody, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to create request body: %v", err.Error())
	}

	resp, err := http.Post("http://127.0.0.1:5000/embed", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error to get user name embedding: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read embedded request body: %v", err.Error())
	}

	var embeddedUserName [][]float32
	if err := json.Unmarshal(body, &embeddedUserName); err != nil {
		return fmt.Errorf("failed to unmarshal embedded user name: %w, Response body: %s", err, string(body))
	}

	userNameColumn := entity.NewColumnVarChar("userName", []string{*userName})
	vectorColumn := entity.NewColumnFloatVector("vector", 1024, embeddedUserName)

	_, err = milvus.c.Insert(milvus.ctx, userCollName, "", userNameColumn, vectorColumn)
	if err != nil {
		return fmt.Errorf("failed to insert user: %v", err.Error())
	}

	err = milvus.c.Flush(milvus.ctx, userCollName, false)
	if err != nil {
		return fmt.Errorf("failed to flush collection: %v", err.Error())
	}

	return nil
}

// func search() error {
// 	milvus, err := getMilvusInstance()
// 	if err != nil {
// 		return fmt.Errorf("failed to connect to milvus: %v", err.Error())
// 	}

// 	aaa := entity.SearchParam{}
// 	milvus.c.Search(milvus.ctx, userCollName, nil, nil, []string{"userName"}, make([][]float32, 0, 1024), "vector", entity.COSINE, nil)

// 	return nil
// }

func createUserCollection() error {
	milvus, err := getMilvusInstance()
	if err != nil {
		return fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	schema := &entity.Schema{
		CollectionName: userCollName,
		Description:    "usuario",
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     "userName",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "200",
				},
				PrimaryKey: false,
				AutoID:     false,
			},
			{
				Name:     "vector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: "1024",
				},
			},
		},
	}
	err = milvus.c.CreateCollection(milvus.ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err.Error())
	}

	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 2)
	if err != nil {
		return fmt.Errorf("fail to create ivf flat index: %v", err.Error())
	}

	err = milvus.c.CreateIndex(milvus.ctx, userCollName, "vector", idx, false)
	if err != nil {
		return fmt.Errorf("fail to create index: %v", err.Error())
	}

	err = milvus.c.LoadCollection(milvus.ctx, userCollName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}
