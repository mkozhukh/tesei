package llm

import (
	"fmt"

	"github.com/mkozhukh/echo"
	templates "github.com/mkozhukh/echo-templates"
	"github.com/mkozhukh/tesei"
	"github.com/mkozhukh/tesei/files"
)

var model string
var apiKey string
var templatesPath string
var templatesSource templates.TemplateSource

func init() {
	model = "google/fast"
	templatesPath = "~/.prompts"
}

func SetTemplatesPath(path string) {
	templatesPath = path
}

func SetTemplatesSource(source templates.TemplateSource) {
	templatesSource = source
}

func SetModel(m string) {
	model = m
}

func SetAPIKey(a string) {
	apiKey = a
}

type Echo struct {
	Model         string
	APIKey        string
	TemplatesPath string
	Client        echo.Client

	templatesEngine templates.TemplateEngine
}

func (c *Echo) init(ctx *tesei.Thread) error {
	if c.Client != nil {
		return nil
	}

	m := c.Model
	if m == "" {
		m = model
	}

	a := c.APIKey
	if a == "" {
		a = apiKey
	}

	var err error
	c.Client, err = echo.NewClient(m, a)
	if err != nil {
		ctx.Error() <- err
		return err
	}

	return nil
}

func (c *Echo) initTemplatesEngine(ctx *tesei.Thread) error {
	path := c.TemplatesPath
	if path == "" {
		path = templatesPath
	}

	if path == "" && templatesSource == nil {
		err := fmt.Errorf("templates path is not set")
		ctx.Error() <- err
		return err
	}

	source := templatesSource

	var err error
	if source == nil {
		source, err = templates.NewFileSystemSource(path)
		if err != nil {
			ctx.Error() <- err
			return err
		}
	}

	c.templatesEngine, err = templates.New(templates.Config{Source: source})
	if err != nil {
		ctx.Error() <- err
		return err
	}

	return nil
}

type CompleteContent struct {
	Echo
	Prompt string
}

func (c CompleteContent) Run(ctx *tesei.Thread, in <-chan *tesei.Message[files.TextFile], out chan<- *tesei.Message[files.TextFile]) {
	err := c.init(ctx)
	if err != nil {
		return
	}

	tesei.Transform(ctx, in, out, func(msg *tesei.Message[files.TextFile]) (*tesei.Message[files.TextFile], error) {
		response, err := c.Client.Call(ctx, echo.QuickMessage(msg.Data.Content), echo.WithSystemMessage(c.Prompt))
		if err != nil {
			return msg, fmt.Errorf("complete: %w", err)
		}

		msg.Data.Content = response.Text
		return msg, nil
	})
}

type CompleteTemplateString struct {
	Echo
	Vars     map[string]any
	Template string
}

func (c CompleteTemplateString) Run(ctx *tesei.Thread, in <-chan *tesei.Message[files.TextFile], out chan<- *tesei.Message[files.TextFile]) {
	err := c.init(ctx)
	if err != nil {
		return
	}

	tesei.Transform(ctx, in, out, func(msg *tesei.Message[files.TextFile]) (*tesei.Message[files.TextFile], error) {
		vars := extend(msg.Metadata, c.Vars, msg)
		messages, meta, err := templates.GenerateWithMetadata(c.Template, vars)
		if err != nil {
			return msg, fmt.Errorf("complete: %w", err)
		}

		opts := templates.CallOptions(meta)
		response, err := c.Client.Call(ctx, messages, opts...)
		if err != nil {
			return msg, fmt.Errorf("complete: %w", err)
		}

		msg.Data.Content = response.Text
		return msg, nil
	})
}

type CompleteTemplate struct {
	Echo
	Vars     map[string]any
	Template string
}

func (c CompleteTemplate) Run(ctx *tesei.Thread, in <-chan *tesei.Message[files.TextFile], out chan<- *tesei.Message[files.TextFile]) {
	err := c.init(ctx)
	if err != nil {
		return
	}

	err = c.initTemplatesEngine(ctx)
	if err != nil {
		return
	}

	tesei.Transform(ctx, in, out, func(msg *tesei.Message[files.TextFile]) (*tesei.Message[files.TextFile], error) {
		vars := extend(msg.Metadata, c.Vars, msg)
		messages, meta, err := c.templatesEngine.GenerateWithMetadata(c.Template, vars)
		if err != nil {
			return msg, fmt.Errorf("complete: %w", err)
		}

		opts := templates.CallOptions(meta)
		response, err := c.Client.Call(ctx, messages, opts...)
		if err != nil {
			return msg, fmt.Errorf("complete: %w", err)
		}

		msg.Data.Content = response.Text
		return msg, nil
	})
}

func extend(metadata map[string]any, vars map[string]any, msg *tesei.Message[files.TextFile]) map[string]any {
	out := templates.Extend(metadata, msg.Data.Content)

	for k, v := range vars {
		if s, ok := v.(string); ok {
			out[k] = files.ResolveString(s, msg)
		} else {
			out[k] = v
		}
	}

	return out
}
