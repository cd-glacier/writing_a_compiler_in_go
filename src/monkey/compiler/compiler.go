package compiler

import (
	"fmt"

	"github.com/g-hyoga/writing_a_compiler_in_go/src/monkey/ast"
	"github.com/g-hyoga/writing_a_compiler_in_go/src/monkey/code"
	"github.com/g-hyoga/writing_a_compiler_in_go/src/monkey/object"
)

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions        code.Instructions
	constants           []object.Object
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
	symbolTable         *SymbolTable
}

func New() *Compiler {
	return &Compiler{
		instructions:        code.Instructions{},
		constants:           []object.Object{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
		symbolTable:         NewSymbolTable(),
	}
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstrucion(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) setLastInstrucion(op code.Opcode, pos int) {
	previous := c.lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.previousInstruction = previous
	c.lastInstruction = last
}

func (c *Compiler) lastInstructionIsPop() bool {
	return c.lastInstruction.Opcode == code.OpPop
}

func (c *Compiler) removeLastPop() {
	c.instructions = c.instructions[:c.lastInstruction.Position]
	c.lastInstruction = c.previousInstruction
}

func (c *Compiler) replaceInstructions(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.instructions[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstructions(opPos, newInstruction)
}

func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		return c.compileProgram(node)

	case *ast.ExpressionStatement:
		return c.compileExpressionStatement(node)

	case *ast.InfixExpression:
		return c.compileInfixExpression(node)

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.PrefixExpression:
		return c.compilePrefixExpression(node)

	case *ast.IfExpression:
		return c.compileIfExpression(node)

	case *ast.BlockStatement:
		return c.compileBlockStatement(node)

	case *ast.LetStatement:
		return c.compileLetStatement(node)

	case *ast.Identifier:
		return c.compileIdentifier(node)

	}

	return nil
}

func (c *Compiler) compileProgram(node *ast.Program) error {
	for _, s := range node.Statements {
		err := c.Compile(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileExpressionStatement(node *ast.ExpressionStatement) error {
	err := c.Compile(node.Expression)
	if err != nil {
		return err
	}
	c.emit(code.OpPop)
	return nil
}

func (c *Compiler) compileInfixExpression(node *ast.InfixExpression) error {
	if node.Operator == "<" {
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		err = c.Compile(node.Left)
		if err != nil {
			return err
		}

		c.emit(code.OpGreaterThan)
		return nil
	}
	err := c.Compile(node.Left)
	if err != nil {
		return err
	}

	err = c.Compile(node.Right)
	if err != nil {
		return err
	}

	switch node.Operator {
	case "+":
		c.emit(code.OpAdd)
	case "-":
		c.emit(code.OpSub)
	case "*":
		c.emit(code.OpMul)
	case "/":
		c.emit(code.OpDiv)
	case ">":
		c.emit(code.OpGreaterThan)
	case "==":
		c.emit(code.OpEqual)
	case "!=":
		c.emit(code.OpNotEqual)
	default:
		return fmt.Errorf("unknown operator %s", node.Operator)
	}

	return nil
}

func (c *Compiler) compilePrefixExpression(node *ast.PrefixExpression) error {
	err := c.Compile(node.Right)
	if err != nil {
		return nil
	}

	switch node.Operator {
	case "!":
		c.emit(code.OpBang)
	case "-":
		c.emit(code.OpMinus)
	default:
		return fmt.Errorf("unknown operator %s", node.Operator)
	}

	return nil
}

func (c *Compiler) compileIfExpression(node *ast.IfExpression) error {
	err := c.Compile(node.Condition)
	if err != nil {
		return err
	}

	// Emit an `OpJumpNotTruthy` with a bogus value
	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

	err = c.Compile(node.Consequence)
	if err != nil {
		return err
	}

	if c.lastInstructionIsPop() {
		c.removeLastPop()
	}

	// Emit an `OpJump` with a bogus value
	jumpPos := c.emit(code.OpJump, 9999)

	afterConsequencePos := len(c.instructions)
	c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

	if node.Alternative == nil {
		c.emit(code.OpNull)
	} else {
		err := c.Compile(node.Alternative)
		if err != nil {
			return err
		}

		if c.lastInstructionIsPop() {
			c.removeLastPop()
		}

	}

	afterAlternativePos := len(c.instructions)
	c.changeOperand(jumpPos, afterAlternativePos)

	return nil
}

func (c *Compiler) compileBlockStatement(node *ast.BlockStatement) error {
	for _, s := range node.Statements {
		err := c.Compile(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) compileLetStatement(node *ast.LetStatement) error {
	err := c.Compile(node.Value)
	if err != nil {
		return err
	}
	symbol := c.symbolTable.Define(node.Name.Value)
	c.emit(code.OpSetGlobal, symbol.Index)

	return nil
}

func (c *Compiler) compileIdentifier(node *ast.Identifier) error {
	symbol, ok := c.symbolTable.Resolve(node.Value)
	if !ok {
		return fmt.Errorf("undefined variable %s", node.Value)
	}
	c.emit(code.OpGetGlobal, symbol.Index)

	return nil
}
