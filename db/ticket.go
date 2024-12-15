package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type TicketService struct {
	InsertAllTickets func() (string, error)
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

func insertAllTickets() (string, error) {
	milvus, err := getMilvusInstance()
	if err != nil {
		return "", fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	collExists, err := milvus.c.HasCollection(milvus.ctx, ticketCollName)
	if err != nil {
		return "", fmt.Errorf("failed to check if collection exists: %v", err.Error())
	}

	if !collExists {
		err = createTicketCollection()
		if err != nil {
			return "", fmt.Errorf("failed to create collection: %v", err.Error())
		}
	}

	rawData, err := os.ReadFile("./ai/data/outputs/id_list.json")
	if err != nil {
		return "", fmt.Errorf("failed to read from json file: %v", err.Error())
	}

	var jsonData TicketRawData
	err = json.Unmarshal(rawData, &jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal raw ticket data: %v", err.Error())
	}

	maxBodyLength := 0
	ticketMax := ""
	ticketsAmount := 0
	ticketsIds := []string{}
	ticketContents := []string{}
	ticketContentVector := [][]float32{}
	for ticketId, ticketMessages := range jsonData {
		fmt.Println("processing: ", ticketId)
		fullBodyMassage := ""
		for _, message := range ticketMessages {
			fullBodyMassage = fmt.Sprintf("%s | %s", fullBodyMassage, message.Body)
		}

		if len(fullBodyMassage) > 65534 {
			continue
		}

		data := map[string][]string{
			"inputs": {fullBodyMassage},
		}
		requestBody, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to create request body: %v", err.Error())
		}

		resp, err := http.Post("http://127.0.0.1:5000/embed", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			return "", fmt.Errorf("error to get body message embedding: %v", err.Error())
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read embedded request body: %v", err.Error())
		}

		var embeddedBodyMessage [][]float32
		if err := json.Unmarshal(body, &embeddedBodyMessage); err != nil {
			return "", fmt.Errorf("failed to unmarshal embedded body message: %w, Response body: %s", err, string(body))
		}
		if len(embeddedBodyMessage) > 1 {
			fmt.Println(ticketId)
			fmt.Println(embeddedBodyMessage)
			fmt.Println(len(fullBodyMassage))
			break
		}

		ticketsIds = append(ticketsIds, ticketId)
		ticketContents = append(ticketContents, fullBodyMassage)
		ticketContentVector = append(ticketContentVector, embeddedBodyMessage...)

		if len(fullBodyMassage) > maxBodyLength && len(fullBodyMassage) < 29164 {
			maxBodyLength = len(fullBodyMassage)
			ticketMax = ticketId
			ticketsAmount++
		}
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
			return "", fmt.Errorf("failed to insert tickets: %v", err.Error())
		}
	}

	// ticketIdColumn := entity.NewColumnVarChar("ticketId", ticketsIds)
	// ticketContentColumn := entity.NewColumnVarChar("ticketContent", ticketContents)
	// ticketContentVecColumn := entity.NewColumnFloatVector("ticketContentVector", 1024, ticketContentVector)

	// _, err = milvus.c.Insert(milvus.ctx, ticketCollName, "", ticketIdColumn, ticketContentColumn, ticketContentVecColumn)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to insert tickets: %v", err.Error())
	// }
	fmt.Println("DONE inserting...")

	err = milvus.c.Flush(milvus.ctx, ticketCollName, false)
	if err != nil {
		return "", fmt.Errorf("failed to flush collection: %v", err.Error())
	}

	err = milvus.c.LoadCollection(milvus.ctx, ticketCollName, false)
	if err != nil {
		return "", fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return fmt.Sprintf("ticketAmount: %s, ticket: %s, length: %s", fmt.Sprint(ticketsAmount), ticketMax, fmt.Sprint(maxBodyLength)), nil
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
