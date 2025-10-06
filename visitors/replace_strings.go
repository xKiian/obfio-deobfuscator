package visitors

import (
	"github.com/t14raptor/go-fast/ast"
)

type stringReplacerGather struct {
	ast.NoopVisitor
	EndNum      float64
	ShuffleExpr *ast.BinaryExpression
}

func (v *stringReplacerGather) VisitExpressionStatement(n *ast.ExpressionStatement) {
	n.VisitChildrenWith(v)

	cexpr, ok := n.Expression.Expr.(*ast.CallExpression)
	if !ok || len(cexpr.ArgumentList) != 2 {
		return
	}
	if _, ok = cexpr.ArgumentList[0].Expr.(*ast.Identifier); !ok {
		return
	}
	num, ok := cexpr.ArgumentList[1].Expr.(*ast.NumberLiteral)
	if !ok {
		return
	}
	v.EndNum = num.Value

	callee, ok := cexpr.Callee.Expr.(*ast.FunctionLiteral)
	if !ok {
		return
	}
	if len(callee.Body.List) < 2 {
		return
	}
	body, ok := callee.Body.List[1].Stmt.(*ast.WhileStatement)
	if !ok {
		return
	}
	block, ok := body.Body.Stmt.(*ast.BlockStatement)
	if !ok {
		return
	}
	if len(block.List) != 1 {
		return
	}

	try, ok := block.List[0].Stmt.(*ast.TryStatement)
	if !ok {
		return
	}
	if len(try.Body.List) != 2 {
		return
	}

	varDecl, ok := try.Body.List[0].Stmt.(*ast.VariableDeclaration)
	if !ok {
		return
	}
	varDeclor, ok := varDecl.List[0].Initializer.Expr.(*ast.BinaryExpression)
	if !ok {
		return
	}
	v.ShuffleExpr = varDeclor
}

func ReplaceStrings(p *ast.Program) {
	f := &stringReplacerGather{}
	f.V = f
	p.VisitWith(f)
}
