package text

import (
	"regexp"
	"strings"

	"github.com/mkozhukh/tesei"
	"github.com/mkozhukh/tesei/files"
)

type Markdown struct {
	EscapeTagsInContent bool
	LowerCaseLinks      bool
}

type codeBlock struct {
	start int
	end   int
}

func (m Markdown) Run(ctx *tesei.Thread, in <-chan *tesei.Message[files.TextFile], out chan<- *tesei.Message[files.TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[files.TextFile]) (*tesei.Message[files.TextFile], error) {
		if m.EscapeTagsInContent {
			msg.Data.Content = m.escapeTagsInContent(msg.Data.Content)
		}
		if m.LowerCaseLinks {
			msg.Data.Content = m.lowerCaseLinks(msg.Data.Content)
		}
		return msg, nil
	})
}

func (m Markdown) escapeTagsInContent(content string) string {
	// First, identify all code blocks
	blocks := m.findCodeBlocks(content)

	// Find and escape HTML-like tags that are not in code blocks
	// This pattern captures optional markdown formatting (bold/italic) around tags
	// Updated to match tags with attributes like <tag attr="value"> or <tag attr={value}>
	// Also matches self-closing tags like <br/> or <img />
	tagPattern := regexp.MustCompile(`(\*{1,2}|_{1,2})?(<[a-zA-Z]+(?:\s+[^>]*)?/?>)(\*{1,2}|_{1,2})?`)

	// Work with bytes since regex returns byte positions
	result := []byte(content)
	offset := 0

	matches := tagPattern.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		// match[0], match[1] - full match start and end (in bytes)
		// match[2], match[3] - first formatting group (prefix)
		// match[4], match[5] - the tag itself
		// match[6], match[7] - second formatting group (suffix)

		fullStart := match[0]
		fullEnd := match[1]
		tagStart := match[4]
		tagEnd := match[5]

		// Check if this match is inside any code block
		if m.isInCodeBlock(tagStart, tagEnd, blocks) {
			continue
		}

		// Get the components
		fullMatch := content[fullStart:fullEnd]
		tag := content[tagStart:tagEnd]

		// Check if we have matching markdown formatting
		prefixStart, prefixEnd := match[2], match[3]
		suffixStart, suffixEnd := match[6], match[7]

		var prefix, suffix string
		if prefixStart >= 0 {
			prefix = content[prefixStart:prefixEnd]
		}
		if suffixStart >= 0 {
			suffix = content[suffixStart:suffixEnd]
		}

		// Determine the replacement
		var replacement string
		if prefix != "" && suffix != "" && prefix == suffix {
			// We have matching bold/italic markers, remove them
			replacement = "`" + tag + "`"
		} else {
			// No matching markers or only one side, just wrap the tag
			replacement = fullMatch[:tagStart-fullStart] + "`" + tag + "`" + fullMatch[tagEnd-fullStart:]
		}

		// Calculate the position with offset (in bytes)
		adjustedStart := fullStart + offset
		adjustedEnd := fullEnd + offset

		// Replace in result (working with bytes)
		replacementBytes := []byte(replacement)
		newResult := make([]byte, 0, len(result)+(len(replacementBytes)-(adjustedEnd-adjustedStart)))
		newResult = append(newResult, result[:adjustedStart]...)
		newResult = append(newResult, replacementBytes...)
		newResult = append(newResult, result[adjustedEnd:]...)

		result = newResult
		offset += len(replacementBytes) - (fullEnd - fullStart)
	}

	return string(result)
}

func (m Markdown) findCodeBlocks(content string) []codeBlock {
	var blocks []codeBlock

	// Find triple backtick code blocks
	tripleBacktickPattern := regexp.MustCompile("(?s)```.*?```")
	tripleMatches := tripleBacktickPattern.FindAllStringIndex(content, -1)
	for _, match := range tripleMatches {
		blocks = append(blocks, codeBlock{start: match[0], end: match[1]})
	}

	// Find inline code blocks (single backticks on the same line)
	lines := strings.Split(content, "\n")
	currentPos := 0

	for _, line := range lines {
		inlinePattern := regexp.MustCompile("`[^`\n]+`")
		lineMatches := inlinePattern.FindAllStringIndex(line, -1)

		for _, match := range lineMatches {
			absoluteStart := currentPos + match[0]
			absoluteEnd := currentPos + match[1]

			// Check if this inline block is inside a triple backtick block
			isInsideTriple := false
			for _, tripleBlock := range blocks {
				if absoluteStart >= tripleBlock.start && absoluteEnd <= tripleBlock.end {
					isInsideTriple = true
					break
				}
			}

			if !isInsideTriple {
				blocks = append(blocks, codeBlock{start: absoluteStart, end: absoluteEnd})
			}
		}

		currentPos += len(line) + 1 // +1 for newline
	}

	return blocks
}

func (m Markdown) isInCodeBlock(start, end int, blocks []codeBlock) bool {
	for _, block := range blocks {
		// Check if the range overlaps with any code block
		if start >= block.start && start < block.end {
			return true
		}
		if end > block.start && end <= block.end {
			return true
		}
		if start <= block.start && end >= block.end {
			return true
		}
	}
	return false
}

func (m Markdown) lowerCaseLinks(content string) string {
	// Find markdown links: [text](url)
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	result := linkPattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the parts of the link
		matches := linkPattern.FindStringSubmatch(match)
		if len(matches) != 3 {
			return match
		}

		linkText := matches[1]
		linkURL := matches[2]

		// Check if the URL starts with http:// or https://
		if strings.HasPrefix(strings.ToLower(linkURL), "http://") ||
			strings.HasPrefix(strings.ToLower(linkURL), "https://") {
			// Keep external links as-is
			return match
		}

		// Lowercase internal links
		return "[" + linkText + "](" + strings.ToLower(linkURL) + ")"
	})

	return result
}
