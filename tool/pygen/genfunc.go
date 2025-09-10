package pygen

import (
	"go/token"
	"go/ast"
	"go/types"
	"github.com/goplus/gogen"
)


// generate global function
func (ctx *context) genFunc(pkg *gogen.Package, sym *symbol) {
	if len(sym.Name) == 0 || sym.Name[0] == '_' {
		return  // skip private functions
	}
	if sym.Sig == "" {
		ctx.skips = append(ctx.skips, *sym)
		return
	}
	// signature
	params, variadic := ctx.genParams(pkg, sym.Sig, false, false)
	name := genName(sym.Name, -1)
	sig := types.NewSignatureType(nil, nil, nil, params, ctx.ret, variadic) // ret: *py.Object
	fn := pkg.NewFuncDecl(token.NoPos, name, sig)
	// doc
	docList := ctx.genDoc(sym.Doc)
	if len(docList) > 0 {
		docList = append(docList, emptyCommentLine)
	}
	docList = append(docList, ctx.genLinkname(name, sym))
	fn.SetComments(pkg, &ast.CommentGroup{List: docList})
}


func (ctx *context) genLinkname(name string, sym *symbol) *ast.Comment {
	return &ast.Comment{Text: "//go:linkname " + name + " py." + sym.Name}
}