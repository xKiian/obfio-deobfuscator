package visitors

import (
	"strings"

	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/generator"
)

type extractDecoder struct {
	ast.NoopVisitor
	DecoderFunc *ast.FunctionLiteral
}

func (v *extractDecoder) VisitFunctionLiteral(n *ast.FunctionLiteral) {
	code := generator.Generate(n)
	if strings.Contains(code, "return decodeURIComponent") {
		v.DecoderFunc = n
		//fmt.Println("[+] Found Decoder Function")
	}
}

func ExtractDecoderFunc(p *ast.Program) {
	f := &extractDecoder{}
	f.V = f

	p.VisitWith(f)
}
