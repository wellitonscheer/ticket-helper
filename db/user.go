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

func save(userName *string) error {
	if userName == nil {
		return errors.New("invalid user name")
	}

	milvus, err := getMilvusInstance()
	if err != nil {
		return fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	collectionName := "user"

	collExists, err := milvus.c.HasCollection(milvus.ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection exists: %v", err.Error())
	}
	// if collExists {
	// 	_ = milvus.c.DropCollection(milvus.ctx, collectionName)
	// }
	if !collExists {
		schema := &entity.Schema{
			CollectionName: collectionName,
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
	}

	collections, err := milvus.c.ListCollections(milvus.ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %v", err.Error())
	}
	for _, collection := range collections {
		log.Printf("Collection id: %d, name: %s\n", collection.ID, collection.Name)
	}

	requestBody := []byte(fmt.Sprintf(`{"inputs": "["%s"]"}`, *userName))
	resp, err := http.Post("http://127.0.0.1:5000/embed", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("error to get user name embedding: %v", err.Error())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read embedded request body: %v", err.Error())
	}

	var embeddedUserName []float32
	if err := json.Unmarshal(body, &embeddedUserName); err != nil {
		return fmt.Errorf("failed to unmarshal embedded user name: %v", err.Error())
	}
	fmt.Printf("embedded response body: %v", embeddedUserName)

	// idColumn := entity.NewColumnInt64("id", nil)
	// userNameColumn := entity.NewColumnVarChar("userName", []string{*userName})
	// vectorColumn := entity.NewColumnFloatVector("vector", 1024, [][]float32{})

	return nil
}
