package obfio_deobfuscator

import (
	"github.com/t14raptor/go-fast/ast"
	"github.com/t14raptor/go-fast/transform/simplifier"
	"github.com/xkiian/obfio-deobfuscator/visitors"
)

func Deobfuscate(a *ast.Program) {
	simplifier.Simplify(a, true)
	v := []func(p *ast.Program){
		visitors.ReplaceReassignments,
		visitors.ReplaceStrings,
		visitors.ConcatStrings,
		visitors.UnrollProxyFunctions,
		visitors.SequenceUnroller,
		visitors.SequenceUnroller,
	}

	for _, fn := range v {
		fn(a)
	}
}
