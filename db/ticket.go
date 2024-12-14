package db

import (
	"encoding/json"
	"fmt"
	"os"
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
	rawData, err := os.ReadFile("./ai/data/outputs/id_list.json")
	if err != nil {
		return "", fmt.Errorf("failed to read from json file: %v", err.Error())
	}

	var jsonData TicketRawData
	err = json.Unmarshal(rawData, &jsonData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal raw ticket data: %v", err.Error())
	}

	for key, value := range jsonData["88"] {
		fmt.Println("key: ", key, "value: ", value)
	}

	return fmt.Sprint(len(jsonData)), nil
}
