package compiler

import (
	"github.com/g-hyoga/writing_a_compiler_in_go/src/monkey/ast"
	"github.com/g-hyoga/writing_a_compiler_in_go/src/monkey/code"
	"github.com/g-hyoga/writing_a_compiler_in_go/src/monkey/object"
)

type Compiler struct {
	instructions code.Instructions
	constans     []object.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constans:     []object.Object{},
	}
}

func (c *Compiler) Compile(node ast.Node) error {
	//TODO: Implement
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constans,
	}
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
