package types

type ClientEmbeddingInputs struct {
	Inputs []string `json:"inputs"`
}

type ClientEmbeddings = [][]float32
