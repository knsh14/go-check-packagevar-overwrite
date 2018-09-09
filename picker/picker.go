package picker

import (
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"

	"github.com/pkg/errors"
)

func GetComment(packageName, valName string) ([]*ast.CommentGroup, error) {
	fset := token.NewFileSet()
	v, err := build.Default.Import(packageName, "", 0)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to import %s", packageName)
	}

	for _, s := range v.GoFiles {
		f, err := parser.ParseFile(fset, v.Dir+"/"+s, nil, parser.ParseComments)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse %s", s)
		}

		for _, decl := range f.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
				for _, spec := range gen.Specs {
					valSpec, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					for _, name := range valSpec.Names {
						if name.Name == valName {
							cmap := ast.NewCommentMap(fset, f, f.Comments)
							return cmap[gen], nil
						}
					}
				}
			}
		}
	}
	return nil, errors.Errorf("not found %s", valName)
}
