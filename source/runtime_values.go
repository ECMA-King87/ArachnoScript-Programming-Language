package main

import (
	"fmt"
	"reflect"
	"strings"
)

type RuntimeVal interface {
	Value() any
	String(depth int, sep string) string
	noAnsi() string
}

// Numbers
type NumberVal struct {
	value float64
}

func MK_NUMBER(n float64) *NumberVal {
	return &NumberVal{n}
}

func (n *NumberVal) Value() any {
	return n.value
}

func (n *NumberVal) noAnsi() string {
	return fmt.Sprint(n.value)
}

func (n *NumberVal) String(_ int, _ string) string {
	return fmt.Sprintf("\x1b[33m%v\x1b[0m", n.value)
}

// Strings
type StringVal struct {
	value string
}

func MK_STRING(s string) *StringVal {
	return &StringVal{s}
}

func (s *StringVal) Value() any {
	return s.value
}

func (n *StringVal) noAnsi() string {
	return n.value
}

func (s *StringVal) String(depth int, _ string) string {
	surrounding_pairs := []string{"", ""}
	if depth == 0 {
		// do nothing at all
	} else if strings.Contains(s.value, "\"") && strings.Contains(s.value, "'") {
		surrounding_pairs = []string{"\x1b[32m`", "`\x1b[0m"}
	} else if strings.Contains(s.value, "\"") {
		surrounding_pairs = []string{"\x1b[32m'", "'\x1b[0m"}
	} else {
		surrounding_pairs = []string{"\x1b[32m\"", "\"\x1b[0m"}
	}
	return fmt.Sprintf(surrounding_pairs[0]+"%v"+surrounding_pairs[1], s.value)
}

// Undefined
type BoolVal struct {
	value bool
}

func MK_BOOL(b bool) *BoolVal {
	return &BoolVal{b}
}

func (b *BoolVal) Value() any {
	return b.value
}

func (b *BoolVal) noAnsi() string {
	return fmt.Sprintf("%t", b.value)
}

func (n *BoolVal) String(_ int, _ string) string {
	return fmt.Sprintf("\x1b[33m%t\x1b[0m", n.value)
}

// Undefined
type Undefined struct{}

func MK_UD() *Undefined {
	return &Undefined{}
}

func (u *Undefined) Value() any {
	return nil
}

func (u *Undefined) String(_ int, _ string) string {
	return "\x1b[1;97mundefined\x1b[0m"
}

func (u *Undefined) noAnsi() string {
	return "undefined"
}

// Null
type NullVal struct{}

func MK_NULL() *NullVal {
	return &NullVal{}
}

func (n *NullVal) Value() any {
	return nil
}

func (u *NullVal) noAnsi() string {
	return "null"
}

func (n *NullVal) String(_ int, _ string) string {
	return "\x1b[1;97mnull\x1b[0m"
}

// Object
type ObjectVal struct {
	// "own" properties
	properties *Map[RuntimeVal, string]
	// a property that every object will inherit.
	// Value: either null or object
	prototype RuntimeVal
	body_env  *Environment
	r         *Interpreter
	value     string
}

// (key: property key, value: reference to value)
type ObjectProps *Map[RuntimeVal, string]

func MK_OBJECT(props ObjectProps, body_env *Environment, r *Interpreter) *ObjectVal {
	if props == nil {
		props = NewMap[RuntimeVal, string]()
	}
	return &ObjectVal{
		properties: props,
		value:      "\x1b[36m[object]\x1b[0m",
		prototype:  null,
		body_env:   body_env,
		r:          r,
	}
}

func (obj *ObjectVal) Value() any {
	return obj.value
}

func (n *ObjectVal) noAnsi() string {
	return "[object]"
}

var maxDepth = 4

func (obj *ObjectVal) String(depth int, sep string) string {
	if depth > maxDepth {
		return obj.value
	}
	debug_symbol := MK_STRING(symbol_table.get("debug").noAnsi())
	proto_ml := ""
	if proto, ok := obj.prototype.(*ObjectVal); ok {
		proto_ml = GetPropMlFromProto(debug_symbol, proto)
		if len(proto_ml) > 0 {
			method, ok := Memory.get(proto_ml).(*FunctionVal)
			if ok {
				v, ok := method.Call(obj.body_env, []RuntimeVal{MK_STRING(sep)}, obj.r, Pos{}).(*StringVal)
				if ok {
					return v.value
				}
			}
		}
	}
	pairs := []string{}
	object := MapEntries(obj.properties)
	fullLength := len(object)
	props := [][]any{}
	for i := 0; i < len(object); i++ {
		k := object[i][0].(RuntimeVal)
		prop := Memory.get(object[i][1].(string))
		props = append(props, []any{
			k.String(depth, ""),
			prop.String(depth+1, sep),
			depth + 1,
		},
		)
	}
	for i := 0; i < len(props); i++ {
		k := props[i][0].(string)
		prop := props[i][1].(string)
		d := props[i][2].(int)
		str := ""
		if fullLength > 3 {
			str = "\n" + strings.Repeat(sep, d)
		} else {
			str = sep
		}
		str += k + ": " + prop
		if i == len(object)-1 {
			if fullLength > 3 {
				if d > 0 {
					d -= 1
				}
				str += "\r\n" + strings.Repeat(sep, d)
			} else {
				str += sep
			}
		} else {
			str += ","
		}
		pairs = append(pairs, str) // always indent
	}
	return "{" + JoinSlice(pairs, "") + "}"
}

type ArrayValue struct {
	slice  []string
	length int
}

// Array
type ArrayVal struct {
	elements ArrayValue
	value    string
}

func (arr *ArrayVal) getRef(index int) string {
	if index >= arr.elements.length {
		return ud_ref
	}
	ml := arr.elements.slice[index]
	return ml
}

func (arr *ArrayVal) get(index int) RuntimeVal {
	if index >= arr.elements.length {
		return undefined
	}
	ml := arr.elements.slice[index]
	return Memory.get(ml)
}

func (arr *ArrayVal) set(index int, value RuntimeVal) RuntimeVal {
	ml := GenerateRadix(16)
	Memory.set(ml, value)
	if index >= arr.elements.length {
		arr.elements.length = index + 1
		for i := index; i >= index; i-- {
			arr.elements.slice = append(arr.elements.slice, "")
		}
	}
	arr.elements.slice[index] = ml
	return value
}

func (arr *ArrayVal) forEach(callback callback[int, RuntimeVal]) {
	for index := range arr.elements.slice {
		callback(index, arr.get(index))
	}
}

func MK_ARRAY(elements ...RuntimeVal) *ArrayVal {
	_array_ := ArrayValue{}
	array_val := &ArrayVal{
		elements: _array_,
		value:    "\x1b[36m[array]\x1b[0m",
	}
	array_val.Push(elements...)
	return array_val
}

func (arr *ArrayVal) Push(elements ...RuntimeVal) *ArrayVal {
	length := &arr.elements.length
	index := 0
	if *length > 0 {
		index = *length - 1
	}
	// do not use range over loop
	for i := 0; i < len(elements); i++ {
		el := elements[i]
		ml := GenerateRadix(16)
		*length++
		arr.elements.slice = append(arr.elements.slice, ml)
		index++
		Memory.set(ml, el)
	}
	return arr
}

func (arr *ArrayVal) Value() any {
	return arr.value
}

func (n *ArrayVal) noAnsi() string {
	return "[array]"
}

func (arr *ArrayVal) String(depth int, sep string) string {
	if depth > maxDepth {
		return arr.value
	}
	visited := NewMap[RuntimeVal, bool]()
	elements := arr.elements
	length := elements.length
	fullLength := length
	array := []string{}
	// do not use range over loop
	for i := 0; i < length; i++ {
		el := arr.get(i)
		if RtvAreEqual(arr, el) {
			visited.set(el, true)
		}
		if visited.has(el) {
			str := ""
			if fullLength > 5 {
				d := depth
				if depth == 0 {
					d++
				}
				str = "\n" + strings.Repeat(sep, d)
			}
			str += "\x1b[36m[Circular]\x1b[0m"
			if i == length-1 {
				if fullLength > 5 {
					str += "\r\n"
				}
			} else {
				str += ", "
			}
			array = append(array, str)
		} else {
			str := ""
			if length > 5 {
				d := depth
				if depth == 0 {
					d++
				}
				str += "\r\n" + strings.Repeat(sep, d)
			}
			str += el.String(depth+1, sep)
			if i != length-1 {
				str += ", "
			} else {
				str += " "
			}
			if length > 5 {
				str += "\r\n"
			}
			array = append(array, str)
		}
	}
	return "[ " + JoinSlice(array, "") + "]"
}

type NativeFunction func(args []RuntimeVal, env *Environment, pos Pos, r *Interpreter) RuntimeVal

// Macro
type Macro struct {
	call  NativeFunction
	name  string
	value string
}

func (m *Macro) noAnsi() string {
	return m.name
}

// String implements RuntimeVal.
func (m *Macro) String(_ int, _ string) string {
	return m.value
}

// Value implements RuntimeVal.
func (m *Macro) Value() any {
	return m.call
}

func (m *Macro) Call(env *Environment, args []RuntimeVal, r *Interpreter, pos Pos) RuntimeVal {
	value := m.call(args, env, pos, r)
	return value
}

func MK_MACRO(name string, call NativeFunction) *Macro {
	return &Macro{
		call:  call,
		name:  name,
		value: "\x1b[34m[macro " + name + "]\x1b[0m",
	}
}

func (m *Macro) DeclareMacro(env *Environment, r *Interpreter) *Macro {
	env.DeclareVar(m.name, m, "static", 1, 1, 1, env.sourcePath, r)
	return m
}

// Function
type FunctionVal struct {
	*ObjectVal
	name      string
	body      []Node
	params    []Node
	declEnv   *Environment
	r         *Interpreter
	async     bool
	anonymous bool
	arrow     bool
}

func MK_FUNCTION(name string, body, params []Node, declEnv *Environment, async, anonymous, arrow bool, r *Interpreter) *FunctionVal {
	fn_val := &FunctionVal{
		ObjectVal: &ObjectVal{
			value:      "\x1b[36m[function]\x1b[0m",
			properties: NewMap[RuntimeVal, string](),
			prototype:  MK_OBJECT(nil, declEnv, nil),
		},
		name:      name,
		body:      body,
		async:     async,
		anonymous: anonymous,
		params:    params,
		declEnv:   declEnv,
		arrow:     arrow,
		r:         r,
	}
	return fn_val
}

func (fn *FunctionVal) Value() any {
	return fn.value
}

func (fn *FunctionVal) noAnsi() string {
	return fn.name
}

func (fn *FunctionVal) String(_ int, _ string) string {
	return "\x1b[36m[function " + fn.name + "]\x1b[0m"
}

func (fn *FunctionVal) Call(env *Environment, args []RuntimeVal, r *Interpreter, pos Pos) RuntimeVal {
	value, _ := CallFunction(fn, env, args, r, pos)
	return value
}

// Class
type ClassVal struct {
	*ObjectVal
	name      string
	anonymous bool
	ctor      *Constructor
	fields    []*ClassProperty
	methods   []*ClassMethod
	declEnv   *Environment
	extends   string
}

func MK_CLASS(
	name string,
	ctor *Constructor,
	props []*ClassProperty, methods []*ClassMethod,
	declEnv *Environment,
	extends string,
	r *Interpreter,
) *ClassVal {
	class := &ClassVal{
		ObjectVal: &ObjectVal{
			value:      "\x1b[36m[class]\x1b[0m",
			properties: NewMap[RuntimeVal, string](),
			prototype:  MK_OBJECT(nil, nil, r),
		},
		name:    name,
		fields:  props,
		extends: extends,
		methods: methods,
		declEnv: declEnv,
		ctor:    ctor,
	}
	return class
}

func (class *ClassVal) Value() any {
	return class.value
}

func (class *ClassVal) noAnsi() string {
	return class.name
}

func (class *ClassVal) String(_ int, _ string) string {
	name := class.name
	if name[0] == '$' {
		name = "(anonymous)"
	}
	return "\x1b[36m[class " + name + "]\x1b[0m"
}

// Class
type NativeClass struct {
	*ObjectVal
	name string
	ctor *Macro
	// key: identifier, value: 0x0
	properties *Map[string, string]
	methods    []*Macro
	declEnv    *Environment
	extends    string
}

func MK_NT_CLASS(
	name string,
	ctor *Macro,
	props *Map[string, string], methods []*Macro,
	declEnv *Environment,
	extends string,
) *NativeClass {
	class := &NativeClass{
		ObjectVal: &ObjectVal{
			value:      "\x1b[36m[class]\x1b[0m",
			properties: NewMap[RuntimeVal, string](),
			prototype:  MK_OBJECT(nil, nil, nil),
		},
		name:       name,
		extends:    extends,
		methods:    methods,
		declEnv:    declEnv,
		ctor:       ctor,
		properties: props,
	}
	return class
}

func (native *NativeClass) noAnsi() string {
	return native.name
}

func (class *NativeClass) Value() any {
	return class.value
}

func (class *NativeClass) String(_ int, _ string) string {
	name := class.name
	if name[0] == '$' {
		name = "(anonymous)"
	}
	return "\x1b[36m[class " + name + "]\x1b[0m"
}

// Instance
type Instance struct {
	*ObjectVal
	name     string
	_default string
	// reference to it's constructor
	class      string
	r          *Interpreter
	class_body *Environment
}

func MK_INSTANCE(
	name, class string,
	proto *Map[RuntimeVal, string],
	_default string,
	class_body *Environment,
	r *Interpreter,
) *Instance {
	instance := &Instance{
		ObjectVal: &ObjectVal{
			value:      "\x1b[35m[object Instance]\x1b[0m",
			properties: NewMap[RuntimeVal, string](),
			prototype:  MK_OBJECT(proto, class_body, r),
		},
		name:       name,
		class:      class,
		r:          r,
		class_body: class_body,
		_default:   _default,
	}
	return instance
}

func (i *Instance) noAnsi() string {
	return i.name
}

func (i *Instance) Value() any {
	return i.value
}

func (i *Instance) String(depth int, sep string) string {
	if depth > maxDepth {
		return "\x1b[36m[" + i.name + "]\x1b[0m"
	}
	debug_symbol := MK_STRING(symbol_table.get("debug").noAnsi())
	proto_ml := ""
	if proto, ok := i.prototype.(*ObjectVal); ok {
		proto_ml = GetPropMlFromProto(debug_symbol, proto)
		if len(proto_ml) > 0 {
			method, ok := Memory.get(proto_ml).(*FunctionVal)
			if ok {
				rv := method.Call(i.class_body, []RuntimeVal{MK_STRING(sep)}, i.r, Pos{})
				v, ok := rv.(*StringVal)
				if ok {
					return v.value
				}
			}
		}
	}
	if len(i.properties.get(debug_symbol)) > 0 {
		method, ok := Memory.get(i.properties.get(debug_symbol)).(*FunctionVal)
		if ok {
			rv := method.Call(i.class_body, []RuntimeVal{MK_STRING(sep)}, i.r, Pos{})
			v, ok := rv.(*StringVal)
			if ok {
				return v.value
			}
		}
	}
	pairs := []string{}
	object := MapEntries(i.properties)
	fullLength := len(object)
	props := [][]any{}
	for i := 0; i < len(object); i++ {
		k := object[i][0].(RuntimeVal)
		prop := Memory.get(object[i][1].(string))
		props = append(props, []any{
			k.String(depth, ""),
			prop.String(depth+1, sep),
			depth + 1,
		},
		)
	}
	for i := 0; i < len(props); i++ {
		k := props[i][0].(string)
		prop := props[i][1].(string)
		d := props[i][2].(int)
		str := ""
		if fullLength > 3 {
			str = "\n" + strings.Repeat(sep, d)
		} else {
			str = sep
		}
		str += k + ": " + prop
		if i == len(object)-1 {
			if fullLength > 3 {
				if d > 0 {
					d -= 1
				}
				str += "\r\n" + strings.Repeat(sep, d)
			} else {
				str += sep
			}
		} else {
			str += ","
		}
		pairs = append(pairs, str) // always indent
	}
	return i.name + " {" + JoinSlice(pairs, "") + "}"
}

// Symbol
type Symbol struct {
	symbol string
}

func MK_SYMBOL(
	sym string,
) *Symbol {
	symbol := &Symbol{"Symbol(" + sym + ")"}
	return symbol
}

func (i *Symbol) noAnsi() string {
	return i.symbol
}

func (i *Symbol) Value() any {
	return i.symbol
}

func (i *Symbol) String(_ int, _ string) string {
	return "\x1b[32m" + i.symbol + "\x1b[0m"
}

type RAW struct{}

// RawVal
type RawVal[T any] struct {
	RAW
	value T
}

func MK_RAW[T any](
	value T,
) *RawVal[T] {
	RAW := &RawVal[T]{
		value: value,
		RAW:   RAW{},
	}
	return RAW
}

func (i *RawVal[T]) noAnsi() string {
	return fmt.Sprintf("%T", i.value)
}

func (i *RawVal[T]) Value() any {
	return i.value
}

func (i *RawVal[T]) String(_ int, _ string) string {
	return fmt.Sprintf("%+v", i.value)
}

func HasEmbededType(t any, e any) bool {
	typeof := reflect.TypeOf(t)
	fields := reflect.VisibleFields(typeof)
	for _, v := range fields {
		if v.Anonymous {
			return v.Type.AssignableTo(reflect.TypeOf(e))
		}
	}
	return false
}

func JoinSlice(slice []string, sep string) string {
	str := ""
	for i := 0; i < len(slice); i++ {
		str += slice[i]
		if i < len(slice) {
			str += sep
		}
	}
	return str
}

func PrintRtv(value RuntimeVal) {
	print(value.String(0, "  "))
}
