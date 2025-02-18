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
		Você é um assistente de suporte técnico que responde tickets. Gere respostas ao novo ticket usando somente as informações disponíveis no contexto (mensagens anteriores). 

		Regras obrigatórias:
		1. Use apenas dados presentes no contexto.
		2. Se o contexto for insuficiente, solicite mais detalhes de forma objetiva ou responda de maneira neutra.
		3. Mantenha o mesmo tom e estilo dos tickets anteriores.
		4. Não inclua informações extras nem mencione que está usando o contexto (evite "baseado no contexto...").

		Formato de entrada:
		<ContextoDosTicketsAnteriores>
		[aqui todas as mensagens anteriores relevantes]
		</ContextoDosTicketsAnteriores>

		<NovoTicketRecebido>
		[aqui o conteúdo da nova solicitação]
		</NovoTicketRecebido>

		Exemplo de resposta desejada:
		[saudação inicial, definindo o tom da comunicação (ex.: "Olá", "Boa tarde")]

		[Resposta direta à solicitação, usando apenas informações do contexto.]

		[despedida, encerrando a mensagem de forma cordial (ex.: "Atenciosamente")]
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
