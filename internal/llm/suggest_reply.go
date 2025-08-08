package llm

import (
	"fmt"

	"github.com/wellitonscheer/ticket-helper/internal/client"
	"github.com/wellitonscheer/ticket-helper/internal/context"
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

func SuggestReply(appCtx context.AppContext, search, context *string) (string, error) {
	systemRole := `
		Você é um assistente de suporte técnico que responde tickets. Gere respostas ao novo ticket usando somente as informações disponíveis no contexto (mensagens anteriores). 

		Regras obrigatórias:
		1. Nunca invente informações. Use apenas os dados presentes no contexto.
		2. Se o contexto for insuficiente, solicite mais detalhes de forma objetiva ou responda de maneira neutra.
		3. Mantenha o mesmo tom e estilo dos tickets anteriores.
		4. **Não mencione o contexto nem o uso de IA**. Fale como se fosse uma pessoa continuando o atendimento.
		5. Mensagens anteriores podem estar desatualizadas. **Evite tratar eventos antigos como se ainda estivessem válidos.**
		6. Use o histórico apenas para entender padrões, responsáveis ou procedimentos, mas não afirme que algo antigo ainda está ocorrendo.

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
	`, *context, *search)

	fullContextTokens := (len(userRole) + len(systemRole)) / 4
	if fullContextTokens > appCtx.Config.LLM.LLMContextLengthTokens {
		return "", fmt.Errorf("failed to get suggestion, to many context tokens")
	}

	modelResponse, err := client.LmstudioModel(appCtx, &types.LMSMessages{
		types.LMSRoleMessage{
			Role:    "system",
			Content: systemRole,
		},
		types.LMSRoleMessage{
			Role:    "user",
			Content: userRole,
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get suggestion: %s", err.Error())
	}

	if len(modelResponse.Choices) == 0 {
		return "", fmt.Errorf("model didnt return any suggestion")
	}

	return modelResponse.Choices[0].Message.Content, nil
}
