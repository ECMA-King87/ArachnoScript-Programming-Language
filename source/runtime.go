package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
)

//#region Memory

// key / value map for storing the reference and values of variables
var Memory *Map[string, RuntimeVal] = NewMap[string, RuntimeVal]()

var (
	undefined = MK_UD()
	null      = MK_NULL()
)

// #endregion
// #region Interpreter
type Interpreter struct {
	returned_from_function bool
	terminated             bool
	_break                 bool
	_continue              bool
	MemThreshold           uint64
	CallStack              *Stack
	microTaskQueue         *MicroTaskQueue
	exports                *Map[RuntimeVal, string]
}

//#region Methods

func (r *Interpreter) Evaluate(node Node, env *Environment) RuntimeVal {
	switch node := node.(type) {
	// Literals
	case *Number:
		return &NumberVal{node.Value}
	case *String:
		return &StringVal{node.Value}
	case *ObjectLiteral:
		return r.Eval_object(node, env)
	case *ArrayLiteral:
		return r.Eval_array(node, env)
	// Expressions
	case *BinaryExpr:
		return r.Eval_binary_expr(node, env)
	case *Identifier:
		return r.Eval_identifier(node, env)
	case *AssignmentExpr:
		return r.Eval_assignment(node, env)
	case *ComparisonExpr:
		return r.Eval_comparison_expr(node, env)
	case *TypeOfExpr:
		return r.Eval_typeof(node, env)
	case *VoidExpr:
		return r.Eval_void_expr(node, env)
	case *MemberExpr:
		return r.Eval_member_expr(node, env)
	case *InExpr:
		return r.Eval_in_expr(node, env)
	case *IncrementExpr:
		return r.Eval_increment_expr(node, env)
	case *GroupingExpr:
		return r.Eval_grouping_expr(node, env)
	case *globalThis:
		return r.Eval_globalThis(node, env)
	case *globalThisMember:
		return r.Eval_globalThisMember(node, env)
	case *globalThisMemberAssignment:
		return r.Eval_globalThisMemberAssignment(node, env)
	case *CallExpr:
		return r.Eval_call_expr(node, env)
	case *AwaitExpr:
		return r.Eval_await_expr(node, env)
	case *NewExpr:
		return r.Eval_new_expr(node, env)
	case *SuperExpr:
		return r.Eval_super_expr(node, env)
	case *LogicalExpr:
		return r.Eval_logical_expr(node, env)
	case *FromExpr:
		return r.Eval_from_expr(node, env)
	case *DynamicImport:
		return r.Eval_dynamic_import(node, env)
	case *MatchExpr:
		return r.Eval_match_expr(node, env)
	case *InstanceofExpr:
		return r.Eval_instanceof_expr(node, env)
	case *TernaryExpr:
		return r.Eval_ternary_expr(node, env)
	// Statements
	case *Program:
		return r.EvalProgram(node, env)
	case *VarDecl:
		return r.EvalVarDecl(node, env)
	case *IfStmt:
		return r.EvalIfStmt(node, env)
	case *WhileLoop:
		return r.EvalWhileLoop(node, env)
	case *ThrowStmt:
		return r.EvalThrowStmt(node, env)
	case *TryCatch:
		return r.EvalTryStmt(node, env)
	case *BlockStmt:
		return r.EvalBlockStmt(node, env)
	case *DeleteStmt:
		return r.EvalDeleteStmt(node, env)
	case *ForLoop:
		return r.EvalForLoop(node, env)
	case *ForIteratorLoop:
		return r.EvalForIterLoop(node, env)
	case *FunctionDecl:
		decl, _ := r.EvalFunctionDecl(node, env)
		return decl
	case *ReturnStmt:
		return r.EvalReturnStmt(node, env)
	case *BreakStmt:
		return r.EvalBreakStmt(node, env)
	case *ContinueStmt:
		return r.EvalContinueStmt(node, env)
	case *Label:
		return undefined
	case *ClassDecl:
		decl, _ := r.EvalClassDecl(node, env)
		return decl
	case *ImportStmt:
		return r.EvalImportStmt(node, env)
	case *ExportStmt:
		return r.EvalExportStmt(node, env)
	case *SwitchStmt:
		return r.EvalSwitchStmt(node, env)
	default:
		throwMessage(fmt.Sprintf("This AST node has not yet been setup for interpretation: %T", node))
	}
	return nil
}

//#region Statements

func (r *Interpreter) EvalProgram(p *Program, env *Environment) *ObjectVal {
	ud_ref = GetUDRef(r)
	// sync code
	r.EvalBlock(p.body, env)
	// micro tasks
	for r.microTaskQueue.length > 0 {
		r.microTaskQueue.execCurrentTask()
	}
	return MK_OBJECT(r.exports, env, r)
}

func (r *Interpreter) EvalBlock(body []Node, env *Environment) RuntimeVal {
	var lastEval RuntimeVal = undefined
	for i := 0; i < len(body); i++ {
		if r.terminated {
			return lastEval
		}
		stmt := body[i]
		lastEval = r.Evaluate(stmt, env)
	}
	return lastEval
}

func (r *Interpreter) EvalVarDecl(decl *VarDecl, env *Environment) RuntimeVal {
	rhs := r.Evaluate(decl.right, env)
	r.DeclareVar(decl, rhs, env)
	return undefined
}

func (r *Interpreter) DeclareVar(decl *VarDecl, rhs RuntimeVal, env *Environment) *Map[string, string] {
	// key: ident, value: ref
	decls := NewMap[string, string]()
	switch left := decl.left.(type) {
	case *Identifier:
		ref, _ := env.DeclareVarRef(
			left.Symbol,
			rhs,
			decl._type,
			decl.line,
			decl.col,
			decl.count,
			env.sourcePath,
			r,
		)
		decls.set(left.Symbol, ref)
	case *ObjectLiteral:
		decls.copy(DestructureObjectDecl(rhs, left, decl._type, env, r))
	case *ArrayLiteral:
		decls.copy(DestructureArrayDecl(rhs, left, decl._type, env, r))
	default:
		panic("unimplemented")
	}
	return decls
}

func (r *Interpreter) EvalIfStmt(stmt *IfStmt, env *Environment) RuntimeVal {
	decl := false
	var condition RuntimeVal
	switch c := stmt.condition.(type) {
	case *VarDecl:
		condition = r.Evaluate(c.right, env)
		decl = true
	default:
		condition = r.Evaluate(c, env)
	}
	block_scope := NewEnv(env, "block", env.sourcePath)
	var lastEval RuntimeVal
	if RtvToBool(condition) {
		if decl {
			decl := stmt.condition.(*VarDecl)
			lastEval = r.EvalVarDecl(decl, block_scope)
			// [TODO]: Handle Decls (if (spawn a = true) {})
		}
		lastEval = r.EvalBlock(stmt.body, block_scope)
	} else if stmt.elseBody != nil {
		lastEval = r.EvalBlock(stmt.elseBody, block_scope)
	}
	return lastEval
}

func (r *Interpreter) EvalWhileLoop(stmt *WhileLoop, env *Environment) RuntimeVal {
	do := stmt.do
	var condition RuntimeVal
	if do {
	start:
		if r._break {
			r._break = false
			r.terminated = false
			return undefined
		}
		if r._continue {
			r._continue = false
			r.terminated = false
			return undefined
		}
		scope := NewEnv(env, "loop", env.sourcePath)
		r.EvalBlock(stmt.body, scope)
		// before condition check
		condition = r.Evaluate(stmt.condition, env)
		if RtvToBool(condition) {
			goto start
		}
	} else {
		condition = r.Evaluate(stmt.condition, env)
		for RtvToBool(condition) {
			if r._break {
				r.terminated = false
				r._break = false
				return undefined
			}
			if r._continue {
				r._continue = false
				r.terminated = false
				return undefined
			}
			scope := NewEnv(env, "loop", env.sourcePath)
			r.EvalBlock(stmt.body, scope)
			// keep at end of loop
			condition = r.Evaluate(stmt.condition, env)
		}
	}
	return undefined
}

func (r *Interpreter) EvalThrowStmt(stmt *ThrowStmt, env *Environment) RuntimeVal {
	value := r.Evaluate(stmt.value, env)
	r.terminated = true
	env.throwValue(value, r)
	return undefined
}

func (r *Interpreter) EvalTryStmt(stmt *TryCatch, env *Environment) RuntimeVal {
	try_block := NewEnv(env, "try", env.sourcePath)
	try_block.catch_block = stmt.catch
	// always identifier
	try_block.catch_param = stmt.catch_param
	// try_block.finally_block = node.finally
	// let's try
	r.EvalBlock(stmt.try, try_block) // if theres an error, it will be caught
	// finally
	finally_block := NewEnv(env, "block", env.sourcePath)
	r.EvalBlock(stmt.finally, finally_block)
	return undefined
}

func (r *Interpreter) EvalBlockStmt(stmt *BlockStmt, env *Environment) RuntimeVal {
	block := NewEnv(env, "block", env.sourcePath)
	lastEval := r.EvalBlock(stmt.body, block)
	return lastEval
}

func (r *Interpreter) EvalDeleteStmt(stmt *DeleteStmt, env *Environment) RuntimeVal {
	switch operand := stmt.operand.(type) {
	case *Identifier:
		// memory location
		ml := env.DeleteVar(operand.Symbol, operand.line, operand.col, operand.count, env.sourcePath, r)
		Memory.delete(ml)
	case *MemberExpr:
		ml := r.Get_Member(operand, env)
		Memory.delete(ml)
	default:
		pos := getPosFromNode(operand)
		env.ThrowSyntaxError(
			"the operand of the \"delete\" keyword must be a variable or property access",
			SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""),
		)
		// fmt.Printf("Unhandled delete expression: %T", node)
	}
	return undefined
}

func (r *Interpreter) EvalForLoop(stmt *ForLoop, env *Environment) RuntimeVal {
	scope := NewEnv(env, "loop", env.sourcePath)
	switch expr := (stmt.before).(type) {
	case *AssignmentExpr:
		if expr.op == "=" {
			var decl *VarDecl = &VarDecl{
				left:  expr.left,
				right: expr.right,
				_type: "mutable",
				Pos:   expr.Pos,
			}
			r.Evaluate(decl, scope)
		} else {
			// up the same scope
			r.Evaluate(expr, env)
		}
	default:
		// up the same scope
		r.Evaluate(expr, env)
	}
start:
	loop := NewEnv(scope, "loop", env.sourcePath)
	condition := r.Evaluate(stmt.condition, loop)
	if RtvToBool(condition) {
		if r._break {
			r.terminated = false
			r._break = false
			return undefined
		}
		if r._continue {
			r._continue = false
			r.terminated = false
			return undefined
		}
		r.EvalBlock(stmt.body, loop)
		r.Evaluate(stmt.after, loop)
		goto start
	}
	return undefined
}

func (r *Interpreter) EvalForIterLoop(stmt *ForIteratorLoop, env *Environment) RuntimeVal {
	value := r.Evaluate(stmt.right, env)
	pos := stmt.Pos
	var iterable []RuntimeVal
	if stmt.op == "in" {
		// check if value is iterable
		switch v := value.(type) {
		case *ObjectVal:
			v.properties.forEach(func(key RuntimeVal, _ string) {
				iterable = append(iterable, key)
			})
		case *Instance:
			v.properties.forEach(func(key RuntimeVal, _ string) {
				iterable = append(iterable, key)
			})
		case *ArrayVal:
			v.forEach(func(key int, _ RuntimeVal) {
				iterable = append(iterable, MK_NUMBER(float64(key)))
			})
		default:
			env.ThrowTypeError("type", ValueType(value), "is not iterable in for..in loop"+SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
		}
	} else {
		// for..of
		// check if value is iterable
		switch v := value.(type) {
		case *ObjectVal:
			v.properties.forEach(func(_ RuntimeVal, ml string) {
				iterable = append(iterable, Memory.get(ml))
			})
		case *ArrayVal:
			v.forEach(func(_ int, value RuntimeVal) {
				iterable = append(iterable, value)
			})
		case *StringVal:
			for i := 0; i < len(v.value); i++ {
				iterable = append(iterable, MK_STRING(string(v.value[i])))
			}
		case *Instance:
			sym := MK_STRING(symbol_table.get("iterator").noAnsi())
			if sym == nil {
				sym = MK_STRING(MK_SYMBOL("iterator").noAnsi())
				symbol_table.set("iterator", MK_SYMBOL(sym.value))
			}
			if proto, ok := v.prototype.(*ObjectVal); ok {
				proto_ml := GetPropMlFromProto(sym, proto)
				if len(proto_ml) > 0 {
					method, ok := Memory.get(proto_ml).(*FunctionVal)
					if ok {
						pos := getPosFromNode(stmt.right)
						iter, fn_scope := CallFunction(method, v.class_body, []RuntimeVal{}, v.r, pos)
						ml := ""
						switch i := iter.(type) {
						case *Instance:
							ml = GetPropMlFromProto(MK_STRING("next"), i.prototype)
							if len(ml) == 0 {
								goto err
							}
						case *ObjectVal:
							ml = i.properties.get(MK_STRING("next"))
							if len(ml) == 0 {
								goto err
							}
						default:
							goto err
						}
						next := Memory.get(ml).(*FunctionVal)
						stop := false
						for !stop {
							v := next.Call(fn_scope, []RuntimeVal{}, r, pos)
							if obj, ok := v.(*ObjectVal); ok {
								done := obj.properties.get(MK_STRING("done"))
								if len(done) > 0 {
									done := Memory.get(done)
									stop = RtvToBool(done)
									if !stop {
										value := obj.properties.get(MK_STRING("value"))
										var v RuntimeVal = undefined
										if len(value) > 0 {
											v = Memory.get(value)
										}
										iterable = append(iterable, v)
									}
								}
							} else {
								goto err
							}
						}
						goto start
					}
				}
			}
			if len(v.properties.get(sym)) > 0 {
				method, ok := Memory.get(v.properties.get(sym)).(*FunctionVal)
				if ok {
					pos := getPosFromNode(stmt.right)
					iter, fn_scope := CallFunction(method, v.class_body, []RuntimeVal{}, v.r, pos)
					ml := ""
					switch i := iter.(type) {
					case *Instance:
						ml = GetPropMlFromProto(MK_STRING("next"), i.prototype)
						if len(ml) == 0 {
							goto err
						}
					case *ObjectVal:
						ml = i.properties.get(MK_STRING("next"))
						if len(ml) == 0 {
							goto err
						}
					default:
						goto err
					}
					next := Memory.get(ml).(*FunctionVal)
					stop := false
					for !stop {
						v := next.Call(fn_scope, []RuntimeVal{}, r, pos)
						if obj, ok := v.(*ObjectVal); ok {
							done := obj.properties.get(MK_STRING("done"))
							if len(done) > 0 {
								done := Memory.get(done)
								stop = RtvToBool(done)
								if !stop {
									value := obj.properties.get(MK_STRING("value"))
									var v RuntimeVal = undefined
									if len(value) > 0 {
										v = Memory.get(value)
									}
									iterable = append(iterable, v)
								}
							}
						} else {
							goto err
						}
					}
					goto start
				}
			}
		err:
			env.ThrowTypeError("an instance must have a Symbol.iterator method that returns an iterator: for..in loop" + SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
		default:
			env.ThrowTypeError("type", ValueType(value), "is not iterable in for..in loop"+SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
		}
	}
start:
	for i := 0; i < len(iterable); i++ {
		v := iterable[i]
		if r._break {
			r._break = false
			r.terminated = false
			return undefined
		}
		if r._continue {
			r._continue = false
			r.terminated = false
			return undefined
		}
		scope := NewEnv(env, "loop", env.sourcePath)
		r.DeclareVar(&VarDecl{
			left:  stmt.left,
			right: nil,
			_type: stmt._type,
			Pos:   pos,
		}, v, scope)
		r.EvalBlock(stmt.body, scope)
	}
	return undefined
}

func (r *Interpreter) EvalFunctionDecl(decl *FunctionDecl, env *Environment) (*FunctionVal, string) {
	name := ""
	if decl.name.dynamic {
		name = r.Evaluate(decl.name.node, env).noAnsi()
	} else if !decl.anonymous {
		// always identifier
		name = decl.name.node.(*Identifier).Symbol
	}
	fn := MK_FUNCTION(name, decl.body, decl.params, env, decl.async, decl.anonymous, decl._type == "arrow", r)
	if decl.anonymous && len(fn.name) == 0 {
		fn.name = "(anonymous)"
		ml := GenerateRadix(16)
		Memory.set(ml, fn)
		return fn, ml
	}
	ml, value := env.DeclareVarRef(name, fn, "constant", decl.line, decl.col, decl.count, env.sourcePath, r)
	return value.(*FunctionVal), ml
}

func (r *Interpreter) EvalReturnStmt(stmt *ReturnStmt, env *Environment) RuntimeVal {
	if env.ResolveEnv("function", r) == nil {
		env.ThrowSyntaxError("illegal use of the return keyword, return statements can only be used in the body of functions",
			SourceLog(stmt.line, stmt.col, stmt.count, env.sourcePath, ""))
	}
	var value RuntimeVal = undefined
	if stmt.value != nil {
		value = r.Evaluate(stmt.value, env)
	}
	r.CallStack.Pop()
	r.returned_from_function = true
	r.terminated = true
	return value
}

func (r *Interpreter) EvalBreakStmt(stmt *BreakStmt, env *Environment) RuntimeVal {
	if env.ResolveEnv("loop", r) == nil {
		env.ThrowSyntaxError("illegal use of the break keyword, break statements can only be used in the body of loops",
			SourceLog(stmt.line, stmt.col, stmt.count, env.sourcePath, ""))
	}
	r._break = true
	r.terminated = true
	return undefined
}

func (r *Interpreter) EvalContinueStmt(stmt *ContinueStmt, env *Environment) RuntimeVal {
	if env.ResolveEnv("loop", r) == nil {
		env.ThrowSyntaxError("illegal use of the continue keyword, continue statements can only be used in the body of loops",
			SourceLog(stmt.line, stmt.col, stmt.count, env.sourcePath, ""))
	}
	r._continue = true
	r.terminated = true
	return undefined
}

var anonyClassCount = 0

func (r *Interpreter) EvalClassDecl(decl *ClassDecl, env *Environment) (*ClassVal, string) {
	extends := ""
	if len(decl.extends) > 0 {
		pos := decl.Pos
		ml := env.ReferenceOf(decl.extends, pos.line, pos.col, pos.count, env.sourcePath, r)
		c, ok := Memory.get(ml).(*ClassVal)
		if !ok {
			pos := getPosFromNode(decl)
			env.ThrowTypeError("cannot extend type", ValueType(c)+",", "it is not a class and is not constructable", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
		}
		extends = ml
	}
	class := MK_CLASS(decl.name, decl.constructor, decl.properties, decl.methods, env, extends, r)
	if class.anonymous {
		class.name = "$" + string(rune(anonyClassCount))
		ml := GenerateRadix(16)
		return class, ml
	}
	ml, v := env.DeclareVarRef(class.name, class, "constant", decl.line, decl.col, decl.count, env.sourcePath, r)
	return v.(*ClassVal), ml
}

func (r *Interpreter) EvalImportStmt(node *ImportStmt, env *Environment) RuntimeVal {
	if node.from != nil {
		module := r.Eval_from_expr(node.from, env)
		if len(node.namespace) > 0 {
			env.DeclareVar(node.namespace, module, "static", node.line, node.col, node.count, env.sourcePath, r)
		} else {
			DestructureObjectDecl(module, node.names, "constant", env, r)
		}
	} else {
		path := node.path
		current_module_path := env.sourcePath
		path = RealPath(RelativePathToFile(current_module_path, path))
		env.sourcePath = path
		runtime := NewRuntime()
		parser := NewParser(path, "module", "")
		AST := parser.Parse(false)
		runtime.EvalProgram(AST, env)
		env.sourcePath = current_module_path
	}
	return undefined
}

func (r *Interpreter) EvalExportStmt(node *ExportStmt, env *Environment) RuntimeVal {
	switch export := node.export.(type) {
	case *VarDecl:
		rhs := r.Evaluate(export.right, env)
		decls := r.DeclareVar(export, rhs, env)
		decls.forEach(func(key, value string) {
			r.exports.set(MK_STRING(key), value)
		})
	case *FunctionDecl:
		fn, ml := r.EvalFunctionDecl(export, env)
		r.exports.set(MK_STRING(fn.name), ml)
	case *ClassDecl:
		cl, ml := r.EvalClassDecl(export, env)
		r.exports.set(MK_STRING(cl.name), ml)
	case *ObjectLiteral:
		obj := r.Eval_object(export, env)
		r.exports.copy(obj.properties)
	default:
		panic("unimplemented")
	}
	return undefined
}

func (r *Interpreter) EvalSwitchStmt(stmt *SwitchStmt, env *Environment) *Undefined {
	condition := r.Evaluate(stmt.on, env)
	new_block := NewEnv(env, "block", env.sourcePath)
	found := false
	for i := 0; i < len(stmt.cases); i++ {
		c := stmt.cases[i]
		_case := r.Evaluate(c.condition, env)
		if RtvAreEqual(_case, condition) {
			r.EvalBlock(c.body, new_block)
			found = true
			break
		}
	}
	if !found {
		r.EvalBlock(stmt.def, new_block)
	}
	return undefined
}

// Declares Destructured Properties
func DestructureObjectDecl(val RuntimeVal, destructuring *ObjectLiteral, _type string, env *Environment, r *Interpreter) *Map[string, string] {
	var obj *ObjectVal
	proto := MK_OBJECT(nil, nil, nil)
	switch v := val.(type) {
	case *ObjectVal:
		obj = v
		p, ok := v.prototype.(*ObjectVal)
		if ok {
			proto = p
		}
	case *Instance:
		obj = v.ObjectVal
		p, ok := v.prototype.(*ObjectVal)
		if ok {
			proto = p
		}
	default:
		pos := getPosFromNode(destructuring)
		env.ThrowTypeError("cannot destructure type", ValueType(val), "it is not an object"+SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
	}
	decls := NewMap[string, string]()
	destructuring.properties.forEach(func(key DynamicNode, value Node) {
		prop_key := ""
		ident := ""
		if key.dynamic {
			prop_key = r.Evaluate(key.node, env).noAnsi()
		} else {
			prop_key = key.node.(*Identifier).Symbol
		}
		if value != nil {
			switch i := value.(type) {
			case *Identifier:
				ident = i.Symbol
			default:
				pos := getPosFromNode(value)
				env.ThrowSyntaxError("unexpected token in object destructuring, identifier expected:" + SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
			}
		} else {
			ident = prop_key
		}
		key_v := MK_STRING(prop_key)
		ml := obj.properties.get(key_v)
		if len(ml) == 0 {
			ml = GetPropMlFromProto(key_v, proto)
			if len(ml) == 0 {
				pos := getPosFromNode(key.node)
				env.ThrowReferenceError("type object has no property named", prop_key, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
			}
		}
		property := Memory.get(ml)
		pos := getPosFromNode(destructuring)
		env.DeclareVar(ident, property, _type, pos.line, pos.col, pos.count, env.sourcePath, r)
		decls.set(ident, ml)
	})
	return decls
}

// Assigns Destructured Properties
func DestructureObjectAssign(val RuntimeVal, destructuring *ObjectLiteral, env *Environment, r *Interpreter) {
	obj, ok := val.(*ObjectVal)
	if !ok {
		pos := getPosFromNode(destructuring)
		env.ThrowTypeError("cannot destructure type", ValueType(val), "it is not an object"+SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
	}
	destructuring.properties.forEach(func(key DynamicNode, value Node) {
		prop_key := ""
		ident := ""
		if key.dynamic {
			prop_key = r.Evaluate(key.node, env).noAnsi()
		} else {
			prop_key = key.node.(*Identifier).Symbol
		}
		if value != nil {
			switch i := value.(type) {
			case *Identifier:
				ident = i.Symbol
			default:
				pos := getPosFromNode(value)
				env.ThrowSyntaxError("unexpected token in object destructuring, identifier expected:" + SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
			}
		} else {
			ident = prop_key
		}
		ml := obj.properties.get(MK_STRING(prop_key))
		if len(ml) == 0 {
			pos := getPosFromNode(key.node)
			env.ThrowReferenceError("type object has no property named", prop_key, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
		}
		property := Memory.get(ml)
		pos := getPosFromNode(destructuring)
		env.AssignVar(ident, property, pos.line, pos.col, pos.count, env.sourcePath, r)
	})
}

// Declares Destructured Properties
func DestructureArrayDecl(val RuntimeVal, destructuring *ArrayLiteral, _type string, env *Environment, r *Interpreter) *Map[string, string] {
	arr, ok := val.(*ArrayVal)
	if !ok {
		pos := getPosFromNode(destructuring)
		env.ThrowTypeError("cannot destructure type", ValueType(val), "it is not an array", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
	}
	decls := NewMap[string, string]()
	for i := 0; i < len(destructuring.elements); i++ {
		node := destructuring.elements[i]
		ident := node.(*Identifier)
		ml := arr.getRef(i)
		pos := getPosFromNode(ident)
		env.DeclareVar(ident.Symbol, Memory.get(ml), _type, pos.line, pos.col, pos.count, env.sourcePath, r)
		decls.set(ident.Symbol, ml)
	}
	return decls
}

// Assigns Destructured Properties
func DestructureArrayAssign(val RuntimeVal, destructuring *ArrayLiteral, env *Environment, r *Interpreter) *Map[string, string] {
	arr, ok := val.(*ArrayVal)
	if !ok {
		pos := getPosFromNode(destructuring)
		env.ThrowTypeError("cannot destructure type", ValueType(val), "it is not an array", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
	}
	decls := NewMap[string, string]()
	for i, v := range destructuring.elements {
		ident := v.(*Identifier)
		var ml string = arr.getRef(i)
		pos := getPosFromNode(ident)
		env.AssignVar(ident.Symbol, Memory.get(ml), pos.line, pos.col, pos.count, env.sourcePath, r)
		decls.set(ident.Symbol, ml)
	}
	return decls
}

var ud_ref = ""

func GetUDRef(r *Interpreter) string {
	return GetGlobalEnv(r).variables.get("undefined")
}

func RtvToBool(condition RuntimeVal) bool {
	switch val := condition.(type) {
	case *BoolVal:
		return val.value
	case *StringVal:
		return len(val.value) > 0
	case *NumberVal:
		return val.value != 0
	case *ObjectVal:
		return val.properties.length > 0
	case *ArrayVal:
		return val.elements.length > 0
	default:
		return false
	}
}

//#endregion

//#region Expressions

func (r *Interpreter) Eval_ternary_expr(expr *TernaryExpr, env *Environment) RuntimeVal {
	condition := r.Evaluate(expr.condition, env)
	if RtvToBool(condition) {
		return r.Evaluate(expr.then, env)
	}
	return r.Evaluate(expr._else, env)
}

func (r *Interpreter) Eval_instanceof_expr(expr *InstanceofExpr, env *Environment) RuntimeVal {
	lhs := r.Evaluate(expr.left, env)
	rhs := r.Evaluate(expr.right, env)
	boolean := false
	if ValueType(rhs) == "class" && ValueType(lhs) == "instance" {
		instance := lhs.(*Instance)
		class := rhs.(*ClassVal)
		pos := expr.Pos
		ml := env.ReferenceOf(class.name, pos.line, pos.col, pos.count, env.sourcePath, r)
		boolean = instance.class == ml
	}
	return MK_BOOL(boolean)
}

func (r *Interpreter) Eval_dynamic_import(expr *DynamicImport, env *Environment) RuntimeVal {
	DynamicImportMacro := MK_MACRO("import", func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal {
		const err = "Dynamic import's specifier must be of type 'string', but here has type"
		if len(args) == 0 {
			env.ThrowTypeError(err, ValueType(undefined))
		}
		specifier := args[0]
		spec_type := ValueType(specifier)
		if spec_type != "string" {
			env.ThrowTypeError(err, spec_type)
		}
		path := specifier.(*StringVal).value
		module := r.Eval_from_expr(&FromExpr{
			path: path,
			Pos:  pos,
		}, env)
		return module
	})
	args := []RuntimeVal{r.Evaluate(expr.specifier, env)}
	if expr.async {
		return Promise(DynamicImportMacro, args, r, env, expr.Pos)
	}
	return DynamicImportMacro.call(args, env, expr.Pos, r)
}

func (r *Interpreter) Eval_match_expr(expr *MatchExpr, env *Environment) RuntimeVal {
	match_against := r.Evaluate(expr.match, env)
	var value RuntimeVal = null
	for i := 0; i < len(expr.cases); i++ {
		_case := expr.cases[i]
		match := r.Evaluate(_case.match, env)
		if RtvAreEqual(match_against, match) {
			value = r.Evaluate(_case.body, env)
			break
		}
	}
	return value
}

func (r *Interpreter) Eval_from_expr(node *FromExpr, env *Environment) *ObjectVal {
	path := node.path
	path = RealPath(RelativePathToFile(env.sourcePath, path))
	script_env := CreateScriptEnv(r, path)
	runtime := NewRuntime()
	parser := NewParser(path, "module", "")
	AST := parser.Parse(false)
	module := runtime.EvalProgram(AST, script_env)
	return module
}

func (r *Interpreter) Eval_logical_expr(expr *LogicalExpr, env *Environment) RuntimeVal {
	op := expr.op
	var value RuntimeVal
	left := r.Evaluate(expr.left, env)
	lhs := RtvToBool(left)
	switch op {
	case "&&":
		right := r.Evaluate(expr.right, env)
		if !lhs {
			value = left
		} else {
			value = right
		}
	case "||":
		right := r.Evaluate(expr.right, env)
		if lhs {
			value = left
		} else {
			value = right
		}
	case "!":
		value = MK_BOOL(!lhs)
	}
	return value
}

func (r *Interpreter) Eval_new_expr(expr *NewExpr, env *Environment) RuntimeVal {
	var ml string
	args := []RuntimeVal{}
	switch exp := expr.operand.(type) {
	case *CallExpr:
		ml = r.getRef(exp.caller, env)
		args = r.eval_args(exp.args, env)
	default:
		ml = r.getRef(exp, env)
	}
	pos := getPosFromNode(expr)
	return r.Instantiate(ml, env, args, pos)
}

func (r *Interpreter) getRef(exp Node, env *Environment) string {
	switch node := exp.(type) {
	case *MemberExpr:
		return r.Get_Member(node, env)
	case *Identifier:
		pos := getPosFromNode(node)
		return env.ReferenceOf(node.Symbol, pos.line, pos.col, pos.count, env.sourcePath, r)
	default:
		ml := GenerateRadix(16)
		Memory.set(ml, r.Evaluate(node, env))
		return ml
	}
}

func (r *Interpreter) Eval_super_expr(node *SuperExpr, ctor_body *Environment) RuntimeVal {
	pos := getPosFromNode(node)
	this := ctor_body.LookupVar("this", pos.line, pos.col, pos.count, ctor_body.sourcePath, r).(*Instance)
	class := Memory.get(this.class).(*ClassVal)
	if len(class.extends) == 0 {
		return null
	}
	args := []RuntimeVal{}
	for i := 0; i < len(node.args); i++ {
		arg := node.args[i]
		args = append(args, r.Evaluate(arg, ctor_body))
	}
	// executes base contructor
	supers_instance := r.Instantiate(class.extends, ctor_body, args, pos)
	supers_proto, ok := supers_instance.prototype.(*ObjectVal)
	if ok {
		this_proto, ok := this.prototype.(*ObjectVal)
		if ok {
			// adds to the prototype chain
			// this.prototype.prototype (null)
			// this.prototype.prototype = new_proto
			this_proto.prototype = supers_proto
		} else {
			this_proto = supers_proto
		}
	}
	return undefined
}

func (r *Interpreter) Instantiate(class_ml string, env *Environment, args []RuntimeVal, pos Pos) *Instance {
	value := Memory.get(class_ml)
	class, ok := value.(*ClassVal)
	if !ok {
		env.ThrowTypeError("type", ValueType(value), "is not a class and is not constructable", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
	}
	prototype := NewMap[RuntimeVal, string]()
	class_body := NewEnv(class.declEnv, "object", class.declEnv.sourcePath)
	this := MK_INSTANCE(class.name, class_ml, prototype, ud_ref, class_body, r)
	// if len(class.extends) > 0 {
	// }
	class_body.DeclareVar("this", this, "constant", pos.line, pos.col, pos.count, class_body.sourcePath, r)
	for i := 0; i < len(class.fields); i++ {
		prop := class.fields[i]
		pos := getPosFromNode(prop)
		v := r.Evaluate(prop.value, class_body)
		ml, _ := class_body.DeclareVarRef(prop.name, v, "mutable", pos.line, pos.col, pos.count, class_body.sourcePath, r)
		prototype.set(MK_STRING(prop.name), ml)
		if prop._default {
			this._default = ml
		}
	}
	for i := 0; i < len(class.methods); i++ {
		method := class.methods[i]
		v, ml := r.EvalFunctionDecl(&method.decl, class_body)
		Memory.set(ml, v)
		class_body.variables.set(v.name, ml)
		prototype.set(MK_STRING(v.name), ml)
	}
	r.CallCtor(class.ctor, args, class_body, this)
	return this
}

func (r *Interpreter) CallCtor(ctor *Constructor, args []RuntimeVal, env *Environment, this *Instance) RuntimeVal {
	scope := NewEnv(env, "function", env.sourcePath)
	DeclareCtorParams(ctor.params, args, scope, this, r)
	function := MK_FUNCTION("constructor", ctor.body, []Node{}, env, false, false, false, r)
	return r.pushToStack(*function, *scope)
}

func DeclareCtorParams(params []CtorParam, args []RuntimeVal, scope *Environment, this *Instance, r *Interpreter) {
top:
	for i := 0; i < len(params); i++ {
		param := params[i]
		var arg RuntimeVal = undefined
		if i < len(args) {
			arg = args[i]
		}
		pos := getPosFromNode(&param)
		switch param := param.expr.(type) {
		case *Identifier:
			scope.DeclareVar(param.Symbol, arg, "mutable", pos.line, pos.col, pos.count, scope.sourcePath, r)
		case *AssignmentExpr:
			switch l := param.left.(type) {
			case *Identifier:
				value := arg
				if ValIsNullish(arg) {
					value = r.Evaluate(param.right, scope)
				}
				scope.DeclareVar(
					l.Symbol,
					value,
					"mutable",
					pos.line,
					pos.col,
					pos.count,
					scope.sourcePath,
					r,
				)
			}
		case *RestOrSpreadExpr:
			array := MK_ARRAY()
			for j := i; j >= i; j++ {
				if j >= len(args) {
					DeclareParams([]Node{param.operand}, []RuntimeVal{array}, scope, r)
					break top
				}
				arg := args[j]
				array.Push(arg)
			}
		default:
			panic("unimplemented parameter")
		}
	}
}

func (r *Interpreter) Eval_await_expr(expr *AwaitExpr, env *Environment) RuntimeVal {
	var value RuntimeVal
	switch op := expr.operand.(type) {
	case *CallExpr:
		fn, ok := r.Evaluate(op.caller, env).(*FunctionVal)
		async := false
		if ok {
			async = fn.async
			fn.async = false
		}
		args := r.eval_args(op.args, env)
		value = fn.Call(env, args, r, op.Pos)
		fn.async = async
	case *DynamicImport:
		op.async = false
		value = r.Eval_dynamic_import(op, env)
	default:
		value = r.Evaluate(op, env)
	}
	return value
}

func (r *Interpreter) Eval_call_expr(expr *CallExpr, env *Environment) RuntimeVal {
	value := r.Evaluate(expr.caller, env)
	args := r.eval_args(expr.args, env)
	rv, _ := CallFunction(value, env, args, r, expr.Pos)
	return rv
}

// func ExecAsyncFunc(fn *FunctionVal, env *Environment, r *Interpreter) *Instance {}

func CallFunction(value RuntimeVal, env *Environment, args []RuntimeVal, r *Interpreter, pos Pos) (RuntimeVal, *Environment) {
	switch v := value.(type) {
	case *FunctionVal:
		funtion_scope := NewEnv(v.declEnv, "function", env.sourcePath)
		r.ResolveTHIS(v, pos, env, funtion_scope)
		if v.async {
			// sets promise_mem_loc to this new Promise
			return Promise(v, args, r, env, pos), funtion_scope
		}
		DeclareParams(v.params, args, funtion_scope, r)
		return r.pushToStack(*v, *funtion_scope), funtion_scope
	case *Macro:
		return v.call(args, env, pos, r), nil
	default:
		env.ThrowTypeError("type", ValueType(value), "is not a function and is not callable", SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
	}
	// this line is unreachable
	return null, nil
}

func (r *Interpreter) ResolveTHIS(v *FunctionVal, pos Pos, env *Environment, funtion_scope *Environment) {
	// if the function is not enclosed in a class body
	if v.declEnv.ResolveEnv("object", r) == nil {
		if v.arrow {
			object := v.declEnv.LookupVar("this", pos.line, pos.col, pos.count, env.sourcePath, r)
			funtion_scope.DeclareVar("this", object, "constant", pos.line, pos.col, pos.count, env.sourcePath, r)
		} else {
			funtion_scope.DeclareVar("this", MK_OBJECT(nil, funtion_scope, r), "constant", pos.line, pos.col, pos.count, env.sourcePath, r)
		}
	}
}

func (r *Interpreter) pushToStack(function FunctionVal, funtion_scope Environment) RuntimeVal {
	var lastEvaluated RuntimeVal = undefined
	function.declEnv = &funtion_scope
	r.CallStack.Push(function)
	// for r.CallStack.length > 0 {
	callback := r.CallStack.at(-1)
	lastEvaluated = r.EvalBlock(callback.body, callback.declEnv)
	if r.returned_from_function {
		r.terminated = false
		r.returned_from_function = false
	} else {
		lastEvaluated = undefined
		r.CallStack.Pop()
	}
	// }
	return lastEvaluated
}

func DeclareParams(params []Node, args []RuntimeVal, funtion_scope *Environment, r *Interpreter) {
	// loop from first to last
top:
	for i := 0; i < len(params); i++ {
		param := params[i]
		var arg RuntimeVal = undefined
		if i < len(args) {
			arg = args[i]
		}
		pos := getPosFromNode(param)
		switch param := param.(type) {
		case *Identifier:
			funtion_scope.DeclareVar(param.Symbol, DuplicateRtv(arg), "mutable", pos.line, pos.col, pos.count, funtion_scope.sourcePath, r)
		case *AssignmentExpr:
			switch l := param.left.(type) {
			case *Identifier:
				value := arg
				if ValIsNullish(arg) {
					value = r.Evaluate(param.right, funtion_scope)
				}
				funtion_scope.DeclareVar(
					l.Symbol,
					value,
					"mutable",
					pos.line,
					pos.col,
					pos.count,
					funtion_scope.sourcePath,
					r,
				)
			}
		case *RestOrSpreadExpr:
			array := MK_ARRAY()
			for j := i; j >= i; j++ {
				if j >= len(args) {
					DeclareParams([]Node{param.operand}, []RuntimeVal{array}, funtion_scope, r)
					break top
				}
				arg := args[j]
				array.Push(arg)
			}
		default:
			panic("unimplemented parameter")
		}
	}
}

func DuplicateRtv(v RuntimeVal) RuntimeVal {
	switch rtv := v.(type) {
	case *BoolVal, *NullVal, *NumberVal, *StringVal, *Symbol, *Undefined:
		return rtv
	case *Macro:
		m := *rtv
		return &m
	case *ArrayVal:
		new_array := MK_ARRAY()
		rtv.forEach(func(key int, value RuntimeVal) {
			new_array.set(key, value)
		})
		return new_array
	case *ClassVal:
		c := *rtv
		return &c
	case *FunctionVal:
		c := *rtv
		return &c
	case *Instance:
		c := *rtv
		return &c
	case *NativeClass:
		c := *rtv
		return &c
	case *ObjectVal:
		new_object := MK_OBJECT(nil, rtv.body_env, rtv.r)
		// new_map := NewMap[string, RuntimeVal]()
		rtv.properties.forEach(func(key RuntimeVal, value string) {
			ml := GenerateRadix(16)
			Memory.set(ml, Memory.get(value))
			new_object.properties.set(key, ml)
		})
		return new_object
	default:
		r, ok := rtv.(*RawVal[any])
		if ok {
			v := &r.value
			return MK_RAW(*v)
		}
		v := &rtv
		return *v
		// panic(fmt.Sprintf("unexpected main.RuntimeVal: %#v", rtv))
	}
}

func (r *Interpreter) eval_args(arguments []Node, env *Environment) []RuntimeVal {
	args := []RuntimeVal{}
	// do not use range over loop
	for i := 0; i < len(arguments); i++ {
		arg := arguments[i]
		switch arg := arg.(type) {
		case *RestOrSpreadExpr:
			value := r.Evaluate(arg.operand, env)
			valType := ValueType(value)
			if valType == "array" {
				array := value.(*ArrayVal)
				array.forEach(func(_ int, el RuntimeVal) {
					args = append(args, el)
				})
			} else {
				pos := getPosFromNode(arg)
				env.ThrowTypeError("cannot spread type", valType, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
			}
		default:
			args = append(args, r.Evaluate(arg, env))
		}
	}
	return args
}

func (r *Interpreter) Eval_globalThisMemberAssignment(expr *globalThisMemberAssignment, env *Environment) RuntimeVal {
	ml, _, invalidMl := r.Get_globalThis_Member(expr.Symbol, expr.property, expr.line, expr.col, expr.count, env)
	rhs := r.Evaluate(expr.right, env)
	if invalidMl { // variable does not exist
		e := env.ResolveEnv("program", r)
		e.variables.set(expr.property, ml)
		e.varTypes.set(expr.property, "mutable")
	}
	Memory.set(ml, rhs)
	return rhs
}

func (r *Interpreter) Eval_globalThisMember(node *globalThisMember, env *Environment) RuntimeVal {
	ml, ud, invalidMl := r.Get_globalThis_Member(node.Symbol, node.property, node.line, node.col, node.count, env)
	if invalidMl {
		return ud
	}
	return Memory.get(ml)
}

func (r *Interpreter) Get_globalThis_Member(Symbol, property string, line, col, count int, env *Environment) (string, RuntimeVal, bool) {
	object := env.LookupVar(Symbol, line, col, count, env.sourcePath, r).(*ObjectVal)
	ml := object.properties.get(MK_STRING(property))
	if len(ml) == 0 {
		return "", undefined, true
	}
	return ml, nil, false
}

func (r *Interpreter) Eval_globalThis(node *globalThis, env *Environment) RuntimeVal {
	return env.LookupVar(node.Symbol, node.line, node.col, node.count, env.sourcePath, r)
}

func (r *Interpreter) Eval_grouping_expr(node *GroupingExpr, env *Environment) RuntimeVal {
	var lastEval RuntimeVal
	for i := 0; i < len(node.exprs); i++ {
		expr := node.exprs[i]
		lastEval = r.Evaluate(expr, env)
	}
	// this may never run
	if lastEval == nil {
		lastEval = undefined
	}
	return lastEval
}

func (r *Interpreter) Eval_increment_expr(node *IncrementExpr, env *Environment) RuntimeVal {
	operand := node.operand
	op := node.op
	operand_pos := getPosFromNode(operand)
	operand_value := r.Evaluate(operand, env)
	if ValueType(operand_value) != "number" {
		env.ThrowTypeError(
			SourceLog(operand_pos.line, operand_pos.col, operand_pos.count, env.sourcePath, ""),
		)
	}

	var number float64
	var value float64

	number = operand_value.Value().(float64)
	if op == "++" {
		value = number + 1
		if node.pre {
			number += 1
		}
	} else {
		value = number - 1
		if node.pre {
			number -= 1
		}
	}
	switch operand := operand.(type) {
	case *MemberExpr:
		member := r.Get_Member(
			operand,
			env)
		Memory.set(member, MK_NUMBER(value))
	default:
		ref := env.ReferenceOf(
			operand.(*Identifier).Symbol,
			operand_pos.line,
			operand_pos.col,
			operand_pos.count,
			env.sourcePath,
			r,
		)
		Memory.set(ref, MK_NUMBER(value))
	}
	return MK_NUMBER(number)
}

func (r *Interpreter) Eval_member_expr(expr *MemberExpr, env *Environment) RuntimeVal {
	ml := r.Get_Member(expr, env)
	if len(ml) == 0 {
		return undefined
	}
	return Memory.get(ml)
}

func (r *Interpreter) Get_Member(expr *MemberExpr, env *Environment) string {
	object_value := r.Evaluate(expr.object, env)
	computed_property, property := GetMemberExprProp(expr, r, env)
	var prop RuntimeVal
	if expr.computed {
		prop = computed_property
	} else {
		prop = MK_STRING(property)
	}
	pos := getPosFromNode(expr.property)
	switch v := object_value.(type) {
	case *ObjectVal:
		ml := v.properties.get(prop)
		return ml
	case *ArrayVal:
		if !expr.computed {
			env.ThrowTypeError(
				"cannot read properties of type array (reading", property+")",
				SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""),
			)
		}
		switch t := computed_property.(type) {
		case *NumberVal:
			break
		default:
			env.ThrowTypeError(
				"type", ValueType(t),
				"cannot be used to index an array",
				SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""),
			)
		}
		// always number
		index := computed_property.Value().(float64)
		ml := v.elements.slice[int(index)]
		if len(ml) == 0 {
			return ud_ref
		}
		return ml
	case *NullVal, *Undefined, *NumberVal, *BoolVal:
		env.ThrowTypeError(
			"cannot read properties of type", ValueType(v), "(reading", prop.noAnsi()+")",
			SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""),
		)
	case *Instance:
		ml := v.properties.get(prop)
		if len(ml) == 0 && !expr.computed {
			ml = GetPropMlFromProto(prop, v.prototype)
		}
		return ml
	case *StringVal:
		if !expr.computed {
			env.ThrowTypeError(
				"cannot read properties of type string (reading", prop.noAnsi()+")",
				SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""),
			)
		}
		index, ok := prop.Value().(float64)
		if !ok {
			env.ThrowTypeError(
				"type", ValueType(prop),
				"cannot be used to index a string",
				SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""),
			)
		}
		i := int(index)
		if i >= len(v.value) {
			return ud_ref
		} else if i < 0 {
			i += len(v.value)
		}
		char := string(v.value[i])
		ml := GenerateRadix(16)
		Memory.set(ml, MK_STRING(char))
		return ml
	case *FunctionVal:
		ml := v.properties.get(prop)
		if len(ml) == 0 && !expr.computed {
			ml = GetPropMlFromProto(prop, v.prototype)
		}
		return ml
	case *ClassVal:
		ml := v.properties.get(prop)
		if len(ml) == 0 && !expr.computed {
			ml = GetPropMlFromProto(prop, v.prototype)
		}
		return ml
	}
	panic("unimplemented")
}

func GetMemberExprProp(expr *MemberExpr, r *Interpreter, env *Environment) (RuntimeVal, string) {
	var computed_property RuntimeVal
	var property string
	if expr.computed {
		computed_property = r.Evaluate(expr.property, env)
	} else {
		property = expr.property.(*Identifier).Symbol
	}
	return computed_property, property
}

func GetPropMlFromProto(prop RuntimeVal, proto RuntimeVal) string {
	ml := ""
	switch proto := proto.(type) {
	case *ObjectVal:
		ml = proto.properties.get(prop)
		if len(ml) == 0 {
			return GetPropMlFromProto(prop, proto.prototype)
		}
	}
	return ml
}

func (r *Interpreter) Eval_object(object_lit *ObjectLiteral, env *Environment) *ObjectVal {
	object_env := NewEnv(env, "object", env.sourcePath)
	object_val := MK_OBJECT(nil, object_env, r)
	properties := NewMap[string, string]()
	pos := object_lit.Pos
	object_lit.properties.forEach(func(k DynamicNode, v Node) {
		var key RuntimeVal
		env := object_env
		if k.dynamic {
			key = r.Evaluate(k.node, env)
		} else {
			key = MK_STRING(k.node.(*Identifier).Symbol)
		}
		var value RuntimeVal
		if v == nil {
			value = env.LookupVar(key.noAnsi(), pos.line, pos.col, pos.count, env.sourcePath, r)
		} else {
			value = r.Evaluate(v, env)
		}
		switch v := value.(type) {
		case *FunctionVal:
			if v.anonymous || len(v.name) == 0 {
				v.name = key.noAnsi()
			}
		case *ClassVal:
			if v.anonymous || len(v.name) == 0 {
				v.name = key.noAnsi()
			}
		}
		ml := GenerateRadix(16)
		Memory.set(ml, value)
		object_val.properties.set(key, ml)
		properties.set(key.noAnsi(), ml)
	})
	object_env.DeclareVar("this", SCOPE_OBJECT(properties), "constant", pos.line, pos.col, pos.count, env.sourcePath, r)
	return object_val
}

func (r *Interpreter) Eval_array(node *ArrayLiteral, env *Environment) RuntimeVal {
	array := MK_ARRAY()
	for i := 0; i < len(node.elements); i++ {
		el := node.elements[i]
		array.Push(r.Evaluate(el, env))
	}
	return array
}

func (r *Interpreter) Eval_in_expr(node *InExpr, env *Environment) RuntimeVal {
	bool := false
	left := r.Evaluate(node.left, env)
	right := r.Evaluate(node.right, env)
	if ValueType(left) != "string" {
		pos := getPosFromNode(node.left)
		env.ThrowTypeError(
			"'in' cannot check for properties in type", ValueType(right), "with type", ValueType(left),
			SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""),
		)
	}
	// only in "own" properties
	switch right := right.(type) {
	case *ObjectVal:
		bool = right.properties.has(left)
	case *Instance:
		bool = right.properties.has(left)
	default:
		pos := getPosFromNode(node.right)
		env.ThrowTypeError(
			"'in' cannot check for properties in type", ValueType(right),
			SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""),
		)
	}
	return MK_BOOL(bool)
}

func (r *Interpreter) Eval_comparison_expr(expr *ComparisonExpr, env *Environment) RuntimeVal {
	bool := false
	op := expr.op
	left := r.Evaluate(expr.left, env)
	left_pos := getPosFromNode(expr.left)
	right_pos := getPosFromNode(expr.right)
	pos := Pos{
		line:  left_pos.line,
		col:   left_pos.col,
		count: left_pos.count + right_pos.col + right_pos.count - 5,
	}
	right := r.Evaluate(expr.right, env)
	lhs_type := ValueType(left)
	rhs_type := ValueType(right)
	comparison_op_err_msg := "'" + op + "' operator cannot take operands of type " + lhs_type + " and " + rhs_type +
		SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")
	lhs := 0.0
	rhs := 0.0
	if is_value(op, "<", ">", "<=", ">=") {
		lhs = RtvToInt(left, comparison_op_err_msg, env)
		rhs = RtvToInt(right, comparison_op_err_msg, env)
	}
	switch op {
	case "<":
		bool = lhs < rhs
	case ">":
		bool = lhs > rhs
	case "<=":
		bool = lhs <= rhs
	case ">=":
		bool = lhs >= rhs
	case "==":
		bool = RtvAreEqual(left, right)
	case "===":
		bool = RtvAreEqual(left, right) && lhs_type == rhs_type
	case "!=":
		bool = !RtvAreEqual(left, right)
	case "!==":
		bool = !RtvAreEqual(left, right) || lhs_type != rhs_type
	default:
		throwMessage("yo, unknown operator: " + op)
	}
	return MK_BOOL(bool)
}

func RtvAreEqual(v1, v2 RuntimeVal) bool {
	return reflect.DeepEqual(v1, v2)
}

func ValueType(v RuntimeVal) string {
	switch v.(type) {
	case *NumberVal:
		return "number"
	case *StringVal:
		return "string"
	case *BoolVal:
		return "boolean"
	case *NullVal:
		return "null"
	case *Undefined:
		return "undefined"
	case *ObjectVal:
		return "object"
	case *FunctionVal:
		return "function"
	case *Macro:
		return "macro"
	case *ClassVal:
		return "class"
	case *Instance:
		return "instance"
	case *Symbol:
		return "symbol"
	case *ArrayVal:
		return "array"
	case *NativeClass:
		return "class"
	default:
		return "raw"
		// return "\x1b[3munknown-value\x1b[0m"
	}
}

// convert values like strings, objects and arrays to AS numbers
func RtvToInt(runtimeVal RuntimeVal, err_msg string, env *Environment) float64 {
	switch rtv := runtimeVal.(type) {
	case *NumberVal:
		return rtv.value
	case *StringVal:
		return float64(len(rtv.value))
	default:
		env.ThrowTypeError(err_msg)
	}
	return 0
}

func (r *Interpreter) Eval_identifier(node *Identifier, env *Environment) RuntimeVal {
	return env.LookupVar(node.Symbol, node.line, node.col, node.count, env.sourcePath, r)
}

func (r *Interpreter) Eval_assignment(expr *AssignmentExpr, env *Environment) RuntimeVal {
	rhs := r.Evaluate(expr.right, env)
	var value RuntimeVal
	if expr.op == "=" {
		value = rhs
	} else {
		lhs := r.Evaluate(expr.left, env)
		pos := expr.Pos
		switch expr.op {
		case "+=":
			result := r.add(lhs.Value(), rhs.Value(), ValueType(lhs), ValueType(rhs), pos, env)
			switch r := result.(type) {
			case string:
				value = MK_STRING(r)
			case float64:
				value = MK_NUMBER(r)
			}
		case "-=":
			value = MK_NUMBER(r.sub(env, pos, lhs.Value(), rhs.Value()))
		case "/=":
			value = MK_NUMBER(r.div(env, pos, lhs.Value(), rhs.Value()))
		case "*=":
			value = MK_NUMBER(r.mul(env, pos, lhs.Value(), rhs.Value()))
		case "%=":
			value = MK_NUMBER(r.mod(env, pos, lhs.Value(), rhs.Value()))
		case "??=":
			// nullish assignment
			// if the value of lhs is nullish, then assign lhs with rhs
			// otherwise return lhs
			if ValIsNullish(lhs) {
				value = rhs
			} else {
				return lhs
			}
		}
	}
	switch exp := expr.left.(type) {
	case *Identifier:
		env.AssignVar(exp.Symbol, value, expr.line, expr.col, expr.count, env.sourcePath, r)
	case *MemberExpr:
		ml := r.Get_Member(exp, env) // member expression is verified
		if len(ml) == 0 {
			ml = GenerateRadix(16)
			o, ok := ResolveMemberObject(exp.object).(*Identifier)
			if ok {
				pos := getPosFromNode(o)
				e := env.ResolveVarEnv(o.Symbol, env, pos.line, pos.col, pos.count, env.sourcePath, r)
				if e.varTypes.get(o.Symbol) == "static" {
					env.ThrowSyntaxError("Assignment: to static variable",
						SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
				}
			}
			v := r.Evaluate(exp.object, env)
			computed_prop, prop := GetMemberExprProp(exp, r, env)
			var key RuntimeVal
			if exp.computed {
				key = computed_prop
			} else {
				key = MK_STRING(prop)
			}
			switch object := v.(type) {
			case *ObjectVal:
				object.properties.set(key, ml)
			case *Instance:
				object.properties.set(key, ml)
			case *FunctionVal:
				object.properties.set(key, ml)
			case *ClassVal:
				object.properties.set(key, ml)
			}
		}
		Memory.set(ml, value)
	case *ObjectLiteral:
		DestructureObjectAssign(rhs, exp, env, r)
	case *ArrayLiteral:
		DestructureArrayAssign(rhs, exp, env, r)
	default:
		pos := getPosFromNode(expr.left)
		env.ThrowSyntaxError("Invalid left hand side in assignment:" +
			SourceLog(pos.line, pos.col, pos.count, env.sourcePath, ""))
	}
	return value
}

func ResolveMemberObject(node Node) Node {
	n, ok := node.(*MemberExpr)
	if !ok {
		return n
	}
	switch o := n.object.(type) {
	case *MemberExpr:
		return ResolveMemberObject(o)
	default:
		return n.object
	}
}

func ValIsNullish(val RuntimeVal) bool {
	switch val.(type) {
	case *NullVal, *Undefined:
		return true
	default:
		return false
	}
}

func (r *Interpreter) Eval_typeof(node *TypeOfExpr, env *Environment) RuntimeVal {
	value := r.Evaluate(node.operand, env)
	return MK_STRING(ValueType(value))
}

func (r *Interpreter) Eval_void_expr(expr *VoidExpr, env *Environment) RuntimeVal {
	r.Evaluate(expr.operand, env)
	return undefined
}

func (r *Interpreter) Eval_binary_expr(expr *BinaryExpr, env *Environment) RuntimeVal {
	v1 := r.Evaluate(expr.left, env)
	v2 := r.Evaluate(expr.right, env)
	lhs := v1.Value()
	rhs := v2.Value()
	pos := expr.Pos
	var value RuntimeVal
	switch expr.op {
	case "+":
		v := r.add(lhs, rhs, ValueType(v1), ValueType(v2), pos, env)
		switch v := v.(type) {
		case string:
			value = &StringVal{v}
		case float64:
			value = &NumberVal{v}
		}
	case "-":
		v := r.sub(env, pos, lhs, rhs)
		value = &NumberVal{v}
	case "*":
		v := r.mul(env, pos, lhs, rhs)
		value = &NumberVal{v}
	case "/":
		v := r.div(env, pos, lhs, rhs)
		value = &NumberVal{v}
	case "%":
		v := r.mod(env, pos, lhs, rhs)
		value = &NumberVal{v}
	case "**":
		v := r.exp(lhs, rhs, env, pos)
		value = &NumberVal{v}
	}
	return value
}

// returns (string | float64)
func (r *Interpreter) add(n1, n2 any, t1, t2 string, pos Pos, env *Environment) any {
	value := ""
	found_string := false
	switch v := n1.(type) {
	case string:
		found_string = true
		value = v
	case float64:
		value = fmt.Sprint(v)
	default:
		env.ThrowTypeError(fmt.Sprintf("'+' operation between type %s and %s is invalid.%s", t1, t2, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")))
	}
	switch v := n2.(type) {
	case string:
		found_string = true
		value += v
	case float64:
		if found_string {
			value += fmt.Sprint(v)
		} else {
			float, _ := strconv.ParseFloat(value, 64)
			value = fmt.Sprint(float + v)
		}
	default:
		env.ThrowTypeError(fmt.Sprintf("'+' operation between type %s and %s is invalid.%s", t1, t2, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")))
	}
	if found_string {
		return value
	}
	number, _ := strconv.ParseFloat(value, 64)
	return number
}

func (r *Interpreter) sub(env *Environment, pos Pos, values ...any) float64 {
	value := 0.0
	for i := 0; i < len(values); i++ {
		v := values[i]
		switch v := v.(type) {
		case float64:
			if i == 0 {
				value = v
			} else {
				value -= v
			}
		default:
			index := max(i-1, 0)
			prev_val := values[index]
			env.ThrowTypeError(fmt.Sprintf("'-' operation between type %T and %T is invalid.%s", prev_val, v, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")))
		}
	}
	return value
}

func (r *Interpreter) mul(env *Environment, pos Pos, values ...any) float64 {
	value := 0.0
	for i := 0; i < len(values); i++ {
		v := values[i]
		switch v := v.(type) {
		case float64:
			if i == 0 {
				value = v
			} else {
				value *= v
			}
		default:
			index := max(i-1, 0)
			prev_val := values[index]
			env.ThrowTypeError(fmt.Sprintf("'*' operation between type %T and %T is invalid.%s", prev_val, v, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")))
		}
	}
	return value
}

func (r *Interpreter) div(env *Environment, pos Pos, values ...any) float64 {
	value := 0.0
	for i := 0; i < len(values); i++ {
		v := values[i]
		switch v := v.(type) {
		case float64:
			if i == 0 {
				value = v
			} else {
				value /= v
			}
		default:
			index := max(i-1, 0)
			prev_val := values[index]
			env.ThrowTypeError(fmt.Sprintf("'/' operation between type %T and %T is invalid.%s", prev_val, v, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")))
		}
	}
	return value
}

func (r *Interpreter) mod(env *Environment, pos Pos, values ...any) float64 {
	value := 0.0
	for i := 0; i < len(values); i++ {
		v := values[i]
		switch v := v.(type) {
		case float64:
			if i == 0 {
				value = v
			} else {
				value = math.Mod(math.Round(value), v)
			}
		default:
			index := max(i-1, 0)
			prev_val := values[index]
			env.ThrowTypeError(fmt.Sprintf("'%%' operation between type %T and %T is invalid.%s", prev_val, v, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")))
		}
	}
	return value
}

func (r *Interpreter) exp(v1, v2 any, env *Environment, pos Pos) float64 {
	value, ok := v1.(float64)
	f2, v2ok := v2.(float64)
	if !ok || !v2ok {
		env.ThrowTypeError(fmt.Sprintf("'**' operation between type %T and %T is invalid.%s", v1, v2, SourceLog(pos.line, pos.col, pos.count, env.sourcePath, "")))
	}
	p := int(f2)
	for i := p; i <= p; i++ {
		value *= value
	}
	return float64(value)
}

// #region Runtime Func

// Create A new Runtime
func NewRuntime() *Interpreter {
	return &Interpreter{
		returned_from_function: false,
		terminated:             false,
		_break:                 false,
		_continue:              false,
		MemThreshold:           100 * 1024 * 1024, // 100MB
		CallStack:              NewStack(),
		microTaskQueue: &MicroTaskQueue{
			queue:  []Task{},
			length: 0,
		},
		exports: NewMap[RuntimeVal, string](),
	}
}

func GenerateRadix(radix int) string {
	integer := rand.Int()
	return strconv.FormatInt(int64(integer), radix)
}

// ----------------------------------------------------
// ----------------------------------------------------
// --- -- ------  --- --  - -----  -  ----      -- -- -
// --- -- --     ---  --  -- ----  -   ----    -- -- --
// --- -- ----   ---  --  ---- --  -    ----  -- -- ---
// --- -- --     ---  --  ------   -     ------ -------
// --- -- ------  --- --  -------  -      ---- --  ----
// ----------------------------------------------------
// ----------------------------------------------------
//#region Env | Scope

func CreateGlobalEnv(r *Interpreter) *Environment {
	_env_ := NewEnv(nil, "global", "")
	// _env_.DeclareVar("yo", MK_NUMBER(600), "constant", 1, 1, 1, _env_.sourcePath, r)
	_env_.DeclareVar("true", MK_BOOL(true), "static", 1, 1, 1, _env_.sourcePath, r)
	_env_.DeclareVar("false", MK_BOOL(false), "static", 1, 1, 1, _env_.sourcePath, r)
	_env_.DeclareVar("undefined", undefined, "static", 1, 1, 1, _env_.sourcePath, r)
	_env_.DeclareVar("null", null, "static", 1, 1, 1, _env_.sourcePath, r)
	_env_.DeclareVar("this", MK_OBJECT(nil, _env_, r), "static", 1, 1, 1, _env_.sourcePath, r)
	macros.forEach(func(_ string, value *Macro) {
		value.DeclareMacro(_env_, r)
	})
	return _env_
}

var globalEnv *Environment

func GetGlobalEnv(r *Interpreter) *Environment {
	if globalEnv == nil {
		globalEnv = CreateGlobalEnv(r)
	}
	return globalEnv
}

type Environment struct {
	parent *Environment
	// key: variable identifier, value: reference
	variables *Map[string, string]
	// key: variable identifier, value: type ("constant" | "mutable" | "static" | "var")
	varTypes *Map[string, string]
	// ("global", "script", "block", "function")
	_type      string
	sourcePath string
	// Error handling ...
	catch_block   []Node
	catch_param   Node
	finally_block []Node
}

// get all variable names and references from the current scope to the global scope
func (env *Environment) all() *Map[string, string] {
	vars := NewMap[string, string]()
	vars.copy(env.variables)
	if env.parent != nil {
		vars.copy(env.parent.all())
	}
	return vars
}

func (env *Environment) DeleteVar(
	symbol string,
	line, col, count int, path string,
	r *Interpreter,
) string {
	decl_env := env.parent.ResolveVarEnv(symbol, env, line, col, count, path, r)
	ml := decl_env.ReferenceOf(symbol, line, col, count, path, r)
	decl_env.varTypes.delete(symbol)
	decl_env.variables.delete(symbol)
	return ml
}

func (env *Environment) ReferenceOf(
	varname string,
	line, col, count int, path string,
	r *Interpreter,
) string {
	e := env.ResolveVarEnv(varname, env, line, col, count, path, r)
	return e.variables.get(varname)
}

func (env *Environment) DeclareVar(
	varname string, value RuntimeVal, _type string,
	line, col, count int, path string,
	r *Interpreter,
) RuntimeVal {
	if env.variables.has(varname) {
		env.ThrowSyntaxError("cannot redeclare " + env.varTypes.get(varname) + " variable " + varname +
			SourceWithinRange(path, line, col, count, "") +
			SourceAtPosition(path, line, col))
	}
	ml := GenerateRadix(16)
	env.variables.set(varname, ml)
	env.varTypes.set(varname, _type)
	Memory.set(ml, value)
	return value
}

func (env *Environment) DeclareVarRef(name string, value RuntimeVal, _type string, line int, col int, count int, path string, r *Interpreter) (string, RuntimeVal) {
	if env.variables.has(name) {
		env.ThrowSyntaxError("cannot redeclare " + env.varTypes.get(name) + " variable " + name +
			SourceWithinRange(path, line, col, count, "") +
			SourceAtPosition(path, line, col))
	}
	ml := GenerateRadix(16)
	env.variables.set(name, ml)
	env.varTypes.set(name, _type)
	Memory.set(ml, value)
	return ml, value
}

func (env *Environment) AssignVar(
	varname string, value RuntimeVal,
	line, col, count int, path string,
	r *Interpreter,
) RuntimeVal {
	e := env.ResolveVarEnv(varname, env, line, col, count, path, r)
	if is_value(env.varTypes.get(varname), "constant", "static") {
		env.ThrowSyntaxError("assignment to " + env.varTypes.get(varname) + " variable: \x1b[34m" + varname + "\x1b[0m" +
			SourceWithinRange(path, line, col, count, "") +
			SourceAtPosition(path, line, col))
	}
	ml := e.variables.get(varname)
	Memory.set(ml, value)
	return value
}

func (env *Environment) ResolveVarEnv(
	varname string, e *Environment,
	line, col, count int, path string,
	r *Interpreter,
) *Environment {
	if env.variables.has(varname) {
		return env
	}
	if env.parent != nil {
		return env.parent.ResolveVarEnv(varname, e, line, col, count, path, r)
	}
	e.ThrowReferenceError("could not resolve variable `" + varname + "` as it does not exist" +
		SourceWithinRange(path, line, col, count, "") +
		SourceAtPosition(path, line, col))
	return nil
}

func (env *Environment) LookupVar(
	symbol string,
	line, col, count int, path string,
	r *Interpreter,
) RuntimeVal {
	ml := env.ReferenceOf(symbol, line, col, count, path, r)
	if symbol != "globalThis" && ml == env.ReferenceOf("globalThis", line, col, count, path, r) {
		env.ThrowReferenceError("invalid reference to globalThis" + SourceLog(line, col, count, path, ""))
	}
	return Memory.get(ml)
}

func (env *Environment) ResolveEnv(_type string, r *Interpreter) *Environment {
	if env._type == _type {
		return env
	}
	if env.parent != nil {
		return env.parent.ResolveEnv(_type, r)
	}
	return nil
}

func (env *Environment) ThrowSyntaxError(message ...string) {
	print("\x1b[31mSyntaxError\x1b[0m: ")
	env.throwError(message)
}

func (env *Environment) ThrowReferenceError(message ...string) {
	print("\x1b[31mReferenceError\x1b[0m: ")
	env.throwError(message)
}

func (env *Environment) ThrowTypeError(s ...string) {
	print("\x1b[31mTypeError\x1b[0m: ")
	env.throwError(s)
}

// func (env *Environment) throwRuntimeError(message string) {
// 	print("\x1b[31mRuntimeError\x1b[0m: ")
// 	env.throwError(message)
// }

func (env *Environment) throwValue(value RuntimeVal, r *Interpreter) {
	if try_ := env.ResolveEnv("try", r); try_ != nil {
		r.terminated = false
		catch_block := NewEnv(try_.parent, "block", env.sourcePath)
		// always identifier (for now)
		ident := try_.catch_param.(*Identifier)
		pos := getPosFromNode(ident)
		catch_block.DeclareVar(
			ident.Symbol,
			value,
			"mutable",
			pos.line,
			pos.col,
			pos.count,
			try_.sourcePath,
			r,
		)
		r.EvalBlock(try_.catch_block, catch_block)
		return
	}
	print("Uncaught \x1b[31mError\x1b[0m: ")
	PrintRtv(value)
	os.Exit(1)
}

// errors thrown with this cannot be caught
func (env *Environment) throwError(message []string,
) {
	print(JoinSlice(message, " "), "\r\n")
	os.Exit(1)
}

func CreateScriptEnv(r *Interpreter, path string) *Environment {
	if !IsAbs(path) {
		path = AbsPath(path)
	}
	script := NewEnv(GetGlobalEnv(r), "program", path)
	script.DeclareVar("globalThis", SCOPE_OBJECT(script.all()), "constant", 1, 1, 1, script.sourcePath, r)
	return script
}

func SCOPE_OBJECT(obj *Map[string, string]) *ObjectVal {
	props := NewMap[RuntimeVal, string]()
	obj.forEach(func(key, value string) {
		props.set(MK_STRING(key), value)
	})
	return &ObjectVal{
		properties: props,
		prototype:  null,
		value:      "",
	}
}

func NewEnv(parent *Environment, _type, path string) *Environment {
	return &Environment{
		parent:    parent,
		variables: NewMap[string, string](),
		varTypes:  NewMap[string, string](),
		_type:     _type, sourcePath: path,
	}
}
