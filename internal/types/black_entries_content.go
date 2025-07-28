package types

type BlackEntryContent struct {
	Id      int    `json:"id"`
	Content string `json:"content"`
}

type BlackEntriesContent = []BlackEntryContent
