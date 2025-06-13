package milservi

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/wellitonscheer/ticket-helper/internal/client"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/database/milvus"
	"github.com/wellitonscheer/ticket-helper/internal/service"
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

type TicketSearchResult struct {
	Id            int64
	TicketId      string
	TicketContent string
	Score         float32
}

type TicketSearchResults = []TicketSearchResult

type TicketService struct {
	AppContext context.AppContext
	Milvus         *milvus.MilvusClient
	CollectionName string
}

func NewTicketService(appContext context.AppContext) TicketService {
	return TicketService{
		AppContext: appContext,
		Milvus:         appContext.Milvus,
		CollectionName: "ticket",
	}
}

func (t *TicketService) InsertAllTickets() error {
	collExists, err := t.Milvus.Client.HasCollection(t.Milvus.Ctx, t.CollectionName)
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

		data := client.GetTextEmbeddingsInput{
			Inputs: []string{fullBodyMessage},
		}
		clientEmbedding := client.NewEmbeddingClient(t.AppContext)
		embeddedBodyMessage, err := clientEmbedding.GetTextEmbeddings(&data)
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

		_, err = t.Milvus.Client.Insert(t.Milvus.Ctx, t.CollectionName, "", ticketIdColumn, ticketContentColumn, ticketContentVecColumn)
		if err != nil {
			return fmt.Errorf("failed to insert tickets: %v", err.Error())
		}
	}

	fmt.Println("DONE inserting...")

	err = t.Milvus.Client.Flush(t.Milvus.Ctx, t.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to flush collection: %v", err.Error())
	}

	err = t.Milvus.Client.LoadCollection(t.Milvus.Ctx, t.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}

func (t *TicketService) IsBlackListed(search *string) (bool, error) {
	if search == nil {
		return false, errors.New("invalid search value")
	}

	hasColl, err := t.Milvus.Client.HasCollection(t.Milvus.Ctx, "black_ticket")
	if err != nil {
		return false, fmt.Errorf("failed to check if has collection")
	}
	if !hasColl {
		return false, fmt.Errorf("'%s' collection doesnt exist", "black_ticket")
	}

	embedInput := client.GetTextEmbeddingsInput{
		Inputs: []string{*search},
	}
	clientEmbedding := client.NewEmbeddingClient(t.AppContext)

	searchEmbedding, err := clientEmbedding.GetTextEmbeddings(&embedInput)
	if err != nil {
		return false, fmt.Errorf("failed to get search embeddings: %v", err.Error())
	}

	vector := entity.FloatVector(searchEmbedding[0])
	sp, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return false, fmt.Errorf("failed to create new index flat search param: %v", err.Error())
	}
	searchResults, err := t.Milvus.Client.Search(t.Milvus.Ctx, "black_ticket", nil, "", []string{"id"}, []entity.Vector{vector}, "ticketContentVector", entity.COSINE, 1, sp)
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

func (t *TicketService) VectorSearchTicketsIds(search string) (types.TicketSearchResults, error) {
	if search == "" {
		return nil, errors.New("invalid search value")
	}

	// hasColl, err := t.Milvus.Client.HasCollection(t.Milvus.Ctx, t.CollectionName)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to check if has collection")
	// }
	// if !hasColl {
	// 	return nil, fmt.Errorf("'%s' collection doesnt exist", t.CollectionName)
	// }

	

	embedInput := client.GetTextEmbeddingsInput{
		Inputs: []string{search},
	}
	clientEmbedding := client.NewEmbeddingClient(t.AppContext)
	searchEmbedding, err := clientEmbedding.GetTextEmbeddings(&embedInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get search embeddings: %v", err.Error())
	}

	vector := entity.FloatVector(searchEmbedding[0])
	sp, err := entity.NewIndexFlatSearchParam()
	if err != nil {
		return nil, fmt.Errorf("failed to create new index flat search param: %v", err.Error())
	}
	searchResults, err := t.Milvus.Client.Search(t.Milvus.Ctx, t.CollectionName, nil, "", []string{"ticketId"}, []entity.Vector{vector}, "ticketContentVector", entity.COSINE, 20, sp)
	if err != nil {
		return nil, fmt.Errorf("failed to search ticket: %v", err.Error())
	}

	var ticketsResult types.TicketSearchResults
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
				ticketsResult = append(ticketsResult, types.TicketSearchResult{TicketId: ticketsIds[i], Score: score})
			}
		}
	}

	return ticketsResult, nil
}

func (t *TicketService) VectorSearch(search *string) (TicketSearchResults, error) {
	if search == nil {
		return nil, errors.New("invalid search value")
	}

	hasColl, err := t.Milvus.Client.HasCollection(t.Milvus.Ctx, t.CollectionName)
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
	searchResults, err := t.Milvus.Client.Search(t.Milvus.Ctx, t.CollectionName, nil, "", []string{"ticketId", "id", "ticketContent"}, []entity.Vector{vector}, "ticketContentVector", entity.COSINE, 20, sp)
	if err != nil {
		return nil, fmt.Errorf("failed to search ticket: %v", err.Error())
	}

	var ticketsResult TicketSearchResults

	var idColumn *entity.ColumnInt64
	var ticketIdColumn *entity.ColumnVarChar
	var ticketContentColumn *entity.ColumnVarChar
	for _, result := range searchResults {
		for _, field := range result.Fields {
			switch field.Name() {
			case "id":
				c, ok := field.(*entity.ColumnInt64)
				if ok {
					idColumn = c
				}
			case "ticketId":
				c, ok := field.(*entity.ColumnVarChar)
				if ok {
					ticketIdColumn = c
				}
			case "ticketContent":
				c, ok := field.(*entity.ColumnVarChar)
				if ok {
					ticketContentColumn = c
				}
			}
		}

		ids := idColumn.Data()
		ticketsIds := ticketIdColumn.Data()
		ticketsContent := ticketContentColumn.Data()

		for i, score := range result.Scores {
			ticketsResult = append(ticketsResult, TicketSearchResult{Id: ids[i], TicketId: ticketsIds[i], TicketContent: ticketsContent[i], Score: score})
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

	err := t.Milvus.Client.CreateCollection(t.Milvus.Ctx, &schema, 1)
	if err != nil {
		return fmt.Errorf("failed to create collection: %v", err.Error())
	}

	idx, err := entity.NewIndexIvfFlat(entity.COSINE, 2)
	if err != nil {
		return fmt.Errorf("fail to create ivf flat index: %v", err.Error())
	}

	err = t.Milvus.Client.CreateIndex(t.Milvus.Ctx, t.CollectionName, "ticketContentVector", idx, false)
	if err != nil {
		return fmt.Errorf("fail to create index: %v", err.Error())
	}

	err = t.Milvus.Client.LoadCollection(t.Milvus.Ctx, t.CollectionName, false)
	if err != nil {
		return fmt.Errorf("failed to load collection: %v", err.Error())
	}

	return nil
}
