package visitors

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/generator"
	"github.com/xkiian/obfio-deobfuscator/utils"
)

type stringReplacerGather struct {
	ast.NoopVisitor
	stopValue   float64
	ShuffleExpr *ast.Expression
	DecoderFunc ast.Id
	offset      int
	stringArray []string
}

type stringReplacer struct {
	ast.NoopVisitor
	DecoderFunc ast.Id
	decoder     *utils.Rc4StringDecoder
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
	v.stopValue = num.Value

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

	_, ok = varDecl.List[0].Initializer.Expr.(*ast.BinaryExpression)
	if !ok {
		return
	}

	v.ShuffleExpr = varDecl.List[0].Initializer
	//fmt.Println(generator.Generate(varDeclor))
}

func (v *stringReplacerGather) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	n.VisitChildrenWith(v)
	code := generator.Generate(n)
	if !strings.Contains(code, "return decodeURIComponent(") {
		return
	}

	if len(n.Body.List) != 2 {
		return
	}
	returnStmt, ok := n.Body.List[1].Stmt.(*ast.ReturnStatement)
	if !ok {
		return
	}

	seqExpr, ok := returnStmt.Argument.Expr.(*ast.SequenceExpression)
	if !ok {
		return
	}

	aExpr, ok := seqExpr.Sequence[0].Expr.(*ast.AssignExpression)
	if !ok {
		return
	}

	fLit, ok := aExpr.Right.Expr.(*ast.FunctionLiteral)
	if !ok {
		return
	}
	if len(fLit.Body.List) < 2 {
		return
	}

	ExprStmt, ok := fLit.Body.List[0].Stmt.(*ast.ExpressionStatement)
	if !ok {
		return
	}

	aExpr, ok = ExprStmt.Expression.Expr.(*ast.AssignExpression)
	if !ok {
		return
	}

	right, ok := aExpr.Right.Expr.(*ast.BinaryExpression)
	if !ok {
		return
	}

	val := int(right.Right.Expr.(*ast.NumberLiteral).Value)

	op := right.Operator.String()
	switch op {
	case "+":
		v.offset = val
	case "-":
		v.offset = -val
	default:
		panic("unsupported array offset | op: " + op)
	}

}

func (v *stringReplacerGather) VisitVariableDeclaration(n *ast.VariableDeclaration) {
	n.VisitChildrenWith(v)
	if len(n.List) != 1 {
		return
	}
	if n.List[0].Initializer == nil {
		return
	}
	varDecl, ok := n.List[0].Initializer.Expr.(*ast.ArrayLiteral)
	if !ok {
		return
	}
	var values []string
	for _, val := range varDecl.Value {
		strLit, ok := val.Expr.(*ast.StringLiteral)
		if !ok {
			return
		}
		values = append(values, strLit.Value)
	}

	v.stringArray = values
}

func (v *stringReplacerGather) VisitFunctionDeclaration(n *ast.FunctionDeclaration) {
	n.VisitChildrenWith(v)
	code := generator.Generate(n)
	if !strings.Contains(code, "return decodeURIComponent(") {
		return
	}
	v.DecoderFunc = n.Function.Name.ToId()
}

func (v *stringReplacer) VisitExpression(n *ast.Expression) {
	n.VisitChildrenWith(v)

	callExpr, ok := n.Expr.(*ast.CallExpression)
	if !ok {
		return
	}

	callee, ok := callExpr.Callee.Expr.(*ast.Identifier)
	if !ok {
		return
	}
	if callee.ToId() != v.DecoderFunc {
		return
	}

	indexLit, ok := callExpr.ArgumentList[0].Expr.(*ast.NumberLiteral)
	if !ok {
		return
	}
	index := int(indexLit.Value)

	keyLit, ok := callExpr.ArgumentList[1].Expr.(*ast.StringLiteral)
	if !ok {
		return
	}
	key := keyLit.Value

	decoded := v.decoder.Get(index, key)

	n.Expr = &ast.StringLiteral{Value: decoded}
}

var normalRe = regexp.MustCompile(`^[0-9][a-zA-Z0-9+\-*/%()=<>!&|^.,\s]*$`)
var shuffleCheckerRe = regexp.MustCompile(`parseInt\s*\(\s*.\s*\(\s*(0x[0-9a-fA-F]+)\s*,\s*(['"])(.*?)\s*'\)\s*\)`)

type Entry struct {
	index int
	key   string
}

func ReplaceStrings(p *ast.Program) {
	f := &stringReplacerGather{}
	f.V = f
	p.VisitWith(f)

	decoder := utils.NewRc4StringDecoder(f.stringArray, f.offset)
	if f.ShuffleExpr == nil {
		return
	}
	matches := shuffleCheckerRe.FindAllStringSubmatch(generator.Generate(f.ShuffleExpr), -1)
	var out []Entry
	for _, m := range matches {
		hexStr := m[1]
		key := m[3]
		val, err := strconv.ParseInt(hexStr, 0, 64)
		if err != nil {
			fmt.Fprintln(os.Stderr, "parse error:", err)
			continue
		}
		out = append(out, Entry{
			index: int(val),
			key:   key,
		})
	}
outer:
	for {
		for _, entry := range out {
			text := decoder.Get(entry.index, entry.key)
			if !normalRe.MatchString(text) {
				decoder.Shift()
				continue outer
			}
		}
		break
	}

	f2 := &stringReplacer{
		DecoderFunc: f.DecoderFunc,
		decoder:     decoder,
	}
	f2.V = f2
	p.VisitWith(f2)
}
