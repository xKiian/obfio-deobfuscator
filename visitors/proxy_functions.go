package visitors

import (
	"github.com/t14raptor/go-fast/ast"
)

type proxySimplify struct {
	ast.NoopVisitor
	proxy map[ast.Id]map[string]any
}

type proxySimplifier struct {
	ast.NoopVisitor
	proxy map[ast.Id]map[string]any
}

func (v *proxySimplify) VisitExpression(n *ast.Expression) {
	n.VisitChildrenWith(v)

	switch n.Expr.(type) {
	case *ast.MemberExpression:
		compProp, ok := n.Expr.(*ast.MemberExpression).Property.Prop.(*ast.ComputedProperty)
		if !ok {
			return
		}

		strLit, ok := compProp.Expr.Expr.(*ast.StringLiteral)
		if !ok {
			return
		}

		if len(strLit.Value) != 5 {
			return
		}

		n.Expr.(*ast.MemberExpression).Property.Prop = &ast.Identifier{
			Name: strLit.Value,
		}
	}

}

func (v *proxySimplify) VisitVariableDeclarator(n *ast.VariableDeclarator) {
	n.VisitChildrenWith(v)

	if n.Initializer == nil {
		return
	}
	obj, ok := n.Initializer.Expr.(*ast.ObjectLiteral)
	if !ok {
		return
	}

	target, ok := n.Target.Target.(*ast.Identifier)
	if !ok {
		return
	}
	id := target.ToId()
	if _, ok := v.proxy[id]; !ok {
		v.proxy[id] = map[string]any{}
	}

	for _, val := range obj.Value {
		key, ok := val.Prop.(*ast.PropertyKeyed)
		if !ok {
			continue
		}

		strLit, ok := key.Key.Expr.(*ast.StringLiteral)
		if !ok {
			continue
		}

		switch key.Value.Expr.(type) {
		case *ast.StringLiteral:
			v.proxy[id][strLit.Value] = key.Value.Expr.(*ast.StringLiteral).Value

		case *ast.FunctionLiteral:
			v.proxy[id][strLit.Value] = key.Value.Expr
		}
	}
}

func (v *proxySimplifier) VisitExpression(n *ast.Expression) {
	n.VisitChildrenWith(v)
	memExpr, ok := n.Expr.(*ast.MemberExpression)
	if !ok {
		return
	}
	obj, ok := memExpr.Object.Expr.(*ast.Identifier)
	if !ok {
		return
	}
	id := obj.ToId()

	switch memExpr.Property.Prop.(type) {
	case *ast.Identifier:
		idProp, ok := memExpr.Property.Prop.(*ast.Identifier)
		if !ok {
			return
		}
		if _, ok := v.proxy[id]; !ok {
			return
		}
		val := v.proxy[id][idProp.Name]
		if val == nil {
			return
		}
		strLit, ok := val.(string)
		if !ok {
			return
		}

		n.Expr = &ast.StringLiteral{Value: strLit}
	}
	//TODO do function unwrap too

}

func UnrollProxyFunctions(p *ast.Program) {
	f := &proxySimplify{
		proxy: map[ast.Id]map[string]any{},
	}
	f.V = f
	p.VisitWith(f)

	f2 := &proxySimplifier{
		proxy: f.proxy,
	}
	f2.V = f2
	p.VisitWith(f2)
}
