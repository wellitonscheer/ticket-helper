package utils

import (
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

type EntryCleaner struct {
	Strip *bluemonday.Policy
}

func NewEntryCleaner() EntryCleaner {
	strip := bluemonday.StrictPolicy()

	return EntryCleaner{
		Strip: strip,
	}
}

func (cl EntryCleaner) Clean(entry string) string {
	if isUseless := cl.IsUselessEntry(entry); isUseless {
		return ""
	}

	noPastEmails := cl.RemovePastEmailsFromEntry(entry)
	if noPastEmails == "" {
		return ""
	}

	noHtml := cl.RemoveHTMLTags(noPastEmails)
	if noHtml == "" {
		return ""
	}

	return strings.TrimSpace(noHtml)
}

func (cl EntryCleaner) RemovePastEmailsFromEntry(entry string) string {
	padrao := `(?i)(De:|Em (seg|ter|qua|qui|sex|sáb|dom)\.?,? \d{1,2} de .+?escreveu:|---------- Forwarded message ---------|----- Original message -----)`

	re := regexp.MustCompile(padrao)
	indices := re.FindStringIndex(entry)

	if indices != nil {
		return strings.TrimSpace(entry[:indices[0]])
	}

	return entry
}

func (cl EntryCleaner) RemoveHTMLTags(entry string) string {
	return cl.Strip.Sanitize(entry)
}

func (cl EntryCleaner) IsUselessEntry(entry string) bool {
	uselessEntries := []string{
		"Seu feedback é muito importante",
		"Recebemos sua solicitação, assim que possível lhe retornamos",
		"A solicitação foi encerrada",
		"Task closed",
		"fechado automaticamente pela ausência de retorno",
		"Hello suporte@setrem.com.br",
	}

	for _, useless := range uselessEntries {
		if strings.Contains(entry, useless) {
			return true
		}
	}

	if strings.Contains(entry, "Olá, SETREM") && strings.Contains(entry, "A disciplina") && strings.Contains(entry, "foi aprovada") {
		return true
	}

	return false
}
