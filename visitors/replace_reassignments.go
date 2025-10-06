package visitors

import (
	"github.com/t14raptor/go-fast/ast"
)

type gather struct {
	ast.NoopVisitor
	Decls map[ast.Id]*ast.Identifier
}

type replace struct {
	ast.NoopVisitor
	decls map[ast.Id]*ast.Identifier
}

func (v *gather) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	n.VisitChildrenWith(v)
	if n.Initializer == nil {
		return
	}
	init, ok := n.Initializer.Expr.(*ast.Identifier)
	if !ok {
		return
	}

	target, ok := n.Target.Target.(*ast.Identifier)
	if !ok {
		return
	}
	if target.ToId() == init.ToId() {
		return
	}
	v.Decls[target.ToId()] = init
}

func (v *replace) VisitExpression(n *ast.Expression) {
	n.VisitChildrenWith(v)
	id, ok := n.Expr.(*ast.Identifier)
	if !ok {
		return
	}

	val := v.decls[id.ToId()]
	if val == nil {
		return
	}
	for {
		resolved := v.decls[val.ToId()]
		if resolved == nil {
			break
		}
		val = resolved
	}
	n.Expr = val
}

func ReplaceReassignments(p *ast.Program) {
	f := &gather{
		Decls: make(map[ast.Id]*ast.Identifier),
	}
	f.V = f
	p.VisitWith(f)

	ff := &replace{
		decls: f.Decls,
	}
	ff.V = ff

	p.VisitWith(ff)
}
