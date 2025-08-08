package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mkozhukh/tesei"
)

type TextFile struct {
	Name    string
	Folder  string
	Content string
}

type ListDir struct {
	Path string
	Ext  string
}

func (l ListDir) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	defer close(out)
	files, err := os.ReadDir(l.Path)
	if err != nil {
		select {
		case ctx.Error() <- fmt.Errorf("read dir: %w", err):
		case <-ctx.Done():
			return
		}
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !strings.HasSuffix(file.Name(), l.Ext) {
			continue
		}

		textFile := TextFile{
			Name:   file.Name(),
			Folder: l.Path,
		}

		select {
		case out <- tesei.NewMessageWithID(filepath.Join(l.Path, file.Name()), &textFile):
		case <-ctx.Done():
			return
		}
	}
}

type ReadFile struct{}

func (r ReadFile) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
		data, err := os.ReadFile(filepath.Join(msg.Data.Folder, msg.Data.Name))
		if err != nil {
			return nil, err
		}
		msg.Data.Content = string(data)
		return msg, nil
	})
}

type WriteFile struct {
	DryRun bool
}

func (w WriteFile) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
		target := filepath.Join(msg.Data.Folder, msg.Data.Name)
		if w.DryRun {
			fmt.Println("write file:", target)
			return msg, nil
		}

		err := os.WriteFile(target, []byte(msg.Data.Content), 0644)
		if err != nil {
			return msg.WithError(err, "write file"), nil
		}
		return msg, nil
	})
}

type PrintContent struct{}

func (p PrintContent) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
		fmt.Println(msg.ID)
		fmt.Println(msg.Data.Content)
		return msg, nil
	})
}
