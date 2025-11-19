package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mkozhukh/tesei"
)

type TextFile struct {
	Name    string
	Folder  string
	Content string
}

type Source struct {
	Files []TextFile
}

func (s Source) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	defer close(out)
	for _, file := range s.Files {
		select {
		case out <- tesei.NewMessageWithID(file.Name, &file):
		case <-ctx.Done():
			return
		}
	}
}

type ListDir struct {
	Path          string
	Ext           string
	Log           bool
	Limit         int
	Nested        bool
	MaxDepth      int
	FilterFolders func(name, path string) bool
	FilterFiles   func(name, path string) bool
}

func (l ListDir) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	defer close(out)
	l.processDirectory(ctx, l.Path, "", out, 0, 0)
}

func (l ListDir) processDirectory(ctx *tesei.Thread, dirPath, relPath string, out chan<- *tesei.Message[TextFile], level int, count int) int {
	// Check if we've reached max depth
	if l.MaxDepth > 0 && level >= l.MaxDepth {
		return -1
	}

	files, err := os.ReadDir(dirPath)

	if err != nil {
		select {
		case ctx.Error() <- fmt.Errorf("read dir: %w", err):
		case <-ctx.Done():
			return -1
		}
		return -1
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		baseName := file.Name()
		if file.IsDir() {
			if l.Nested {
				if l.FilterFolders != nil && !l.FilterFolders(baseName, filepath.Join(relPath, baseName)) {
					continue
				}
				count = l.processDirectory(ctx, filepath.Join(dirPath, file.Name()), filepath.Join(relPath, file.Name()), out, level+1, count)
				if count < 0 || (l.Limit > 0 && count >= l.Limit) {
					return count
				}
			}
			continue
		}

		if !strings.HasSuffix(file.Name(), l.Ext) {
			continue
		}

		if l.FilterFiles != nil && !l.FilterFiles(baseName, filepath.Join(relPath, baseName)) {
			continue
		}

		textFile := TextFile{
			Name:   baseName,
			Folder: dirPath,
		}

		if l.Log {
			fmt.Println("list:", textFile.Name, textFile.Folder)
		}

		select {
		case out <- tesei.NewMessageWithID(filepath.Join(dirPath, file.Name()), &textFile):
		case <-ctx.Done():
			return -1
		}

		count++
		if l.Limit > 0 && count >= l.Limit {
			return count
		}
	}
	return count
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
	BasePath string
	Folder   string
	DryRun   bool
	Log      bool
}

func (w WriteFile) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
		var target string

		if w.Folder != "" {
			if w.BasePath != "" {
				// Replace base path while preserving nested structure
				relativePath := strings.TrimPrefix(msg.Data.Folder, w.BasePath)
				relativePath = strings.TrimPrefix(relativePath, string(filepath.Separator))
				target = filepath.Join(w.Folder, relativePath, msg.Data.Name)
			} else {
				// Single folder behavior: completely replace folder
				target = filepath.Join(w.Folder, msg.Data.Name)
			}
		} else {
			// Use original folder
			target = filepath.Join(msg.Data.Folder, msg.Data.Name)
		}

		if !w.DryRun {
			targetDir := filepath.Dir(target)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return msg.WithError(err, "create directory"), nil
			}

			err := os.WriteFile(target, []byte(msg.Data.Content), 0644)
			if err != nil {
				return msg.WithError(err, "write file"), nil
			}
		}

		if w.Log {
			fmt.Println("write file:", target)
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

type HashContent struct {
	Key  string
	Size int
}

func (h HashContent) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
		key := h.Key
		if key == "" {
			key = "hash"
		}
		msg.Metadata[key] = hashBase62(msg.Data.Content, h.Size)
		return msg, nil
	})
}

type RenameFile struct {
	Suffix string
	Ext    string
}

func (r RenameFile) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error) {
		ext := ResolveString(r.Ext, msg)
		if ext == "" {
			ext = filepath.Ext(msg.Data.Name)
		}
		suffix := ResolveString(r.Suffix, msg)

		prevExt := filepath.Ext(msg.Data.Name)
		msg.Data.Name = strings.TrimSuffix(msg.Data.Name, prevExt) + suffix + ext
		return msg, nil
	})
}

type Transform struct {
	Handler func(msg *tesei.Message[TextFile]) (*tesei.Message[TextFile], error)
}

func (t Transform) Run(ctx *tesei.Thread, in <-chan *tesei.Message[TextFile], out chan<- *tesei.Message[TextFile]) {
	tesei.Transform(ctx, in, out, t.Handler)
}
