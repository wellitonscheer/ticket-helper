package db

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/wellitonscheer/ticket-helper/internal/service"
)

type BlackTicket struct {
	Milvus         *MilvusClient
	CollectionName string
}

func NewBlackTicket() (*BlackTicket, error) {
	milvus, err := getMilvusInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	return &BlackTicket{
		Milvus:         milvus,
		CollectionName: "black_ticket",
	}, nil
}

func (tm *BlackTicket) InsertAllTickets() error {
	collExists, err := tm.Milvus.c.HasCollection(tm.Milvus.ctx, tm.CollectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %v", err.Error())
	}

	if collExists {
		err = tm.Milvus.c.DropCollection(tm.Milvus.ctx, tm.CollectionName)
		if err != nil {
			return fmt.Errorf("failed to drop collection: %v", err.Error())
		}
	}

	err = tm.CreateCollection()
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err.Error())
	}

	rawData, err := os.ReadFile("./data_source/black_tickets_content.jsona")
	if err != nil {
		return fmt.Errorf("failed to read from json file: %v", err.Error())
	}

	var blackTicketsContent []string
	err = json.Unmarshal(rawData, &blackTicketsContent)
	if err != nil {
		return fmt.Errorf("failed to unmarshal black tickets content: %v", err.Error())
	}

	ticketContents := []string{}
	ticketContentVector := [][]float32{}
	for _, ticketContent := range blackTicketsContent {
		if len(ticketContent) > 65534 || len(ticketContent) < 1 {
			continue
		}

		data := service.Input{
			Inputs: []string{ticketContent},
		}
		embeddedContent, err := service.GetTextEmbeddings(&data)
		if err != nil {
			return fmt.Errorf("failed to get embeddings: %v", err.Error())
		}

		if len(embeddedContent) > 1 {
			fmt.Printf("embedded black ticket content returned two vectors, content: %s", ticketContent)
			fmt.Println("embedded ticket content: ", embeddedContent)
			break
		}

		ticketContents = append(ticketContents, ticketContent)
		ticketContentVector = append(ticketContentVector, embeddedContent...)
	}

	ticketContentColumn := entity.NewColumnVarChar("ticketContent", ticketContents)
	ticketContentVecColumn := entity.NewColumnFloatVector("ticketContentVector", 1024, ticketContentVector)

	_, err = tm.Milvus.c.Insert(tm.Milvus.ctx, tm.CollectionName, "", ticketContentColumn, ticketContentVecColumn)
	if err != nil {
		return fmt.Errorf("failed to insert tickets: %v", err.Error())
	}

	err = tm.Milvus.c.Flush(tm.Milvus.ctx, tm.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush collection: %v", err.Error())
	}

	err = tm.Milvus.c.LoadCollection(tm.Milvus.ctx, tm.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}

func (tm *BlackTicket) CreateCollection() error {
	schema := entity.Schema{
		CollectionName: tm.CollectionName,
		Description:    "Black Tickets",
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     "ticketContent",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "65534",
				},
				PrimaryKey: false,
				AutoID:     false,
			},
			{
				Name:     "ticketContentVector",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: "1024",
				},
				PrimaryKey: false,
				AutoID:     false,
			},
		},
	}

	err := tm.Milvus.c.CreateCollection(tm.Milvus.ctx, &schema, 1)
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err.Error())
	}

	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 2)
	if err != nil {
		return fmt.Errorf("fail to create ivf flat index: %v", err.Error())
	}

	err = tm.Milvus.c.CreateIndex(tm.Milvus.ctx, tm.CollectionName, "ticketContentVector", idx, false)
	if err != nil {
		return fmt.Errorf("fail to create index: %v", err.Error())
	}

	err = tm.Milvus.c.LoadCollection(tm.Milvus.ctx, tm.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}
