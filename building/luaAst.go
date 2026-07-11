package building

import (
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/ast"
	"github.com/yuin/gopher-lua/parse"
)

type dynamicLuaBlock struct {
	prototype *lua.FunctionProto
	ordinal   int
}

type dynamicLuaDefinition struct {
	blocks []*dynamicLuaBlock
	props  map[string]struct{}
	open   bool
}

func prepareDynamicLua(source string, name string, ordinal int) (*dynamicLuaBlock, map[string]struct{}, bool, error) {
	statements, err := parse.Parse(strings.NewReader(source), name)
	if err != nil {
		return nil, nil, false, err
	}
	analyser := &propAnalyser{props: make(map[string]struct{})}
	if err := analyser.statements(statements); err != nil {
		return nil, nil, false, err
	}
	prototype, err := lua.Compile(statements, name)
	if err != nil {
		return nil, nil, false, err
	}
	return &dynamicLuaBlock{prototype: prototype, ordinal: ordinal}, analyser.props, analyser.open, nil
}

type propAnalyser struct {
	props map[string]struct{}
	open  bool
}

func (a *propAnalyser) statements(statements []ast.Stmt) error {
	for _, statement := range statements {
		if err := a.statement(statement); err != nil {
			return err
		}
	}
	return nil
}

func (a *propAnalyser) statement(statement ast.Stmt) error {
	switch value := statement.(type) {
	case *ast.AssignStmt:
		for _, expression := range value.Lhs {
			if isPropsTarget(expression) {
				return fmt.Errorf("props is read-only")
			}
			if err := a.expression(expression); err != nil {
				return err
			}
		}
		return a.expressions(value.Rhs)
	case *ast.LocalAssignStmt:
		if containsName(value.Names, "props") {
			return fmt.Errorf("props cannot be shadowed")
		}
		return a.expressions(value.Exprs)
	case *ast.FuncCallStmt:
		return a.expression(value.Expr)
	case *ast.DoBlockStmt:
		return a.statements(value.Stmts)
	case *ast.WhileStmt:
		if err := a.expression(value.Condition); err != nil {
			return err
		}
		return a.statements(value.Stmts)
	case *ast.RepeatStmt:
		if err := a.statements(value.Stmts); err != nil {
			return err
		}
		return a.expression(value.Condition)
	case *ast.IfStmt:
		if err := a.expression(value.Condition); err != nil {
			return err
		}
		if err := a.statements(value.Then); err != nil {
			return err
		}
		return a.statements(value.Else)
	case *ast.NumberForStmt:
		if value.Name == "props" {
			return fmt.Errorf("props cannot be shadowed")
		}
		if err := a.expressions([]ast.Expr{value.Init, value.Limit, value.Step}); err != nil {
			return err
		}
		return a.statements(value.Stmts)
	case *ast.GenericForStmt:
		if containsName(value.Names, "props") {
			return fmt.Errorf("props cannot be shadowed")
		}
		if err := a.expressions(value.Exprs); err != nil {
			return err
		}
		return a.statements(value.Stmts)
	case *ast.FuncDefStmt:
		if isPropsTarget(value.Name.Func) {
			return fmt.Errorf("props is read-only")
		}
		if value.Name.Receiver != nil && isPropsTarget(value.Name.Receiver) {
			return fmt.Errorf("props is read-only")
		}
		return a.expression(value.Func)
	case *ast.ReturnStmt:
		return a.expressions(value.Exprs)
	}
	return nil
}

func (a *propAnalyser) expressions(expressions []ast.Expr) error {
	for _, expression := range expressions {
		if expression != nil {
			if err := a.expression(expression); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *propAnalyser) expression(expression ast.Expr) error {
	switch value := expression.(type) {
	case *ast.IdentExpr:
		if value.Value == "props" {
			a.open = true
		}
	case *ast.AttrGetExpr:
		if identifier, ok := value.Object.(*ast.IdentExpr); ok && identifier.Value == "props" {
			if key, ok := value.Key.(*ast.StringExpr); ok {
				a.props[strings.ToLower(key.Value)] = struct{}{}
			} else {
				a.open = true
				return a.expression(value.Key)
			}
			return nil
		}
		if err := a.expression(value.Object); err != nil {
			return err
		}
		return a.expression(value.Key)
	case *ast.TableExpr:
		for _, field := range value.Fields {
			if field.Key != nil {
				if err := a.expression(field.Key); err != nil {
					return err
				}
			}
			if err := a.expression(field.Value); err != nil {
				return err
			}
		}
	case *ast.FuncCallExpr:
		if err := a.expression(value.Func); err != nil {
			return err
		}
		if value.Receiver != nil {
			if err := a.expression(value.Receiver); err != nil {
				return err
			}
		}
		return a.expressions(value.Args)
	case *ast.LogicalOpExpr:
		return a.expressions([]ast.Expr{value.Lhs, value.Rhs})
	case *ast.RelationalOpExpr:
		return a.expressions([]ast.Expr{value.Lhs, value.Rhs})
	case *ast.StringConcatOpExpr:
		return a.expressions([]ast.Expr{value.Lhs, value.Rhs})
	case *ast.ArithmeticOpExpr:
		return a.expressions([]ast.Expr{value.Lhs, value.Rhs})
	case *ast.UnaryMinusOpExpr:
		return a.expression(value.Expr)
	case *ast.UnaryNotOpExpr:
		return a.expression(value.Expr)
	case *ast.UnaryLenOpExpr:
		return a.expression(value.Expr)
	case *ast.FunctionExpr:
		if containsName(value.ParList.Names, "props") {
			return fmt.Errorf("props cannot be shadowed")
		}
		return a.statements(value.Stmts)
	}
	return nil
}

func isPropsTarget(expression ast.Expr) bool {
	if identifier, ok := expression.(*ast.IdentExpr); ok {
		return identifier.Value == "props"
	}
	if attribute, ok := expression.(*ast.AttrGetExpr); ok {
		identifier, ok := attribute.Object.(*ast.IdentExpr)
		if !ok {
			return false
		}
		if identifier.Value == "props" {
			return true
		}
		key, keyIsString := attribute.Key.(*ast.StringExpr)
		return identifier.Value == "_G" && keyIsString && key.Value == "props"
	}
	return false
}

func containsName(names []string, name string) bool {
	for _, candidate := range names {
		if candidate == name {
			return true
		}
	}
	return false
}
