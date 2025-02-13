package llm

import (
	"errors"
	"fmt"

	"github.com/wellitonscheer/ticket-helper/internal/db"
	"github.com/wellitonscheer/ticket-helper/internal/service"
)

func SuggestReply(search *string) (string, error) {
	ticketService, err := db.NewTicketService()
	if err != nil {
		return "", fmt.Errorf("failed create ticket service: %s", err.Error())
	}

	tickets, err := ticketService.VectorSearch(search)
	if err != nil {
		return "", fmt.Errorf("failed to search ticket: %s", err.Error())
	}

	allTicketsContent := ""
	for i, ticket := range tickets {
		if i == 0 {
			allTicketsContent = ticket.TicketContent
			continue
		}

		allTicketsContent += fmt.Sprintf(" - %s", ticket.TicketContent)
	}

	systemRole := `
		Você é um assistente que responde tickets de um serviço de suporte. Sua tarefa é gerar uma resposta para um novo ticket usando apenas o contexto fornecido, que consiste em mensagens trocadas anteriormente sobre problemas similares.

		Regras essenciais:
		1. Nunca inclua informações que não estejam no contexto.
		2. Se o contexto não contiver informações suficientes para responder ao novo ticket, sua resposta deve ser neutra.
		3. Se o novo ticket não tiver informações suficientes para que uma resposta baseada no contexto seja formulada, responda solicitando mais detalhes ao solicitante, mantendo a objetividade e o tom das mensagens anteriores.
		4. Adapte o tom para ser coerente com os tickets anteriores.
		5. Não fale coisas como 'Com base no contexto fornecido anteriormente'

		Entrada do modelo:

		<ContextoDosTicketsAnteriores>
		[aqui todas as mensagens anteriores relevantes]
		</ContextoDosTicketsAnteriores>

		<NovoTicketRecebido>
		[aqui o conteúdo da nova solicitação]
		</NovoTicketRecebido>
	`

	userRole := fmt.Sprintf(`
		<ContextoDosTicketsAnteriores>
		%s
		</ContextoDosTicketsAnteriores>

		<NovoTicketRecebido>
		%s
		</NovoTicketRecebido>
	`, allTicketsContent, *search)

	modelResponse, err := service.LmstudioModel(&service.Messages{
		service.Message{
			Role:    "system",
			Content: systemRole,
		},
		service.Message{
			Role:    "user",
			Content: userRole,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get suggestion: %s", err.Error())
	}

	if len(modelResponse.Choices) == 0 {
		return "", errors.New("model didnt return any suggestion")
	}

	return modelResponse.Choices[0].Message.Content, nil
}
