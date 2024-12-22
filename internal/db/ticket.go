package db

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/wellitonscheer/ticket-helper/internal/service"
)

type TicketService struct {
	InsertAllTickets func() error
}

var Ticket = &TicketService{
	insertAllTickets,
}

var ticketCollName string = "ticket"

type TicketMessage struct {
	Type  string `json:"type"`
	Ordem int32  `json:"ordem"`
	Body  string `json:"body"`
}

type TicketRawData map[string][]TicketMessage

func insertAllTickets() error {
	milvus, err := getMilvusInstance()
	if err != nil {
		return fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	collExists, err := milvus.c.HasCollection(milvus.ctx, ticketCollName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %v", err.Error())
	}

	if !collExists {
		err = createTicketCollection()
		if err != nil {
			return fmt.Errorf("failed to create collection: %v", err.Error())
		}
	}

	rawData, err := os.ReadFile("./ai/data/outputs/id_list.json")
	if err != nil {
		return fmt.Errorf("failed to read from json file: %v", err.Error())
	}

	var jsonData TicketRawData
	err = json.Unmarshal(rawData, &jsonData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal raw ticket data: %v", err.Error())
	}

	ticketsIds := []string{}
	ticketContents := []string{}
	ticketContentVector := [][]float32{}
	for ticketId, ticketMessages := range jsonData {
		fmt.Println("processing: ", ticketId)
		fullBodyMessage := ""
		for _, message := range ticketMessages {
			fullBodyMessage = fmt.Sprintf("%s | %s", fullBodyMessage, message.Body)
		}

		if len(fullBodyMessage) > 65534 {
			continue
		}

		data := service.Input{
			Inputs: []string{fullBodyMessage},
		}
		embeddedBodyMessage, err := service.GetTextEmbeddings(&data)
		if err != nil {
			return fmt.Errorf("failed to get fullBodyMessage embeddings: %v", err.Error())
		}

		if len(embeddedBodyMessage) > 1 {
			fmt.Println("embedded body message returned two vectors, ticketId: ", ticketId)
			fmt.Println("embedded body message: ", embeddedBodyMessage)
			fmt.Println("fullBodyMessage legth used in the embedded: ", len(fullBodyMessage))
			break
		}

		ticketsIds = append(ticketsIds, ticketId)
		ticketContents = append(ticketContents, fullBodyMessage)
		ticketContentVector = append(ticketContentVector, embeddedBodyMessage...)
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
		ticketContentBatch := ticketContents[i:end]
		ticketContentVecBatch := ticketContentVector[i:end]

		ticketIdColumn := entity.NewColumnVarChar("ticketId", ticketIdBatch)
		ticketContentColumn := entity.NewColumnVarChar("ticketContent", ticketContentBatch)
		ticketContentVecColumn := entity.NewColumnFloatVector("ticketContentVector", 1024, ticketContentVecBatch)

		_, err = milvus.c.Insert(milvus.ctx, ticketCollName, "", ticketIdColumn, ticketContentColumn, ticketContentVecColumn)
		if err != nil {
			return fmt.Errorf("failed to insert tickets: %v", err.Error())
		}
	}

	fmt.Println("DONE inserting...")

	err = milvus.c.Flush(milvus.ctx, ticketCollName, false)
	if err != nil {
		return fmt.Errorf("failed to flush collection: %v", err.Error())
	}

	err = milvus.c.LoadCollection(milvus.ctx, ticketCollName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}

func createTicketCollection() error {
	milvus, err := getMilvusInstance()
	if err != nil {
		return fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	schema := entity.Schema{
		CollectionName: ticketCollName,
		Description:    "Tickets",
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

	err = milvus.c.CreateCollection(milvus.ctx, &schema, 1)
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err.Error())
	}

	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 2)
	if err != nil {
		return fmt.Errorf("fail to create ivf flat index: %v", err.Error())
	}

	err = milvus.c.CreateIndex(milvus.ctx, ticketCollName, "ticketContentVector", idx, false)
	if err != nil {
		return fmt.Errorf("fail to create index: %v", err.Error())
	}

	err = milvus.c.LoadCollection(milvus.ctx, ticketCollName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}
