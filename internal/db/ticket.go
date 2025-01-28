package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/wellitonscheer/ticket-helper/internal/service"
)

type TicketSearchResult struct {
	TicketId string
	Score    float32
}

type TicketSearchResults = []TicketSearchResult

type TicketService struct {
	Milvus         *MilvusClient
	CollectionName string
}

func NewTicketService() (*TicketService, error) {
	milvus, err := getMilvusInstance()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to milvus: %v", err.Error())
	}

	return &TicketService{
		Milvus:         milvus,
		CollectionName: "ticket",
	}, nil
}

func (t *TicketService) InsertAllTickets() error {
	collExists, err := t.Milvus.c.HasCollection(t.Milvus.ctx, t.CollectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %v", err.Error())
	}

	if !collExists {
		err = t.CreateTicketCollection()
		if err != nil {
			return fmt.Errorf("failed to create collection: %v", err.Error())
		}
	}

	rawData, err := os.ReadFile("./ai/data/outputs/isd_list.json")
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
		for i, message := range ticketMessages {
			tipo := ""
			switch message.Type {
			case "M":
				tipo = "mensagem"
			case "R":
				tipo = "resposta"
			case "N":
				tipo = "nota interna"
				fmt.Println(tipo)
			}
			if i == 0 {
				fullBodyMessage += fmt.Sprintf("%s %s", message.Subject, message.Body)
				continue
			}
			fullBodyMessage = fmt.Sprintf("%s %s", fullBodyMessage, message.Body)
		}

		if len(fullBodyMessage) > 65534 || len(fullBodyMessage) < 5 {
			fmt.Println("ignored: ", ticketId)
			continue
		}

		isBlackListed, err := t.IsBlackListed(&fullBodyMessage)
		if err != nil {
			return fmt.Errorf("failed to check if content is black listed: %v", err.Error())
		}

		if isBlackListed {
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
		ticketContentBatch := ticketContents[i:end]
		ticketContentVecBatch := ticketContentVector[i:end]

		ticketIdColumn := entity.NewColumnVarChar("ticketId", ticketIdBatch)
		ticketContentColumn := entity.NewColumnVarChar("ticketContent", ticketContentBatch)
		ticketContentVecColumn := entity.NewColumnFloatVector("ticketContentVector", 1024, ticketContentVecBatch)

		_, err = t.Milvus.c.Insert(t.Milvus.ctx, t.CollectionName, "", ticketIdColumn, ticketContentColumn, ticketContentVecColumn)
		if err != nil {
			return fmt.Errorf("failed to insert tickets: %v", err.Error())
		}
	}

	fmt.Println("DONE inserting...")

	err = t.Milvus.c.Flush(t.Milvus.ctx, t.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush collection: %v", err.Error())
	}

	err = t.Milvus.c.LoadCollection(t.Milvus.ctx, t.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}

func (t *TicketService) IsBlackListed(search *string) (bool, error) {
	if search == nil {
		return false, errors.New("invalid search value")
	}

	hasColl, err := t.Milvus.c.HasCollection(t.Milvus.ctx, "black_ticket")
	if err != nil {
		return false, fmt.Errorf("failed to check if has collection")
	}
	if !hasColl {
		return false, fmt.Errorf("'%s' collection doesnt exist", "black_ticket")
	}

	embedInput := service.Input{
		Inputs: []string{*search},
	}
	searchEmbedding, err := service.GetTextEmbeddings(&embedInput)
	if err != nil {
		return false, fmt.Errorf("failed to get search embeddings: %v", err.Error())
	}

	vector := entity.FloatVector(searchEmbedding[0])
	sp, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return false, fmt.Errorf("failed to create new index flat search param: %v", err.Error())
	}
	searchResults, err := t.Milvus.c.Search(t.Milvus.ctx, "black_ticket", nil, "", []string{"id"}, []entity.Vector{vector}, "ticketContentVector", entity.COSINE, 1, sp)
	if err != nil {
		return false, fmt.Errorf("failed to search ticket: %v", err.Error())
	}

	for _, result := range searchResults {
		for _, score := range result.Scores {
			if score > float32(0.900000) {
				return true, nil
			}
		}
	}

	return false, nil
}

func (t *TicketService) VectorSearch(search *string) (TicketSearchResults, error) {
	if search == nil {
		return nil, errors.New("invalid search value")
	}

	hasColl, err := t.Milvus.c.HasCollection(t.Milvus.ctx, t.CollectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if has collection")
	}
	if !hasColl {
		return nil, fmt.Errorf("'%s' collection doesnt exist", t.CollectionName)
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
	searchResults, err := t.Milvus.c.Search(t.Milvus.ctx, t.CollectionName, nil, "", []string{"ticketId"}, []entity.Vector{vector}, "ticketContentVector", entity.COSINE, 20, sp)
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

func (t *TicketService) CreateTicketCollection() error {
	schema := entity.Schema{
		CollectionName: t.CollectionName,
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

	err := t.Milvus.c.CreateCollection(t.Milvus.ctx, &schema, 1)
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err.Error())
	}

	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 2)
	if err != nil {
		return fmt.Errorf("fail to create ivf flat index: %v", err.Error())
	}

	err = t.Milvus.c.CreateIndex(t.Milvus.ctx, t.CollectionName, "ticketContentVector", idx, false)
	if err != nil {
		return fmt.Errorf("fail to create index: %v", err.Error())
	}

	err = t.Milvus.c.LoadCollection(t.Milvus.ctx, t.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}
