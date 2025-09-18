package pygen

import (
	"go/ast"
	"github.com/goplus/gogen"
	"go/token"
)



func (ctx *context) genVar(pkg *gogen.Package, sym *symbol, goName string) {
	name := sym.Name
	if len(name) == 0 || name[0] == '_' {
		return
	}
	def := pkg.NewVarDefs(pkg.Types.Scope())
	// linkname
	docList := make([]*ast.Comment, 0, 2)
	goLinkname := "//go:linkname " + goName + " py." + sym.Name
	docList = append(docList, &ast.Comment{Text: goLinkname})
	def.SetComments(&ast.CommentGroup{List: docList})
	// new
	def.New(token.NoPos, ctx.objPtr, goName)
}





