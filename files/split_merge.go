package files

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mkozhukh/tesei"
)

// Split splits a TextFile into multiple chunks based on a user-defined rule.
type Split struct {
	// By is the function that splits the text content.
	// It returns a slice of strings, where each string is a chunk.
	By func(text string) []string
}

// Run executes the split logic.
func (s Split) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	defer close(out)

	for msg := range in {
		if msg.Error != nil {
			out <- msg
			continue
		}

		chunks := s.By(msg.Data.Content)
		total := len(chunks)

		for i, chunk := range chunks {
			// Create a new message for each chunk
			newMsg := msg.Clone()
			newMsg.ID = fmt.Sprintf("%s_%d", msg.ID, i)
			newMsg.Data.Content = chunk

			// Set metadata for merging
			newMsg.Metadata["split_id"] = msg.ID
			newMsg.Metadata["split_index"] = i
			newMsg.Metadata["split_total"] = total

			select {
			case out <- newMsg:
			case <-ctx.Done():
				return
			}
		}
	}
}

// Merge collects chunks and merges them back into a single file.
type Merge struct {
	// Glue is the string used to join chunks. Defaults to empty string.
	Glue string
	// By is an optional custom function to join chunks.
	// If provided, it overrides Glue.
	By func(chunks []string) string
}

// Run executes the merge logic.
func (m Merge) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	defer close(out)

	// Buffer to store chunks: split_id -> []*tesei.Message[TextFile]
	buffer := make(map[string][]*tesei.Message[TextFile])

	for msg := range in {
		if msg.Error != nil {
			out <- msg
			continue
		}

		splitID, ok := msg.Metadata["split_id"].(string)
		if !ok {
			// Not a split chunk, pass through
			out <- msg
			continue
		}

		splitTotal, _ := msg.Metadata["split_total"].(int)

		buffer[splitID] = append(buffer[splitID], msg)

		// Check if we have all chunks
		if len(buffer[splitID]) == splitTotal {
			chunks := buffer[splitID]
			delete(buffer, splitID)

			// Sort chunks by index
			sort.Slice(chunks, func(i, j int) bool {
				idxI, _ := chunks[i].Metadata["split_index"].(int)
				idxJ, _ := chunks[j].Metadata["split_index"].(int)
				return idxI < idxJ
			})

			// Extract content
			strChunks := make([]string, len(chunks))
			for i, c := range chunks {
				strChunks[i] = c.Data.Content
			}

			// Merge
			var mergedContent string
			if m.By != nil {
				mergedContent = m.By(strChunks)
			} else {
				mergedContent = strings.Join(strChunks, m.Glue)
			}

			// Create output message using the first chunk as a template
			// We restore the original ID (which is split_id)
			outMsg := chunks[0].Clone()
			outMsg.ID = splitID
			outMsg.Data.Content = mergedContent

			// Clean up split metadata
			delete(outMsg.Metadata, "split_id")
			delete(outMsg.Metadata, "split_index")
			delete(outMsg.Metadata, "split_total")

			select {
			case out <- outMsg:
			case <-ctx.Done():
				return
			}
		}
	}
}

// Clone generates multiple messages from a single input message using a custom handler.
// Unlike Split, it does not add metadata for merging.
type Clone struct {
	// By is the function that generates new messages from the input message.
	By func(msg *tesei.Message[TextFile]) []*tesei.Message[TextFile]
}

// Run executes the clone logic.
func (c Clone) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	defer close(out)

	for msg := range in {
		if msg.Error != nil {
			out <- msg
			continue
		}

		if c.By == nil {
			// If no handler provided, clone once
			out <- msg.Clone()
			continue
		}

		results := c.By(msg)
		for _, res := range results {
			select {
			case out <- res:
			case <-ctx.Done():
				return
			}
		}
	}
}
