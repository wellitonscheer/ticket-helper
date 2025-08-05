package utils

import (
	"github.com/wellitonscheer/ticket-helper/internal/types"
)

const (
	defaultChunkSize   = 50
	defaultOverlapSize = 10
)

func ChunkText(input types.ChunkTextInput) []string {
	var chunks []string

	if input.ChunkSize == 0 {
		input.ChunkSize = defaultChunkSize
	}
	if input.OverlapSize == 0 {
		input.OverlapSize = defaultOverlapSize
	}

	start := 0
	end := input.ChunkSize
	iStart := start
	iEnd := end
	textLen := len(input.Text)
	for {
		// fmt.Printf("LOOP: start: '%v', end: '%v', range: %v\n", iStart, iEnd, iEnd-iStart)
		if iStart > 0 && input.Text[iStart] != byte(' ') {
			// fmt.Printf("start char '%s' to '%s'\n", string(input.Text[iStart]), string(input.Text[iStart-1]))
			iStart = iStart - 1
			continue
		}
		if iEnd < textLen && input.Text[iEnd] != byte(' ') {
			// fmt.Printf("end char '%s' to '%s'\n", string(input.Text[iEnd]), string(input.Text[iEnd+1]))
			iEnd = iEnd + 1
			continue
		}

		if iEnd >= textLen {
			chunk := input.Text[iStart:textLen]
			chunks = append(chunks, chunk)
			// fmt.Printf("chunk: '%+s'\n", chunk)
			break
		}

		if iStart > 0 {
			iStart = iStart + 1
		}

		chunk := input.Text[iStart:iEnd]
		chunks = append(chunks, chunk)
		// fmt.Printf("chunk: '%+s'\n", chunk)

		end = (end + input.ChunkSize) - input.OverlapSize
		start = (start + input.ChunkSize) - input.OverlapSize
		iStart = start
		iEnd = end
	}

	return chunks
}
