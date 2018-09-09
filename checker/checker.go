package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/knsh14/go-check-overwrites/picker"
	"github.com/pkg/errors"
)

const (
	readOnlyComment     = "readonly"
	overwritableComment = "overwritable"
	overwriteComment    = "overwrite"
)

type inspector func(ast.Node) ast.Visitor

func (f inspector) Visit(node ast.Node) ast.Visitor {
	return f(node)
}

// Check checks package variables are overwrited.
func Check(path string) error {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return errors.Wrapf(err, "failed to parse %s", path)
	}

	cmap := ast.NewCommentMap(fset, f, f.Comments)
	ast.Inspect(f, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.AssignStmt:
			msgs, err := check(n, cmap, f.Imports)
			if err != nil {
				fmt.Println(err)
				return false
			}
			if len(msgs) > 0 {
				for i := range msgs {
					pos := fset.Position(n.Pos())
					fmt.Fprintln(os.Stdout, pos.Filename, pos.Line, msgs[i])
				}
			}
		}
		return true
	})
	return nil
}

func check(stmt *ast.AssignStmt, cmap ast.CommentMap, imports []*ast.ImportSpec) ([]string, error) {
	var messages []string
	for _, assignee := range stmt.Lhs {
		switch val := assignee.(type) {
		case *ast.SelectorExpr:
			for _, impt := range imports {
				pkgName := val.X.(*ast.Ident).Name
				pkgPath := strings.Trim(impt.Path.Value, "\"")
				imptName := filepath.Base(pkgPath)
				if impt.Name != nil {
					imptName = impt.Name.Name
				}
				if imptName == pkgName {
					comments, err := picker.GetComment(pkgPath, val.Sel.Name)
					if err != nil {
						return nil, errors.Wrap(err, "failed to get comment")
					}
					if msg := checkComment(comments, cmap[stmt]); msg != "" {
						messages = append(messages, fmt.Sprintf("%s %s.%s", msg, imptName, val.Sel.Name))
					}
				}
			}
		case *ast.Ident:
			// 自分のインポートパスを取得してやっていく
		}
	}
	return messages, nil
}

func checkComment(baseComment, usedComment []*ast.CommentGroup) string {
	for _, bc := range baseComment {
		for _, t := range bc.List {
			if strings.Contains(t.Text, readOnlyComment) {
				return "you overwrites read only package variables"
			}
			if strings.Contains(t.Text, overwritableComment) {
				for _, assignComment := range usedComment {
					for _, ac := range assignComment.List {
						if strings.Contains(ac.Text, overwriteComment) {
							return ""
						}
					}
				}
			}
		}
	}
	return "you overwrite package variables"
}
