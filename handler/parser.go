package handler

import (
	"github.com/antlr4-go/antlr/v4"
	qasmparser "github.com/itsubaki/qasm/gen/parser"
	"github.com/itsubaki/qasm/parser"
)

func parseCode(code string) (qasmparser.IProgramContext, error) {
	program, err := parser.Parse(code)
	if err != nil {
		return nil, err
	}

	// Inspect the entire AST before execution, including inactive branches and
	// uncalled functions. Comments and string contents are not include statements.
	pending := []antlr.Tree{program}
	for len(pending) > 0 {
		node := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if _, ok := node.(*qasmparser.IncludeStatementContext); ok {
			return nil, ErrIncludeNotAllowed
		}

		pending = append(pending, node.GetChildren()...)
	}

	return program, nil
}
