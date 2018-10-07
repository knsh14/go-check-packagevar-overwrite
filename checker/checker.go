package checker

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/knsh14/go-check-packagevar-overwrite/picker"
	"github.com/pkg/errors"
)

const (
	readOnlyComment     = "readonly"
	overwritableComment = "overwritable"
	overwriteComment    = "overwrite"

	defaultFormat           = "overwrite package variables: %s"
	readOnlyCommentedFormat = "overwrite readonly package variables: %s"
	allowedOverwriteFormat  = ""
)

type inspector func(ast.Node) ast.Visitor

func (f inspector) Visit(node ast.Node) ast.Visitor {
	return f(node)
}

// Message contains information about overwrited variables
type Message struct {
	Path  string
	Line  int
	Texts []string
}

func CheckDir(dir string) ([]Message, error) {
	var messages []Message
	fset := token.NewFileSet()
	pkgMap, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse dir %s", dir)
	}
	pkgPath := strings.TrimPrefix(dir, filepath.Join(build.Default.GOPATH, "src")+string(filepath.Separator))
	for _, pkg := range pkgMap {
		for filename, file := range pkg.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.AssignStmt:
					msgs, err := checkAssign(fset, n, file, pkgPath)
					if err != nil {
						return false
					}
					pos := fset.Position(n.Pos())
					messages = append(messages, Message{Path: filepath.Join(pkgPath, filename), Line: pos.Line, Texts: msgs})
				}
				return true
			})
		}
	}
	return messages, nil
}

func CheckPkg(pkg *build.Package) ([]Message, error) {
	var messages []Message
	absPath, err := filepath.Abs(pkg.Dir)
	if err != nil {
		return nil, err
	}
	pkgPath := strings.TrimPrefix(absPath, filepath.Join(build.Default.GOPATH, "src")+string(filepath.Separator))
	for _, f := range pkg.GoFiles {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, filepath.Join(absPath, f), nil, parser.ParseComments)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s", f)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch n := node.(type) {
			case *ast.AssignStmt:
				msgs, err := checkAssign(fset, n, file, pkgPath)
				if err != nil {
					return false
				}
				pos := fset.Position(n.Pos())
				messages = append(messages, Message{Path: filepath.Join(pkgPath, f), Line: pos.Line, Texts: msgs})
			}
			return true
		})
	}
	return messages, nil
}

func checkAssign(fset *token.FileSet, stmt *ast.AssignStmt, f *ast.File, selfImportPath string) ([]string, error) {
	var msgs []string
	cmap := ast.NewCommentMap(fset, f, f.Comments)
	for _, assignee := range stmt.Lhs {
		pkgPath := selfImportPath
		varName := ""
		switch val := assignee.(type) {
		case *ast.SelectorExpr:
			varName = val.Sel.Name
			for _, impt := range f.Imports {
				pkgName := val.X.(*ast.Ident).Name
				pkgPath = strings.Trim(impt.Path.Value, "\"")
				imptName := filepath.Base(pkgPath)
				if impt.Name != nil {
					imptName = impt.Name.Name
				}
				if imptName == pkgName {
					break
				}
			}
		case *ast.Ident:
			varName = val.Name
		}
		comments, err := picker.GetComment(pkgPath, varName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get comment")
		}
		if msg := checkComment(comments, cmap[stmt]); msg != allowedOverwriteFormat {
			pkgName := ""
			if pkgPath != selfImportPath {
				pkgName = pkgPath + "."
			}
			msgs = append(msgs, fmt.Sprintf(msg, pkgName+varName))
		}
	}
	return msgs, nil
}

func checkComment(baseCommentGroup, destCommentGroup []*ast.CommentGroup) string {
	isReadonly := true
	for _, comments := range baseCommentGroup {
		for _, t := range comments.List {
			// ここに来るってことはすでに書き換えているので、readonly があれば問答無用で警告を出す
			if strings.Contains(t.Text, readOnlyComment) {
				return readOnlyCommentedFormat
			}
			if strings.Contains(t.Text, overwritableComment) {
				isReadonly = false
			}
		}
	}
	// デフォルトは readonly の挙動にする
	if isReadonly {
		return defaultFormat
	}
	for _, assignComment := range destCommentGroup {
		for _, ac := range assignComment.List {
			if strings.Contains(ac.Text, overwriteComment) {
				return allowedOverwriteFormat
			}
		}
	}
	return defaultFormat
}
