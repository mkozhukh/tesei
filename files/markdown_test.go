package files

import (
	"context"
	"testing"

	"github.com/mkozhukh/tesei"
)

func TestMarkdown_EscapeTagsInContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple HTML tag outside code block",
			input:    "This is a <div> tag that should be escaped",
			expected: "This is a `<div>` tag that should be escaped",
		},
		{
			name:     "Uppercase HTML tag",
			input:    "This is a <DIV> tag that should be escaped",
			expected: "This is a `<DIV>` tag that should be escaped",
		},
		{
			name:     "Mixed case HTML tag",
			input:    "This is a <Div> tag that should be escaped",
			expected: "This is a `<Div>` tag that should be escaped",
		},
		{
			name:     "HTML tag with slash",
			input:    "This is a <br /> tag that should be escaped",
			expected: "This is a `<br />` tag that should be escaped",
		},
		{
			name:     "Uppercase HTML tag with slash",
			input:    "This is a <BR /> tag that should be escaped",
			expected: "This is a `<BR />` tag that should be escaped",
		},
		{
			name:     "Multiple HTML tags with mixed cases",
			input:    "Here are <div> and <SPAN> and <Input> tags",
			expected: "Here are `<div>` and `<SPAN>` and `<Input>` tags",
		},
		{
			name:     "HTML tag inside code block should not be escaped",
			input:    "```\n<div> tag in code block\n```",
			expected: "```\n<div> tag in code block\n```",
		},
		{
			name:     "Uppercase HTML tag inside code block should not be escaped",
			input:    "```\n<DIV> tag in code block\n```",
			expected: "```\n<DIV> tag in code block\n```",
		},
		{
			name:     "HTML tag inside inline code should not be escaped",
			input:    "The `<div>` tag is already in inline code",
			expected: "The `<div>` tag is already in inline code",
		},
		{
			name:     "Uppercase HTML tag inside inline code should not be escaped",
			input:    "The `<DIV>` tag is already in inline code",
			expected: "The `<DIV>` tag is already in inline code",
		},
		{
			name:     "Mixed content with code blocks and HTML tags",
			input:    "Outside <div> tag\n```\n<span> inside code\n```\nAnother <P> tag",
			expected: "Outside `<div>` tag\n```\n<span> inside code\n```\nAnother `<P>` tag",
		},
		{
			name:     "Inline code and regular HTML tags on same line",
			input:    "Use `<code>` for inline and <DIV> for block",
			expected: "Use `<code>` for inline and `<DIV>` for block",
		},
		{
			name:     "No HTML tags",
			input:    "This is plain text without any tags",
			expected: "This is plain text without any tags",
		},
		{
			name:     "HTML tag at beginning and end with mixed cases",
			input:    "<START> some text <End>",
			expected: "`<START>` some text `<End>`",
		},
		{
			name:     "Complex markdown with multiple code blocks and mixed case tags",
			input:    "# Title\n\nParagraph with <Tag> here.\n\n```go\nfunc test() {\n\t<NotEscaped>\n}\n```\n\nInline `<AlreadyEscaped>` and regular <NeedsEscape> tag.\n\n```\n<AnotherBlock>\n```",
			expected: "# Title\n\nParagraph with `<Tag>` here.\n\n```go\nfunc test() {\n\t<NotEscaped>\n}\n```\n\nInline `<AlreadyEscaped>` and regular `<NeedsEscape>` tag.\n\n```\n<AnotherBlock>\n```",
		},
		{
			name:     "All uppercase tags",
			input:    "<HTML> <HEAD> <BODY> <DIV> <SPAN>",
			expected: "`<HTML>` `<HEAD>` `<BODY>` `<DIV>` `<SPAN>`",
		},
		{
			name:     "Self-closing uppercase tags",
			input:    "Line break <BR/> and horizontal rule <HR />",
			expected: "Line break `<BR/>` and horizontal rule `<HR />`",
		},
		{
			name:     "Bold wrapped tag",
			input:    "This is a **<div>** tag with bold",
			expected: "This is a `<div>` tag with bold",
		},
		{
			name:     "Italic wrapped tag with asterisk",
			input:    "This is an *<span>* tag with italic",
			expected: "This is an `<span>` tag with italic",
		},
		{
			name:     "Italic wrapped tag with underscore",
			input:    "This is an _<p>_ tag with italic",
			expected: "This is an `<p>` tag with italic",
		},
		{
			name:     "Double underscore bold wrapped tag",
			input:    "This is a __<DIV>__ tag with bold",
			expected: "This is a `<DIV>` tag with bold",
		},
		{
			name:     "Mixed bold and italic tags",
			input:    "Here are **<div>** and *<span>* and regular <p> tags",
			expected: "Here are `<div>` and `<span>` and regular `<p>` tags",
		},
		{
			name:     "Mismatched formatting markers",
			input:    "This has **<div>* mismatched markers",
			expected: "This has **`<div>`* mismatched markers",
		},
		{
			name:     "Tag with only prefix formatting",
			input:    "This has **<div> without suffix",
			expected: "This has **`<div>` without suffix",
		},
		{
			name:     "Tag with only suffix formatting",
			input:    "This has <div>** without prefix",
			expected: "This has `<div>`** without prefix",
		},
		{
			name:     "Bold wrapped tag in code block should not be escaped",
			input:    "```\n**<div>** tag in code block\n```",
			expected: "```\n**<div>** tag in code block\n```",
		},
		{
			name:     "Bold wrapped tag in inline code should not be escaped",
			input:    "The `**<div>**` tag is already in inline code",
			expected: "The `**<div>**` tag is already in inline code",
		},
		{
			name:     "Complex markdown with formatted tags",
			input:    "# Title\n\nParagraph with **<Tag>** here and *<Another>* there.\n\n```go\n**<NotEscaped>**\n```\n\nRegular <NeedsEscape> tag.",
			expected: "# Title\n\nParagraph with `<Tag>` here and `<Another>` there.\n\n```go\n**<NotEscaped>**\n```\n\nRegular `<NeedsEscape>` tag.",
		},
		{
			name:     "Multiple same tags on different lines",
			input:    "First <Locale> tag\nSecond <Locale> tag\nThird <Locale> tag",
			expected: "First `<Locale>` tag\nSecond `<Locale>` tag\nThird `<Locale>` tag",
		},
		{
			name:     "Multiple tags with mixed formatting",
			input:    "First **<Locale>** tag\nSecond *<Locale>* tag\nThird <Locale> tag\nFourth <Locale> tag",
			expected: "First `<Locale>` tag\nSecond `<Locale>` tag\nThird `<Locale>` tag\nFourth `<Locale>` tag",
		},
		{
			name:     "Tag with className attribute using curly braces",
			input:    "Icon component <i className={item.icon}> here",
			expected: "Icon component `<i className={item.icon}>` here",
		},
		{
			name:     "Tag with className attribute using quotes",
			input:    "Div with class <div className=\"test\"> here",
			expected: "Div with class `<div className=\"test\">` here",
		},
		{
			name:     "Tag with multiple attributes",
			input:    "Button <button onClick={handleClick} className=\"btn primary\"> here",
			expected: "Button `<button onClick={handleClick} className=\"btn primary\">` here",
		},
		{
			name:     "Self-closing tag with attributes",
			input:    "Image <img src=\"/path/to/image.jpg\" alt=\"description\" /> here",
			expected: "Image `<img src=\"/path/to/image.jpg\" alt=\"description\" />` here",
		},
		{
			name:     "Bold wrapped tag with attributes",
			input:    "This is **<span className=\"highlight\">** with bold",
			expected: "This is `<span className=\"highlight\">` with bold",
		},
		{
			name:     "Italic wrapped tag with attributes",
			input:    "This is *<a href=\"https://example.com\">* with italic",
			expected: "This is `<a href=\"https://example.com\">` with italic",
		},
		{
			name:     "Tag with JSX spread attributes",
			input:    "Component <MyComponent {...props}> here",
			expected: "Component `<MyComponent {...props}>` here",
		},
		{
			name:     "Tag with data attributes",
			input:    "Element <div data-id=\"123\" data-name=\"test\"> here",
			expected: "Element `<div data-id=\"123\" data-name=\"test\">` here",
		},
		{
			name:     "Multiple tags with attributes on different lines",
			input:    "First <input type=\"text\" value={value}> tag\nSecond <select onChange={handleChange}> tag\nThird <textarea rows=\"5\"> tag",
			expected: "First `<input type=\"text\" value={value}>` tag\nSecond `<select onChange={handleChange}>` tag\nThird `<textarea rows=\"5\">` tag",
		},
		{
			name:     "Tag with attributes in code block should not be escaped",
			input:    "```\n<div className=\"container\"> in code block\n```",
			expected: "```\n<div className=\"container\"> in code block\n```",
		},
		{
			name:     "Tag with attributes in inline code should not be escaped",
			input:    "The `<span className=\"inline\">` is already in code",
			expected: "The `<span className=\"inline\">` is already in code",
		},
		{
			name:     "Complex JSX with mixed tags",
			input:    "Render <div className={styles.container}> with **<span style={{color: 'red'}}>** and plain <input type=\"checkbox\" />",
			expected: "Render `<div className={styles.container}>` with `<span style={{color: 'red'}}>` and plain `<input type=\"checkbox\" />`",
		},
		{
			name:     "UTF-8 characters before tag",
			input:    "```js title=\"de.js\"\nAuswählen\n```\n\n- <Locale> tag",
			expected: "```js title=\"de.js\"\nAuswählen\n```\n\n- `<Locale>` tag",
		},
		{
			name:     "Multiple UTF-8 characters with tags",
			input:    "Über uns <div> café <span> naïve <p> tag",
			expected: "Über uns `<div>` café `<span>` naïve `<p>` tag",
		},
		{
			name:     "Chinese characters with tags",
			input:    "中文内容 <div> 更多文字 <span> 结束",
			expected: "中文内容 `<div>` 更多文字 `<span>` 结束",
		},
		{
			name:     "Emoji with tags",
			input:    "Hello 👋 <div> world 🌍 <span> end",
			expected: "Hello 👋 `<div>` world 🌍 `<span>` end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fix := Markdown{EscapeTagsInContent: true}
			result := fix.escapeTagsInContent(tt.input)
			if result != tt.expected {
				t.Errorf("escapeTagsInContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdown_Run(t *testing.T) {
	// Create a test message
	in := make(chan *tesei.Message[TextFile], 1)
	out := make(chan *tesei.Message[TextFile], 1)

	testContent := "This has a <DIV> tag and ```\n<CODE> in block\n```"
	expectedContent := "This has a `<DIV>` tag and ```\n<CODE> in block\n```"

	msg := &tesei.Message[TextFile]{
		Data: TextFile{
			Name:    "test.md",
			Folder:  "/test",
			Content: testContent,
		},
	}

	in <- msg
	close(in)

	fix := Markdown{EscapeTagsInContent: true}
	ctx := tesei.NewThread(context.Background(), 10)

	// Run in a goroutine since it processes channels
	go fix.Run(ctx, in, out)

	// Read the result
	result := <-out

	if result.Data.Content != expectedContent {
		t.Errorf("Run() transformed content = %q, want %q", result.Data.Content, expectedContent)
	}
}

func TestMarkdown_DisabledRule(t *testing.T) {
	// Test that when EscapeTagsInContent is false, no transformation occurs
	in := make(chan *tesei.Message[TextFile], 1)
	out := make(chan *tesei.Message[TextFile], 1)

	testContent := "This has a <DIV> tag that should not be escaped"

	msg := &tesei.Message[TextFile]{
		Data: TextFile{
			Name:    "test.md",
			Folder:  "/test",
			Content: testContent,
		},
	}

	in <- msg
	close(in)

	fix := Markdown{EscapeTagsInContent: false}
	ctx := tesei.NewThread(context.Background(), 10)

	// Run in a goroutine since it processes channels
	go fix.Run(ctx, in, out)

	// Read the result
	result := <-out

	if result.Data.Content != testContent {
		t.Errorf("Run() with disabled rule should not transform content, got %q, want %q", result.Data.Content, testContent)
	}
}

func TestMarkdown_LowerCaseLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple internal link with uppercase",
			input:    "Check [this link](/Docs/Guide.md) for more info",
			expected: "Check [this link](/docs/guide.md) for more info",
		},
		{
			name:     "Multiple internal links",
			input:    "See [first](/API/Users.md) and [second](/API/Products.MD)",
			expected: "See [first](/api/users.md) and [second](/api/products.md)",
		},
		{
			name:     "External http link should not change",
			input:    "Visit [Google](http://google.com/SEARCH) for search",
			expected: "Visit [Google](http://google.com/SEARCH) for search",
		},
		{
			name:     "External https link should not change",
			input:    "Visit [Example](https://Example.COM/PATH) for info",
			expected: "Visit [Example](https://Example.COM/PATH) for info",
		},
		{
			name:     "Mixed internal and external links",
			input:    "Check [internal](/Docs/README.md) and [external](https://github.com/USER/REPO)",
			expected: "Check [internal](/docs/readme.md) and [external](https://github.com/USER/REPO)",
		},
		{
			name:     "Relative path link",
			input:    "See [relative](../Parent/File.MD) for details",
			expected: "See [relative](../parent/file.md) for details",
		},
		{
			name:     "Link with anchor",
			input:    "Jump to [section](/Guide/Setup.md#Installation)",
			expected: "Jump to [section](/guide/setup.md#installation)",
		},
		{
			name:     "Link with query parameters",
			input:    "Open [page](/Search?Query=TEST&Page=1)",
			expected: "Open [page](/search?query=test&page=1)",
		},
		{
			name:     "Already lowercase link",
			input:    "This [link](/docs/guide.md) is already lowercase",
			expected: "This [link](/docs/guide.md) is already lowercase",
		},
		{
			name:     "Link with spaces in text",
			input:    "Click [This Is A Link](/Path/To/FILE.md) here",
			expected: "Click [This Is A Link](/path/to/file.md) here",
		},
		{
			name:     "Link with special characters in path",
			input:    "View [file](/Path/With-Special_Chars/FILE.MD)",
			expected: "View [file](/path/with-special_chars/file.md)",
		},
		{
			name:     "Empty link",
			input:    "Empty [link]() here",
			expected: "Empty [link]() here",
		},
		{
			name:     "Link with only filename",
			input:    "Open [file](README.MD)",
			expected: "Open [file](readme.md)",
		},
		{
			name:     "Multiple links on same line",
			input:    "See [first](/First.md), [second](/SECOND.md), and [third](https://EXAMPLE.com)",
			expected: "See [first](/first.md), [second](/second.md), and [third](https://EXAMPLE.com)",
		},
		{
			name:     "Link in list item",
			input:    "- Item with [link](/Docs/API.md)\n- Another item",
			expected: "- Item with [link](/docs/api.md)\n- Another item",
		},
		{
			name:     "Link with uppercase HTTP prefix",
			input:    "Visit [site](HTTP://example.com/PATH) now",
			expected: "Visit [site](HTTP://example.com/PATH) now",
		},
		{
			name:     "Link with uppercase HTTPS prefix",
			input:    "Visit [site](HTTPS://example.com/PATH) now",
			expected: "Visit [site](HTTPS://example.com/PATH) now",
		},
		{
			name:     "No links in text",
			input:    "This text has no markdown links at all",
			expected: "This text has no markdown links at all",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fix := Markdown{LowerCaseLinks: true}
			result := fix.lowerCaseLinks(tt.input)
			if result != tt.expected {
				t.Errorf("lowerCaseLinks() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMarkdown_RunWithLowerCaseLinks(t *testing.T) {
	// Create a test message
	in := make(chan *tesei.Message[TextFile], 1)
	out := make(chan *tesei.Message[TextFile], 1)

	testContent := "Check [this](/Docs/Guide.MD) and visit [external](https://GitHub.com/USER)"
	expectedContent := "Check [this](/docs/guide.md) and visit [external](https://GitHub.com/USER)"

	msg := &tesei.Message[TextFile]{
		Data: TextFile{
			Name:    "test.md",
			Folder:  "/test",
			Content: testContent,
		},
	}

	in <- msg
	close(in)

	fix := Markdown{LowerCaseLinks: true}
	ctx := tesei.NewThread(context.Background(), 10)

	// Run in a goroutine since it processes channels
	go fix.Run(ctx, in, out)

	// Read the result
	result := <-out

	if result.Data.Content != expectedContent {
		t.Errorf("Run() with LowerCaseLinks = %q, want %q", result.Data.Content, expectedContent)
	}
}

func TestMarkdown_BothRulesEnabled(t *testing.T) {
	// Test with both EscapeTagsInContent and LowerCaseLinks enabled
	in := make(chan *tesei.Message[TextFile], 1)
	out := make(chan *tesei.Message[TextFile], 1)

	testContent := "Has <DIV> tag and [link](/Path/To/FILE.md) and [external](https://EXAMPLE.com)"
	expectedContent := "Has `<DIV>` tag and [link](/path/to/file.md) and [external](https://EXAMPLE.com)"

	msg := &tesei.Message[TextFile]{
		Data: TextFile{
			Name:    "test.md",
			Folder:  "/test",
			Content: testContent,
		},
	}

	in <- msg
	close(in)

	fix := Markdown{
		EscapeTagsInContent: true,
		LowerCaseLinks:      true,
	}
	ctx := tesei.NewThread(context.Background(), 10)

	// Run in a goroutine since it processes channels
	go fix.Run(ctx, in, out)

	// Read the result
	result := <-out

	if result.Data.Content != expectedContent {
		t.Errorf("Run() with both rules = %q, want %q", result.Data.Content, expectedContent)
	}
}
