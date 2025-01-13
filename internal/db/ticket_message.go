package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/wellitonscheer/ticket-helper/internal/service"
)

type TicketMessageData struct {
	Type    string `json:"type"`
	Ordem   int32  `json:"ordem"`
	Subject string `json:"subject"`
	Poster  string `json:"poster"`
	Body    string `json:"body"`
}

type TicketRawData map[string][]TicketMessageData

type TicketMessage struct {
	Milvus         *MilvusClient
	CollectionName string
}

func NewTicketMessage() (*TicketMessage, error) {
	milvus, err := getMilvusInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	return &TicketMessage{
		Milvus:         milvus,
		CollectionName: "ticket_message",
	}, nil
}

func (tm *TicketMessage) InsertAllTickets() error {
	collExists, err := tm.Milvus.c.HasCollection(tm.Milvus.ctx, tm.CollectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %v", err.Error())
	}

	if !collExists {
		err = tm.CreateCollection()
		if err != nil {
			return fmt.Errorf("failed to create collection: %v", err.Error())
		}
	}

	rawData, err := os.ReadFile("./ai/data/outputs/id_list.jsona")
	if err != nil {
		return fmt.Errorf("failed to read from json file: %v", err.Error())
	}

	var jsonData TicketRawData
	err = json.Unmarshal(rawData, &jsonData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal raw ticket data: %v", err.Error())
	}

	ticketsIds := []string{}
	ticketMsgTypes := []string{}
	ticketMsgPosters := []string{}
	ticketMsgContents := []string{}
	ticketMsgContentVector := [][]float32{}
	for ticketId, ticketMessages := range jsonData {
		fmt.Println("processing: ", ticketId)
		for i, message := range ticketMessages {
			if len(message.Body) > 65534 || len(message.Body) < 1 {
				fmt.Println("ignored: ", ticketId)
				continue
			}

			data := service.Input{
				Inputs: []string{message.Body},
			}
			embeddedBodyMessage, err := service.GetTextEmbeddings(&data)
			if err != nil {
				return fmt.Errorf("failed to get embeddings: %v", err.Error())
			}

			if len(embeddedBodyMessage) > 1 {
				fmt.Printf("embedded body message returned two vectors, ticketId: %s index: %d", ticketId, i)
				fmt.Println("embedded body message: ", embeddedBodyMessage)
				break
			}

			ticketsIds = append(ticketsIds, ticketId)
			ticketMsgTypes = append(ticketMsgTypes, message.Type)
			ticketMsgPosters = append(ticketMsgPosters, message.Poster)
			ticketMsgContents = append(ticketMsgContents, message.Body)
			ticketMsgContentVector = append(ticketMsgContentVector, embeddedBodyMessage...)
		}
		fmt.Println("finished: ", ticketId)
	}
	fmt.Println("REALY INSERTING NOW!")

	batchSize := 1000
	for i := 0; i < len(ticketsIds); i += batchSize {
		end := i + batchSize
		if end > len(ticketsIds) {
			end = len(ticketsIds)
		}
		fmt.Println(i, " to ", end)

		ticketIdBatch := ticketsIds[i:end]
		ticketMsgTypesBatch := ticketMsgTypes[i:end]
		ticketMsgPostersBatch := ticketMsgPosters[i:end]
		ticketMsgContentBatch := ticketMsgContents[i:end]
		ticketMsgContentVecBatch := ticketMsgContentVector[i:end]

		ticketIdColumn := entity.NewColumnVarChar("ticketId", ticketIdBatch)
		ticketMsgTypeColumn := entity.NewColumnVarChar("type", ticketMsgTypesBatch)
		ticketMsgPosterColumn := entity.NewColumnVarChar("poster", ticketMsgPostersBatch)
		ticketMsgContentColumn := entity.NewColumnVarChar("ticketMessageContent", ticketMsgContentBatch)
		ticketMsgContentVecColumn := entity.NewColumnFloatVector("ticketMessageContentVector", 1024, ticketMsgContentVecBatch)

		_, err = tm.Milvus.c.Insert(tm.Milvus.ctx, tm.CollectionName, "", ticketIdColumn, ticketMsgTypeColumn, ticketMsgPosterColumn, ticketMsgContentColumn, ticketMsgContentVecColumn)
		if err != nil {
			return fmt.Errorf("failed to insert tickets: %v", err.Error())
		}
	}

	fmt.Println("DONE inserting...")

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

func (tm *TicketMessage) VectorSearch(search *string) (TicketSearchResults, error) {
	if search == nil {
		return nil, errors.New("invalid search value")
	}

	hasColl, err := tm.Milvus.c.HasCollection(tm.Milvus.ctx, tm.CollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if has collection")
	}
	if !hasColl {
		return nil, fmt.Errorf("'%s' collection doesnt exist", tm.CollectionName)
	}

	embedInput := service.Input{
		Inputs: []string{*search},
	}
	searchEmbedding, err := service.GetTextEmbeddings(&embedInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get search embeddings: %v", err.Error())
	}

	vector := entity.FloatVector(searchEmbedding[0])
	sp, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return nil, fmt.Errorf("failed to create new index flat search param: %v", err.Error())
	}
	searchResults, err := tm.Milvus.c.Search(tm.Milvus.ctx, tm.CollectionName, nil, "", []string{"ticketId"}, []entity.Vector{vector}, "ticketMessageContentVector", entity.COSINE, 20, sp)
	if err != nil {
		return nil, fmt.Errorf("failed to search ticket: %v", err.Error())
	}

	var ticketsResult TicketSearchResults
	for _, result := range searchResults {
		for _, field := range result.Fields {
			var ticketIdColumn *entity.ColumnVarChar
			if field.Name() == "ticketId" {
				c, ok := field.(*entity.ColumnVarChar)
				if ok {
					ticketIdColumn = c
				}
			}

			if ticketIdColumn == nil {
				return nil, errors.New("result field not match")
			}

			ticketsIds := ticketIdColumn.Data()

			for i, score := range result.Scores {
				ticketsResult = append(ticketsResult, TicketSearchResult{TicketId: ticketsIds[i], Score: score})
			}
		}
	}

	return ticketsResult, nil
}

func (tm *TicketMessage) CreateCollection() error {
	schema := entity.Schema{
		CollectionName: tm.CollectionName,
		Description:    "Tickets Messages",
		AutoID:         true,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     "ticketId",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "20",
				},
				PrimaryKey: false,
				AutoID:     false,
			},
			{
				Name:     "type",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "20",
				},
				PrimaryKey: false,
				AutoID:     false,
			},
			{
				Name:     "poster",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "1000",
				},
				PrimaryKey: false,
				AutoID:     false,
			},
			{
				Name:     "ticketMessageContent",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: "65534",
				},
				PrimaryKey: false,
				AutoID:     false,
			},
			{
				Name:     "ticketMessageContentVector",
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

	err = tm.Milvus.c.CreateIndex(tm.Milvus.ctx, tm.CollectionName, "ticketMessageContentVector", idx, false)
	if err != nil {
		return fmt.Errorf("fail to create index: %v", err.Error())
	}

	err = tm.Milvus.c.LoadCollection(tm.Milvus.ctx, tm.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}
