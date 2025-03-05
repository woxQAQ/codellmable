package main

import (
	_ "embed"
	"flag"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/woxQAQ/llm-code-reader/internal/config"
	"github.com/woxQAQ/llm-code-reader/pkg/fsx"
)

var (
	filterDir  = []string{".cargo", ".git", ".github"}
	filterFile = []string{
		".gitignore",
		".gitmodules",
		".dockerignore",
		"CODE_OF_CONDUCT.md",
		"CONTRIBUTING.md",
		"LICENSE",
		"Makefile",
	}
	filterExt = []string{"toml"}
)

var (
	//go:embed internal/template/src.tpl
	srcTemplateString string
	//go:embed internal/template/tree.tpl
	treeTemplateString string

	srcTemplate  *template.Template
	treeTemplate *template.Template
	homepath     string
	cfg          *config.Config
)

func init() {
	var err error
	srcTemplate, err = template.New("rust_template").
		Parse(srcTemplateString)
	if err != nil {
		log.Fatal(err)
	}
	treeTemplate, err = template.New("tree_template").
		Parse(treeTemplateString)
	if err != nil {
		log.Fatal(err)
	}
	cfg = config.NewConfig("./project.yaml")
	homepath, err = os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range cfg.ExtraExcludePattern {
		filterDir = append(filterDir, p)
	}
	for _, e := range cfg.ExtraExcludePattern {
		filterExt = append(filterExt, e)
	}
	flag.Parse()
}

func main() {
	absSrc, err := filepath.Abs(cfg.Source)
	if err != nil {
		panic(err)
	}
	absOut, err := filepath.Abs(cfg.Target)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(absOut, 0755)
	if err != nil {
		panic(err)
	}

	err = os.Remove(path.Join(absOut, cfg.Project+".md"))
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	f, err := os.Create(path.Join(absOut, cfg.Project+".md"))
	if err != nil && err != os.ErrExist {
		panic(err)
	}

	srcTree, err := fsx.Tree(absSrc, filterDir...)
	if err != nil {
		panic(err)
	}
	err = treeTemplate.Execute(f, map[string]any{
		"project": cfg.Project,
		"content": string(srcTree),
	})
	if err != nil {
		panic(err)
	}
	f.WriteString("\n")

	err = filepath.WalkDir(absSrc, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if slices.Contains(filterFile, d.Name()) {
			return nil
		}

		relRoot, err := filepath.Rel(absSrc, path)
		if err != nil {
			return err
		}
		if d.IsDir() && slices.Contains(filterDir, relRoot) {
			return filepath.SkipDir
		}

		ext := strings.TrimLeft(filepath.Ext(path), ".")

		if ext == "png" || ext == "md" {
			// TODO: add png ref
			return nil
		}

		if ext == "rs" {
			ext = "rust"
		}

		if d.IsDir() {
			return os.MkdirAll(filepath.Join(absSrc, relRoot), 0755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		f.WriteString("\n")

		return srcTemplate.Execute(f, map[string]any{
			"project":  cfg.Project,
			"filePath": relRoot,
			"language": ext,
			"src":      string(content),
		})
	})
	if err != nil {
		panic(err)
	}
}
