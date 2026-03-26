package service

import (
	"sort"
	"text/template/parse"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/repository"
)

// TemplateService manages notification templates with Go template syntax.
type TemplateService struct {
	repo *repository.TemplateRepository
}

func NewTemplateService(repo *repository.TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

func (s *TemplateService) Create(t *model.Template) error {
	t.Variables = extractVariables(t.Subject, t.Body)
	return s.repo.Create(t)
}

func (s *TemplateService) GetByID(id uint) (*model.Template, error) {
	return s.repo.FindByID(id)
}

func (s *TemplateService) List(userID uint) ([]model.Template, error) {
	return s.repo.FindAll(userID)
}

// extractVariables parses Go template strings and returns sorted variable names (e.g. {{.Name}} -> "Name").
func extractVariables(texts ...string) []string {
	seen := make(map[string]struct{})
	for _, text := range texts {
		tree, err := parse.Parse("", text, "{{", "}}")
		if err != nil {
			continue
		}
		for _, t := range tree {
			walkNodes(t.Root, seen)
		}
	}

	vars := make([]string, 0, len(seen))
	for v := range seen {
		vars = append(vars, v)
	}
	sort.Strings(vars)
	return vars
}

// walkNodes recursively traverses a template AST to collect field references.
func walkNodes(node parse.Node, seen map[string]struct{}) {
	if node == nil {
		return
	}
	switch n := node.(type) {
	case *parse.FieldNode:
		if len(n.Ident) > 0 {
			seen[n.Ident[0]] = struct{}{}
		}
	case *parse.ListNode:
		if n != nil {
			for _, child := range n.Nodes {
				walkNodes(child, seen)
			}
		}
	case *parse.ActionNode:
		walkPipe(n.Pipe, seen)
	case *parse.IfNode:
		walkNodes(n.List, seen)
		walkNodes(n.ElseList, seen)
		walkPipe(n.Pipe, seen)
	case *parse.RangeNode:
		walkNodes(n.List, seen)
		walkNodes(n.ElseList, seen)
		walkPipe(n.Pipe, seen)
	}
}

func walkPipe(pipe *parse.PipeNode, seen map[string]struct{}) {
	if pipe == nil {
		return
	}
	for _, cmd := range pipe.Cmds {
		for _, arg := range cmd.Args {
			walkNodes(arg, seen)
		}
	}
}
