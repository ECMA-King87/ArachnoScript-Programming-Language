package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Node interface {
	String() string
	node()
}

// #region Structs

type Pos struct {
	line, col, count int
}

type Parser struct {
	tokens     *TokenArray
	tokenIndex uint
	program    *Program
	scriptType string
	sourcePath string
}

// Program (AST)
type Program struct {
	body       []Node
	scriptType string
	sourcePath string
	main       bool
	Pos
}

// node implements Node.
func (prog *Program) node() {}

// String implements Node.
func (prog *Program) String() string {
	return fmt.Sprintf("Node \x1b[32mProgram\x1b[0m {\r\n  body: %+v,\r\n  source path: \"%s\",\r\n  script type: %s,  \r\n  main: %t\r\n}", prog.body, prog.sourcePath, prog.scriptType, prog.main)
}

// Variable Declaration (AST)
type VarDecl struct {
	left  Node
	right Node
	_type string
	Pos
}

// node implements Node.
func (decl *VarDecl) node() {}

// String implements Node.
func (decl *VarDecl) String() string {
	return fmt.Sprintf("Node \x1b[32mVariable Declaration\x1b[0m {\r\n left: %+v, right: %+v, type: %s \r\n}", decl.left, decl.right, decl._type)
}

// If Statement (AST)
type IfStmt struct {
	condition Node
	body      []Node
	elseBody  []Node
	Pos
}

// node implements Node.
func (stmt *IfStmt) node() {}

// String implements Node.
func (stmt *IfStmt) String() string {
	return fmt.Sprintf("Node \x1b[32mIf Statement\x1b[0m {\r\n  condition: %+v\r\n  body: %+v,\r\n  else: %+v\r\n}", stmt.condition, stmt.body, stmt.elseBody)
}

// While Loop (AST)
type WhileLoop struct {
	condition Node
	body      []Node
	do        bool
	Pos
}

// node implements Node.
func (stmt *WhileLoop) node() {}

// String implements Node.
func (stmt *WhileLoop) String() string {
	return fmt.Sprintf("Node \x1b[32mWhile Loop\x1b[0m {\r\n  condition: %+v\r\n  body: %+v,\r\n  do: %t\r\n}", stmt.condition, stmt.body, stmt.do)
}

// Throw Statement (AST)
type ThrowStmt struct {
	value Node
	Pos
}

// node implements Node.
func (stmt *ThrowStmt) node() {}

// String implements Node.
func (stmt *ThrowStmt) String() string {
	return fmt.Sprintf("Node \x1b[32mThrow Statement\x1b[0m {\r\n  value: %+v\r\n}", stmt.value)
}

// Try-Catch-Finally (AST)
type TryCatch struct {
	try         []Node
	catch       []Node
	finally     []Node
	catch_param Node
	Pos
}

// node implements Node.
func (stmt *TryCatch) node() {}

// String implements Node.
func (stmt *TryCatch) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mTry-Catch-Finally Block\x1b[0m {\r\n  try: %+v\r\n  catch: %+v\r\n finally: %+v\r\n catch block param: %+v\r\n}",
		stmt.try,
		stmt.catch,
		stmt.finally,
		stmt.catch_param,
	)
}

// A block of code (AST)
type BlockStmt struct {
	body []Node
	Pos
}

// node implements Node.
func (stmt *BlockStmt) node() {}

// String implements Node.
func (stmt *BlockStmt) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mBlock\x1b[0m {\r\n  body: %+v\r\n}", stmt.body)
}

// Delete Statement (AST)
type DeleteStmt struct {
	operand Node
	Pos
}

// node implements Node.
func (stmt *DeleteStmt) node() {}

// String implements Node.
func (stmt *DeleteStmt) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mDelete Statement\x1b[0m {\r\n  operand: %+v\r\n}", stmt.operand)
}

// Traditional For Loop (AST)
type ForLoop struct {
	before    Node
	condition Node
	after     Node
	body      []Node
	Pos
}

// node implements Node.
func (stmt *ForLoop) node() {}

// String implements Node.
func (stmt *ForLoop) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mFor Loop\x1b[0m {\r\n  condition: %+v\r\n  before loop: %+v\r\n  after loop: %+v\r\n  body: %+v\r\n}",
		stmt.condition,
		stmt.before,
		stmt.after,
		stmt.body,
	)
}

// For Iterators Loop (AST)
type ForIteratorLoop struct {
	left  Node
	right Node
	_type string
	op    string // ("in" | "of")
	body  []Node
	Pos
}

// node implements Node.
func (stmt *ForIteratorLoop) node() {}

// String implements Node.
func (stmt *ForIteratorLoop) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mFor Iterators Loop\x1b[0m {\r\n  left: %+v\r\n  right: %+v\r\n  body: %+v\r\n}",
		stmt.left,
		stmt.right,
		stmt.body,
	)
}

// Function Declaration (AST)
type FunctionDecl struct {
	name      DynamicNode
	async     bool
	anonymous bool
	_type     string // ("arrow" | "")
	body      []Node
	params    []Node
	Pos
}

// node implements Node.
func (stmt *FunctionDecl) node() {}

// String implements Node.
func (stmt *FunctionDecl) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mFunction Declaration\x1b[0m {\r\n  name: %+v\r\n  parameters: %+v\r\n  async: %t\r\n  type: %s\r\n  body: %+v\r\n}",
		stmt.name,
		stmt.params,
		stmt.async,
		stmt._type,
		stmt.body,
	)
}

// Return Statement (AST)
type ReturnStmt struct {
	value Node
	Pos
}

// node implements Node.
func (stmt *ReturnStmt) node() {}

// String implements Node.
func (stmt *ReturnStmt) String() string {
	return fmt.Sprintf("Node \x1b[32mReturn Statement\x1b[0m {\r\n  name: %s\r\n}", stmt.value)
}

// Break Statement (AST)
type BreakStmt struct {
	Pos
}

// node implements Node.
func (stmt *BreakStmt) node() {}

// String implements Node.
func (stmt *BreakStmt) String() string {
	return fmt.Sprintf("Node \x1b[32mBreak Statement\x1b[0m { pos: %+v }", stmt.Pos)
}

// Continue Statement (AST)
type ContinueStmt struct {
	Pos
}

// node implements Node.
func (stmt *ContinueStmt) node() {}

// String implements Node.
func (stmt *ContinueStmt) String() string {
	return fmt.Sprintf("Node \x1b[32mContinue Statement\x1b[0m { pos: %+v }", stmt.Pos)
}

// Label (AST)
type Label struct {
	name string
	Pos
}

// node implements Node.
func (stmt *Label) node() {}

// String implements Node.
func (stmt *Label) String() string {
	return fmt.Sprintf("Node \x1b[32mLabel\x1b[0m { pos: %+v }", stmt.Pos)
}

// Goto Statement (AST)
type GotoStmt struct {
	label Label
	Pos
}

// node implements Node.
func (stmt *GotoStmt) node() {}

// String implements Node.
func (stmt *GotoStmt) String() string {
	return fmt.Sprintf("Node \x1b[32mGotoStmt\x1b[0m { pos: %+v }", stmt.Pos)
}

type ClassProperty struct {
	private  bool
	_default bool
	static   bool
	name     string
	value    Node
	Pos
}

// node implements Node.
func (stmt *ClassProperty) node() {}

// String implements Node.
func (stmt *ClassProperty) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mClass Property\x1b[0m {\r\n  private: %t\r\n  default: %t\r\n  static: %t\r\n  name: %sr\n  value: %+v\r\n}",
		stmt.private,
		stmt._default,
		stmt.static,
		stmt.name,
		stmt.value,
	)
}

type ClassMethod struct {
	private bool
	static  bool
	name    struct {
		dynamic bool
		node    Node
	}
	decl FunctionDecl
	Pos
}

// node implements Node.
func (stmt *ClassMethod) node() {}

// String implements Node.
func (stmt *ClassMethod) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mClass Method\x1b[0m {\r\n  private: %t\r\n  static: %t\r\n  name: %+v\r\n  decl: %+v\r\n}",
		stmt.private,
		stmt.static,
		stmt.name,
		stmt.decl,
	)
}

type CtorParam struct {
	private bool
	public  bool
	expr    Node
	Pos
}

// node implements Node.
func (stmt *CtorParam) node() {}

// String implements Node.
func (stmt *CtorParam) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mCunstructor Parameter\x1b[0m {\r\n  private: %t\r\n  public: %t\r\n  expr: %+v\r\n}",
		stmt.private,
		stmt.public,
		stmt.expr,
	)
}

type Constructor struct {
	name      string // constructor
	async     bool   // false
	anonymous bool   // false
	params    []CtorParam
	_type     string // constructor
	body      []Node
	Pos
}

// node implements Node.
func (stmt *Constructor) node() {}

// String implements Node.
func (stmt *Constructor) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mConstructor\x1b[0m {\r\n  params: %+v\r\n  body: %+v\r\n}",
		stmt.params,
		stmt.body,
	)
}

// Class Declaration (AST)
type ClassDecl struct {
	name        string
	anonymous   bool
	properties  []*ClassProperty
	methods     []*ClassMethod
	extends     string
	constructor *Constructor
	Pos
}

// node implements Node.
func (stmt *ClassDecl) node() {}

// String implements Node.
func (stmt *ClassDecl) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mClass Declaration\x1b[0m {\r\n  name: %+v\r\n  pos: %+v }",
		stmt.name,
		stmt.Pos,
	)
}

// Import Statement (AST)
type ImportStmt struct {
	path      string
	namespace string
	names     *ObjectLiteral
	from      *FromExpr
	Pos
}

// node implements Node.
func (stmt *ImportStmt) node() {}

// String implements Node.
func (stmt *ImportStmt) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mImport Statement\x1b[0m {\r\n  path: %+v\r\n  pos: %+v }",
		stmt.path,
		stmt.Pos,
	)
}

// Dynamic Import (AST)
type DynamicImport struct {
	specifier Node
	async     bool
	Pos
}

// node implements Node.
func (stmt *DynamicImport) node() {}

// String implements Node.
func (stmt *DynamicImport) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mImport Statement\x1b[0m {\r\n  specifier: %+v\r\n  pos: %+v }",
		stmt.specifier,
		stmt.Pos,
	)
}

// Export Stmt (AST)
type ExportStmt struct {
	export Node // (decl | object literal)
	// decl   bool
	Pos
}

// node implements Node.
func (stmt *ExportStmt) node() {}

// String implements Node.
func (stmt *ExportStmt) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mExport Statement\x1b[0m {\r\n  export: %+v\r\n  pos: %+v }",
		stmt.export,
		stmt.Pos,
	)
}

type Case struct {
	condition Node
	body      []Node
}

// Switch Stmt (AST)
type SwitchStmt struct {
	cases []Case
	on    Node
	def   []Node
	Pos
}

// node implements Node.
func (stmt *SwitchStmt) node() {}

// String implements Node.
func (stmt *SwitchStmt) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mSwitch Statement\x1b[0m {\r\n  cases: %+v\r\n  pos: %+v }",
		stmt.cases,
		stmt.Pos,
	)
}

// -----------------------------------------------
// -----------------------------------------------
// ----------------- Expressions -----------------
// -----------------------------------------------
// -----------------------------------------------

// Ternary (AST)
type TernaryExpr struct {
	condition Node
	then      Node
	_else     Node
	Pos
}

// node implements Node.
func (expr *TernaryExpr) node() {}

// String implements Node.
func (expr *TernaryExpr) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mTernary Expr\x1b[0m {\r\n  then: %+v\r\n  else: %+v\r\npos: %+v }",
		expr.then,
		expr._else,
		expr.Pos,
	)
}

// Instance Of (AST)
type InstanceofExpr struct {
	left  Node
	right Node
	Pos
}

// node implements Node.
func (expr *InstanceofExpr) node() {}

// String implements Node.
func (expr *InstanceofExpr) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mInstance Of Expr\x1b[0m {\r\n  left: %+v\r\n  right: %+v\r\npos: %+v }",
		expr.left,
		expr.right,
		expr.Pos,
	)
}

// Template String (AST)
type TemplateString struct {
	str []Node
	Pos
}

// node implements Node.
func (expr *TemplateString) node() {}

// String implements Node.
func (expr *TemplateString) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mTemplate String\x1b[0m {\r\n  operand: %+v\r\n  pos: %+v }",
		expr.str,
		expr.Pos,
	)
}

// RestOrSpread Expression (AST)
type RestOrSpreadExpr struct {
	operand Node
	Pos
}

// node implements Node.
func (expr *RestOrSpreadExpr) node() {}

// String implements Node.
func (expr *RestOrSpreadExpr) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mRest Or Spread Expression\x1b[0m {\r\n  operand: %+v\r\n  pos: %+v }",
		expr.operand,
		expr.Pos,
	)
}

type Match struct {
	match Node
	body  Node
}

// Match Expression (AST)
type MatchExpr struct {
	cases []Match
	match Node
	Pos
}

// node implements Node.
func (expr *MatchExpr) node() {}

// String implements Node.
func (expr *MatchExpr) String() string {
	return fmt.Sprintf(
		"Node \x1b[32mMatch Expression\x1b[0m {\r\n  cases: %+v\r\n  pos: %+v }",
		expr.cases,
		expr.Pos,
	)
}

// From Expression (AST).
type FromExpr struct {
	path string
	Pos
}

// node implements Node.
func (expr *FromExpr) node() {}

// String implements Node.
func (expr *FromExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mFrom Expression\x1b[0m {\r\n  path: %s\r\n}", expr.path)
}

// Logical Expression (AST).
type LogicalExpr struct {
	left  Node
	right Node
	op    string
	Pos
}

// node implements Node.
func (expr *LogicalExpr) node() {}

// String implements Node.
func (expr *LogicalExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mLogical Expression\x1b[0m {\r\n  left: %+v\r\n  right: %+v\r\n  op: %s\r\n}", expr.left, expr.right, expr.op)
}

// New Expression (AST).
type NewExpr struct {
	operand Node
	Pos
}

// node implements Node.
func (expr *NewExpr) node() {}

// String implements Node.
func (expr *NewExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mNew Expression\x1b[0m {\r\n  operand: %+v\r\n}", expr.operand)
}

// Super Expression (AST).
type SuperExpr struct {
	args []Node
	Pos
}

// node implements Node.
func (expr *SuperExpr) node() {}

// String implements Node.
func (expr *SuperExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mSuper Expression\x1b[0m {\r\n  args: %+v\r\n}", expr.args)
}

// Await Expression (AST).
type AwaitExpr struct {
	operand Node
	Pos
}

// node implements Node.
func (expr *AwaitExpr) node() {}

// String implements Node.
func (expr *AwaitExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mAwait Expression\x1b[0m {\r\n  operand: %+v\r\n}", expr.operand)
}

// Call Expression (AST).
type CallExpr struct {
	caller Node
	args   []Node
	Pos
}

// node implements Node.
func (expr *CallExpr) node() {}

// String implements Node.
func (expr *CallExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mCall Expression\x1b[0m {\r\n  caller: %+v\r\n  arguments: %+v\r\n}", expr.caller, expr.args)
}

// globalThis Member Expression (AST).
//
// cannot be a chain
type globalThisMemberAssignment struct {
	Symbol   string
	property string
	right    Node
	op       string
	Pos
}

// node implements Node.
func (expr *globalThisMemberAssignment) node() {}

// String implements Node.
func (expr *globalThisMemberAssignment) String() string {
	return fmt.Sprintf("Node \x1b[32mglobalThis Member Assignment\x1b[0m {\r\n  position: %+v\r\n  property: %s\r\n  right: %+v }", expr.Pos, expr.property, expr.right)
}

// globalThis Member Expression (AST).
//
// cannot be a chain
type globalThisMember struct {
	Symbol   string
	property string
	Pos
}

// node implements Node.
func (expr *globalThisMember) node() {}

// String implements Node.
func (expr *globalThisMember) String() string {
	return fmt.Sprintf("Node \x1b[32mglobalThis Member\x1b[0m { position: %+v, property: %s }", expr.Pos, expr.property)
}

// globalThis Identifier (AST)
// available only in declarations
type globalThis struct {
	Symbol string
	Pos
}

// node implements Node.
func (expr *globalThis) node() {}

// String implements Node.
func (expr *globalThis) String() string {
	return fmt.Sprintf("Node \x1b[32mglobalThis\x1b[0m { position: %+v }", expr.Pos)
}

// Grouping Expressions (AST)
type GroupingExpr struct {
	exprs []Node
	Pos
}

// node implements Node.
func (expr *GroupingExpr) node() {}

// String implements Node.
func (expr *GroupingExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mGrouping Expression\x1b[0m {\r\n  expressions: %+v\r\n}", expr.exprs)
}

// Increment Expression (AST)
type IncrementExpr struct {
	operand Node
	op      string // (++ | --)
	pre     bool
	Pos
}

// node implements Node.
func (expr *IncrementExpr) node() {}

// String implements Node.
func (expr *IncrementExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mIncrement Expression\x1b[0m {\r\n  operand: %+v\r\n  op: %s\r\n  prefix: %t\r\n}", expr.operand, expr.op, expr.pre)
}

// In Expression (AST)
type InExpr struct {
	left  Node
	right Node
	Pos
}

// node implements Node.
func (expr *InExpr) node() {}

// String implements Node.
func (expr *InExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mIn Expression\x1b[0m {\r\n  left: %+v\r\n  right: %+v\r\n}", expr.left, expr.right)
}

// Member Expression (AST)
type MemberExpr struct {
	object   Node
	property Node
	computed bool
	Pos
}

// node implements Node.
func (expr *MemberExpr) node() {}

// String implements Node.
func (expr *MemberExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mMember Expression\x1b[0m {\r\n  object: %+v\r\n  propery: %+v\r\n  computed: %t\r\n}", expr.object, expr.property, expr.computed)
}

// Void Expression (AST)
type VoidExpr struct {
	operand Node
	Pos
}

// node implements Node.
func (expr *VoidExpr) node() {}

// String implements Node.
func (expr *VoidExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mVoid Expression\x1b[0m {\r\n  operand: %+v\r\n}", expr.operand)
}

// Typeof Expression (AST)
type TypeOfExpr struct {
	operand Node
	Pos
}

// node implements Node.
func (expr *TypeOfExpr) node() {}

// String implements Node.
func (expr *TypeOfExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mTypeof Expression\x1b[0m {\r\n  operand: %+v\r\n}", expr.operand)
}

// Array Literal (AST)
type ArrayLiteral struct {
	elements []Node
	Pos
}

// node implements Node.
func (literal *ArrayLiteral) node() {}

// String implements Node.
func (literal *ArrayLiteral) String() string {
	return fmt.Sprintf("Node \x1b[32mArray Literal\x1b[0m {\r\n  properties: %+v\r\n}", literal.elements)
}

type DynamicNode struct {
	dynamic bool
	node    Node
}

// Object Literal (AST)
type ObjectLiteral struct {
	properties *Map[DynamicNode, Node]
	Pos
}

// node implements Node.
func (literal *ObjectLiteral) node() {}

// String implements Node.
func (literal *ObjectLiteral) String() string {
	return fmt.Sprintf("Node \x1b[32mObject Literal\x1b[0m {\r\n  properties: %+v\r\n}", literal.properties)
}

// Comparison Expression (AST)
type ComparisonExpr struct {
	left  Node
	right Node
	op    string
	Pos
}

// node implements Node.
func (expr *ComparisonExpr) node() {}

// String implements Node.
func (expr *ComparisonExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mComparison Expression\x1b[0m {\r\n left: %+v, right: %+v \r\n  op: %+v\r\n}", expr.left, expr.right, expr.op)
}

// Assignment Expression (AST)
type AssignmentExpr struct {
	left  Node
	right Node
	op    string
	Pos
}

// node implements Node.
func (expr *AssignmentExpr) node() {}

// String implements Node.
func (expr *AssignmentExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mAssignment Expression\x1b[0m {\r\n left: %+v, right: %+v \r\n  op: %+v \r\n}", expr.left, expr.right, expr.op)
}

// Identifiers (AST)
type Identifier struct {
	Symbol string
	Pos
}

// node implements Node.
func (i *Identifier) node() {}

// String implements Node.
func (i *Identifier) String() string {
	return fmt.Sprintf("Node \x1b[32mIdentifier\x1b[0m { symbol: \x1b[34m%s\x1b[0m }", i.Symbol)
}

// Numeric Values (AST)
type Number struct {
	Value float64
	Pos
}

// node implements Node.
func (n *Number) node() {}

// String implements Node.
func (n *Number) String() string {
	return fmt.Sprintf("Node \x1b[32mNumber\x1b[0m { Value: \x1b[33m%v\x1b[0m }", n.Value)
}

// Strings (AST)
type String struct {
	Value string
	Pos
}

// node implements Node.
func (s *String) node() {}

// String implements Node.
func (s *String) String() string {
	return fmt.Sprintf("Node \x1b[32mString\x1b[0m { Value: \x1b[32m%s\x1b[0m }", s.Value)
}

// Binary Expressions (AST)
type BinaryExpr struct {
	left  Node
	right Node
	op    string
	Pos
}

// node implements Node.
func (expr *BinaryExpr) node() {}

// String implements Node.
func (expr *BinaryExpr) String() string {
	return fmt.Sprintf("Node \x1b[32mBinary Expression\x1b[0m {\r\n  left: %+v,\r\n  right: %+v,\r\n  operator: \x1b[33m%s\x1b[0m\r\n}", expr.left, expr.right, expr.op)
}

// #region Methods
func (p *Parser) at(index uint) Token {
	index += p.tokenIndex
	return p.tokens.at(index)
}

func (p *Parser) nextToken() {
	p.tokenIndex++
}

func (p *Parser) expect(typ string) Token {
	tk := p.at(0)
	line, col, count := p.getTkPos(tk)
	tkType := tk.typ
	if tkType != typ {
		p.throwSyntaxError("expected a token of type " + typ + ", but got " + tkType +
			SourceLog(line, col, count, p.sourcePath, ""))
	}
	return p.eat()
}

func (p *Parser) eat() Token {
	tk := p.at(0)
	p.nextToken()
	return tk
}

func (p *Parser) not_eof() bool {
	return p.at(0).typ != TokenType["EOF"]
}

func (p *Parser) throwSyntaxError(message string) {
	throwMessage(SyntaxError(message))
}

// #region Parser

func NewParser(path string, scriptType string, source string) *Parser {
	if !IsAbs(path) {
		path = AbsPath(path)
	}
	var tokens *TokenArray
	if len(source) == 0 {
		source = ReadTextFile(path)
	}
	tokens = Tokenize(source, path)
	return &Parser{
		sourcePath: path,
		tokens:     tokens,
		scriptType: scriptType,
	}
}

func (p *Parser) Parse(main bool) *Program {
	p.program = &Program{
		main:       main,
		scriptType: p.scriptType,
		sourcePath: p.sourcePath,
	}
	for p.not_eof() {
		p.program.body = append(p.program.body, p.parse_stmt())
	}
	return p.program
}

func (p *Parser) parse_stmt() Node {
	switch p.at(0).typ {
	case "if":
		return p.parse_if_stmt()
	case "while", "do":
		return p.parse_while_loop()
	case "throw":
		return p.parse_throw_stmt()
	case "try":
		return p.parse_try_stmt()
	case TokenType["OpenBrace"]:
		return p.parse_block_stmt()
	case "delete":
		return p.parse_delete_stmt()
	case "for":
		return p.parse_for_loop()
	case "function", "async":
		return p.parse_function_decl(false, false, false)
	case "return":
		return p.parse_return_stmt()
	case "break":
		return p.parse_break_stmt()
	case "continue":
		return p.parse_continue_stmt()
	case TokenType["Label"]:
		return p.parse_label()
	case "class":
		return p.parse_class_decl(false)
	case "import":
		return p.parse_import_stmt()
	case "export":
		return p.parse_export_stmt()
	case "switch":
		return p.parse_switch_stmt()
	case
		TokenType["Var"],
		TokenType["Immortal"],
		TokenType["Static"],
		TokenType["Spawn"]:
		return p.parse_var_decl()
	default:
		return p.parse_expr()
	}
}

func (p *Parser) parse_switch_stmt() Node {
	pos := getPosofToken(p.expect("switch"))
	on := p.parse_nested_expr()
	_default := []Node{}
	cases := []Case{}
	p.ExpectOpenBrace()
	for p.not_eof() && p.NotAt(TokenType["CloseBrace"]) {
		if p.IsAt("case") {
			p.eat()
			condition := p.parse_top_expr()
			p.expect(TokenType["Colon"])
			body := p.parse_block()
			_case := Case{
				condition: condition,
				body:      body,
			}
			cases = append(cases, _case)
		} else if p.IsAt("default") {
			p.eat()
			p.expect(TokenType["Colon"])
			_default = p.parse_block()
		} else {
			p.throwUnexpectedTokenError(p.at(0))
		}
	}
	p.ExpectCloseBrace()
	return &SwitchStmt{
		cases: cases,
		on:    on,
		def:   _default,
		Pos:   pos,
	}
}

func (p *Parser) IsAt(typ string) bool {
	return p.at(0).typ == typ
}

func (p *Parser) NotAt(typ string) bool {
	return p.at(0).typ != typ
}

func (p *Parser) ExpectOpenBrace() {
	p.expect(TokenType["OpenBrace"])
}

func (p *Parser) ExpectCloseBrace() {
	p.expect(TokenType["CloseBrace"])
}

func (p *Parser) CurrTkType() string {
	return p.at(0).typ
}

var (
	vardecl_keywords   = NewMap[string, bool]()
	func_decl_keywords = NewMap[string, bool]()
)

func init() {
	vardecl_keywords.set("immortal", true)
	vardecl_keywords.set("spawn", true)
	vardecl_keywords.set("static", true)
	vardecl_keywords.set("var", true)
	func_decl_keywords.set("function", true)
	func_decl_keywords.set("async", true)
}

func (p *Parser) parse_export_stmt() *ExportStmt {
	pos := getPosofToken(p.expect("export"))
	var export Node
	if vardecl_keywords.get(p.at(0).typ) {
		export = p.parse_var_decl()
	} else if func_decl_keywords.get(p.at(0).typ) {
		export = p.parse_function_decl(false, false, false)
	} else if p.at(0).typ == TokenType["OpenBrace"] {
		export = p.parse_object()
	} else if p.at(0).typ == "class" {
		export = p.parse_class_decl(false)
	} else {
		p.throwUnexpectedTokenError(p.at(0))
	}
	return &ExportStmt{
		export: export,
		Pos:    pos,
	}
}

func (p *Parser) parse_import_stmt() *ImportStmt {
	pos := getPosofToken(p.expect("import"))
	path := ""
	namespace := ""
	var names *ObjectLiteral
	var from *FromExpr
	if p.at(0).typ == TokenType["String"] {
		path = p.eat().src
	} else if p.at(0).typ == TokenType["Identifier"] {
		namespace = p.eat().src
		if p.at(0).typ != "from" {
			println("expected the from keyword:")
			p.throwUnexpectedTokenError(p.at(0))
		}
		from = p.parse_from_expr().(*FromExpr)
	} else if p.at(0).typ == TokenType["OpenBrace"] {
		names = p.parse_object_destructuring()
		if p.at(0).typ != "from" {
			println("expected the from keyword:")
			p.throwUnexpectedTokenError(p.at(0))
		}
		from = p.parse_from_expr().(*FromExpr)
	} else {
		println("invalid expression after import keyword")
		p.throwUnexpectedTokenError(p.at(0))
	}
	return &ImportStmt{
		path:      path,
		namespace: namespace,
		names:     names,
		from:      from,
		Pos:       pos,
	}
}

func (p *Parser) parse_object_destructuring() *ObjectLiteral {
	pos := getPosofToken(p.expect(TokenType["OpenBrace"]))
	properties := NewMap[DynamicNode, Node]()
	for p.not_eof() && p.at(0).typ != TokenType["CloseBrace"] {
		var key DynamicNode
		dynamic_key := false
		if p.at(0).typ == TokenType["OpenBracket"] {
			p.eat()
			key.node = p.parse_top_expr()
			p.expect(TokenType["CloseBracket"])
			dynamic_key = true
		} else {
			key.node = p.parse_primary_expr()
		}
		key_pos := getPosFromNode(key.node)
		if !dynamic_key {
			switch key.node.(type) {
			case *Number, *String, *Identifier:
				break
			default:
				p.throwSyntaxError("invalid property key in object literal:" +
					SourceLog(key_pos.line, key_pos.col, key_pos.count, p.sourcePath, ""))
			}
		}
		var value Node
		if p.at(0).typ == TokenType["Colon"] {
			p.eat()
			tk := p.expect(TokenType["Identifier"])
			value = &Identifier{tk.src, getPosofToken(tk)}
		}
		properties.set(key, value)
		if p.at(0).typ != TokenType["CloseBrace"] {
			p.expect(TokenType["Comma"])
		}
	}
	p.expect(TokenType["CloseBrace"])
	return &ObjectLiteral{
		properties: properties,
		Pos:        pos,
	}
}

func (p *Parser) parse_class_decl(expr bool) Node {
	pos := getPosofToken(p.expect("class"))
	anonymous := false
	name := ""
	if expr == false {
		name = p.expect(TokenType["Identifier"]).src
	} else if p.at(0).typ == TokenType["Identifier"] {
		name = p.eat().src
	}
	extends := ""
	if p.at(0).typ == "extends" {
		p.eat()
		extends = p.expect(TokenType["Identifier"]).src
	}
	p.expect(TokenType["OpenBrace"])
	hasConstructor := false
	methods := []*ClassMethod{}
	properties := []*ClassProperty{}
	var constructor *Constructor
	for p.not_eof() && p.at(0).typ != TokenType["CloseBrace"] {
		if prop, ok := p.parse_class_prop(); ok {
			properties = append(properties, prop)
		} else if method, ok := p.parse_class_method(); ok {
			methods = append(methods, method)
		} else if ctor, ok := p.parse_class_ctor(hasConstructor); ok {
			constructor = ctor
		}
	}
	p.expect(TokenType["CloseBrace"])
	return &ClassDecl{
		name:        name,
		properties:  properties,
		methods:     methods,
		extends:     extends,
		constructor: constructor,
		Pos:         pos,
		anonymous:   anonymous,
	}
}

func (p *Parser) parse_class_ctor(hasConstructor bool) (*Constructor, bool) {
	if p.at(0).typ != "constructor" {
		return &Constructor{}, false
	}
	if hasConstructor {
		line, col, count := p.getTkPos(p.at(0))
		p.throwSyntaxError("having mulitiple constructor implementations in one class is not allowed:" + SourceLog(line, col, count, p.sourcePath, ""))
	}
	pos := getPosofToken(p.eat()) // constructor
	p.expect(TokenType["OpenParen"])
	params := []CtorParam{}
	for p.not_eof() && p.at(0).typ != TokenType["CloseParen"] {
		typ := p.at(0).typ
		private := typ == "private"
		public := typ == "public"
		var tk Token
		has_pos := false
		if private || public {
			tk = p.eat()
			has_pos = true
		}
		var pos Pos
		expr := p.parse_arg(true)
		if has_pos {
			pos = getPosFromNode(expr)
		} else {
			pos = getPosofToken(tk)
		}
		params = append(params, CtorParam{
			private: private,
			public:  public,
			expr:    expr,
			Pos:     pos,
		})
		if p.NotAt(TokenType["CloseParen"]) {
			p.expect(TokenType["Comma"])
		}
	}
	p.expect(TokenType["CloseParen"])
	body := p.parse_block()
	return &Constructor{
		name:      "constructor",
		async:     false,
		anonymous: false,
		params:    params,
		_type:     "constructor",
		body:      body,
		Pos:       pos,
	}, true
}

func (p *Parser) parse_class_method() (*ClassMethod, bool) {
	private := false
	static := false
	tk_len := 0
	if is_value(p.at(uint(tk_len)).typ, "private", "public") {
		private = p.at(uint(tk_len)).typ == "private"
		tk_len++
	}
	if p.at(uint(tk_len)).typ != "function" {
		return &ClassMethod{}, false
	}
	pos := Pos{}
	for i := 0; i < tk_len; i++ {
		tk := p.eat()
		if i == 0 {
			pos = getPosofToken(tk)
		}
	}
	decl := *p.parse_function_decl(false, true, false)
	name := decl.name
	// decl.anonymous = true
	return &ClassMethod{
		private: private,
		static:  static,
		name:    name,
		decl:    decl,
		Pos:     pos,
	}, true
}

func (p *Parser) parse_class_prop() (*ClassProperty, bool) {
	private := false
	_default := false
	static := false
	ok := false
	var tk_len uint = 0
	if is_value(p.at(tk_len).typ, "private", "public") {
		private = p.at(tk_len).typ == "private"
		tk_len++
	}
	if is_value(p.at(tk_len).typ, "default") {
		_default = true
		tk_len++
	}
	name := ""
	if p.at(tk_len).typ == TokenType["Identifier"] {
		name = p.at(tk_len).src // identifier
		tk_len++
	}
	if p.at(tk_len).src == "=" {
		ok = true
		tk_len++
	}
	if !ok {
		return &ClassProperty{}, ok
	}
	if len(name) == 0 {
		p.throwUnexpectedTokenError(p.at(tk_len))
	}
	pos := Pos{}
	for i := 0; i < int(tk_len); i++ {
		tk := p.eat()
		if i == 0 {
			pos = getPosofToken(tk)
		}
	}
	value := p.parse_top_expr()
	p.eatSemiColon()
	return &ClassProperty{
		private:  private,
		static:   static,
		_default: _default,
		name:     name,
		value:    value,
		Pos:      pos,
	}, ok
}

func (p *Parser) parse_label() Node {
	tk := p.expect(TokenType["Label"])
	tk.src = strings.Replace(strings.ReplaceAll(tk.src, " ", ""), ":>", "", 1)
	pos := getPosofToken(tk)
	return &Label{
		name: tk.src,
		Pos:  pos,
	}
}

func (p *Parser) parse_continue_stmt() Node {
	return &ContinueStmt{getPosofToken(p.expect("continue"))}
}

func (p *Parser) parse_break_stmt() Node {
	return &BreakStmt{getPosofToken(p.expect("break"))}
}

func (p *Parser) parse_return_stmt() *ReturnStmt {
	pos := getPosofToken(p.expect("return"))
	var value Node
	if p.at(0).typ != TokenType["CloseBrace"] && !p.eatSemiColon() {
		value = p.parse_expr()
	}
	return &ReturnStmt{
		value: value,
		Pos:   pos,
	}
}

func (p *Parser) parse_function_decl(expr bool, method bool, no_leading_keyword bool) *FunctionDecl {
	tk := p.at(0)
	async := false
	if tk.typ == "async" {
		async = true
		p.eat()
		if !no_leading_keyword {
			tk = p.expect("function")
		}
	} else {
		if !no_leading_keyword {
			p.expect("function")
		}
	}
	var name DynamicNode
	anonymous := false
	if method {
		if p.at(0).typ == TokenType["OpenBracket"] {
			p.eat()
			name.node = p.parse_top_expr()
			name.dynamic = true
			p.expect(TokenType["CloseBracket"])
		} else {
			tk := p.expect(TokenType["Identifier"])
			name.node = &Identifier{
				Symbol: tk.src,
				Pos:    getPosofToken(tk),
			}
		}
	} else if no_leading_keyword {
		if p.at(0).typ == TokenType["Identifier"] {
			tk := p.eat()
			name.node = &Identifier{
				Symbol: tk.src,
				Pos:    getPosofToken(tk),
			}
		}
	} else if expr == false {
		tk := p.expect(TokenType["Identifier"])
		name.node = &Identifier{
			Symbol: tk.src,
			Pos:    getPosofToken(tk),
		}
	} else if p.at(0).typ == TokenType["Identifier"] {
		tk := p.eat()
		name.node = &Identifier{
			Symbol: tk.src,
			Pos:    getPosofToken(tk),
		}
	} else {
		anonymous = true
	}
	pos := getPosofToken(tk)
	params := p.parse_args(true)
	body := p.parse_block()
	return &FunctionDecl{
		name:      name,
		async:     async,
		_type:     "",
		body:      body,
		params:    params,
		Pos:       pos,
		anonymous: anonymous,
	}
}

func (p *Parser) parse_args(params bool) []Node {
	p.expect(TokenType["OpenParen"])
	exprs := []Node{}
	for p.at(0).typ != TokenType["CloseParen"] && p.not_eof() {
		p.eatComma()
		expr := p.parse_arg(params)
		exprs = append(exprs, expr)
	}
	p.expect(TokenType["CloseParen"])
	return exprs
}

func (p *Parser) parse_arg(params bool) Node {
	expr := p.parse_restorspread_expr()
	if params {
		switch expr.(type) {
		case *Identifier, *AssignmentExpr, *RestOrSpreadExpr, *ObjectLiteral, *ArrayLiteral:
			break
		default:
			pos := getPosFromNode(expr)
			p.throwSyntaxError("invalid parameter expression, identifier expected" +
				SourceLog(pos.line, pos.col, pos.count, p.sourcePath, ""))
		}
	}
	return expr
}

func (p *Parser) parse_restorspread_expr() Node {
	if p.NotAt("...") {
		return p.parse_nested_expr()
	}
	pos := getPosofToken(p.eat())
	return &RestOrSpreadExpr{
		operand: p.parse_nested_expr(),
		Pos:     pos,
	}
}

func (p *Parser) parse_for_loop() Node {
	pos := getPosofToken(p.expect("for"))
	p.expect(TokenType["OpenParen"])
	if is_value(p.at(0).typ, "spawn", "static", "immortal", "var") {
		return p.parse_for_iterators_loop(pos)
	} else {
		return p.parse_traditional_for_loop(pos)
	}
}

func (p *Parser) parse_traditional_for_loop(pos Pos) Node {
	before := p.parse_nested_expr()
	p.expect(TokenType["SemiColon"])
	condition := p.parse_nested_expr()
	p.expect(TokenType["SemiColon"])
	after := p.parse_nested_expr()
	p.expect(TokenType["CloseParen"])
	body := p.parse_block()
	return &ForLoop{
		before:    before,
		condition: condition,
		after:     after,
		body:      body,
		Pos:       pos,
	}
}

func (p *Parser) parse_for_iterators_loop(pos Pos) Node {
	keyword := p.eat() // keyword
	if !is_value(keyword.typ, "spawn", "var") {
		p.expect(TokenType["Spawn"])
	}
	_type := "mutable"
	switch keyword.src {
	case "immortal":
		_type = "constant"
	case "static":
		_type = "static"
	case "var":
		_type = "var"
	}
	var left Node
	if p.IsAt(TokenType["OpenBrace"]) {
		left = p.parse_object_destructuring()
	} else if p.IsAt(TokenType["OpenBracket"]) {
		left = p.parse_array()
	} else {
		tk := p.expect(TokenType["Identifier"])
		left = &Identifier{tk.src, getPosofToken(tk)}
	}
	if !is_value(p.at(0).typ, "of", "in") {
		p.throwUnexpectedTokenError(p.at(0))
	}
	op := p.eat().src
	right := p.parse_nested_expr()
	p.expect(TokenType["CloseParen"])
	body := p.parse_block()
	return &ForIteratorLoop{
		left:  left,
		right: right,
		_type: _type,
		op:    op,
		body:  body,
		Pos:   pos,
	}
}

func (p *Parser) parse_delete_stmt() *DeleteStmt {
	pos := getPosofToken(p.expect("delete"))
	operand := p.parse_expr()
	switch operand.(type) {
	case *Identifier:
		break
	default:
		operand_pos := getPosFromNode(operand)
		p.throwSyntaxError("the operand of the delete keyword must be a variable or property access" +
			SourceLog(operand_pos.line, operand_pos.col, operand_pos.count, p.sourcePath, ""))
	}
	return &DeleteStmt{operand, pos}
}

func (p *Parser) parse_block_stmt() *BlockStmt {
	pos := getPosofToken(p.at(0)) // open brace
	body := p.parse_block()
	p.eatSemiColon()
	return &BlockStmt{body, pos}
}

func (p *Parser) parse_try_stmt() *TryCatch {
	pos := getPosofToken(p.expect("try")) // try
	try := p.parse_block()
	var catch []Node
	var finally []Node
	var catch_param Node
	if !is_value(p.at(0).typ, "catch", "finally") {
		p.expect("catch")
	}
	if p.at(0).typ == "catch" {
		p.eat()
		if p.at(0).typ == TokenType["OpenParen"] {
			p.eat() // (
			if p.at(0).typ != TokenType["Identifier"] {
				p.throwUnexpectedTokenError(p.at(0))
			}
			catch_param = p.parse_primary_expr()
			p.expect(TokenType["CloseParen"]) // )
		}
		catch = p.parse_block()
	}
	if p.at(0).typ == "finally" {
		p.eat()
		finally = p.parse_block()
	}
	return &TryCatch{
		try:         try,
		catch:       catch,
		finally:     finally,
		catch_param: catch_param,
		Pos:         pos,
	}
}

func (p *Parser) parse_throw_stmt() *ThrowStmt {
	pos := getPosofToken(p.expect("throw"))
	value := p.parse_expr()
	return &ThrowStmt{value, pos}
}

func (p *Parser) parse_while_loop() *WhileLoop {
	do := false
	var pos Pos
	var condition Node
	var body []Node
	if p.at(0).typ == "do" {
		do = true
		pos = getPosofToken(p.eat()) // do-while loop
		body = p.parse_block()
		p.expect("while")
		// condition
		p.expect(TokenType["OpenParen"])
		condition = p.parse_nested_expr()
		p.expect(TokenType["CloseParen"])
	} else {
		// while loop
		pos = getPosofToken(p.expect("while"))
		// condition
		p.expect(TokenType["OpenParen"])
		condition = p.parse_nested_expr()
		p.expect(TokenType["CloseParen"])
		// block
		body = p.parse_block()
	}
	return &WhileLoop{
		condition: condition,
		body:      body,
		do:        do,
		Pos:       pos,
	}
}

func (p *Parser) parse_if_stmt() *IfStmt {
	pos := getPosofToken(p.expect("if")) // if token
	p.expect(TokenType["OpenParen"])
	var condition Node
	if is_value(p.at(0).typ, "spawn", "immortal", "var", "static") {
		condition = p.parse_var_decl()
	} else {
		condition = p.parse_nested_expr()
	}
	p.expect(TokenType["CloseParen"])
	body := p.parse_block()
	var elseBody []Node
	if p.at(0).typ == "else" {
		p.eat() // else
		if p.NotAt(TokenType["OpenBrace"]) {
			elseBody = append(elseBody, p.parse_stmt())
		} else {
			elseBody = p.parse_block()
		}
	}
	return &IfStmt{
		condition: condition,
		body:      body,
		elseBody:  elseBody,
		Pos:       pos,
	}
}

func (p *Parser) parse_block() []Node {
	p.expect(TokenType["OpenBrace"])
	block := []Node{}
	for p.not_eof() && p.at(0).typ != TokenType["CloseBrace"] {
		block = append(block, p.parse_stmt())
	}
	p.expect(TokenType["CloseBrace"])
	return block
}

func (p *Parser) parse_var_decl() *VarDecl {
	keyword := p.eat() // keyword
	if !is_value(keyword.typ, "spawn", "var") {
		p.expect(TokenType["Spawn"])
	}
	_type := "mutable"
	switch keyword.src {
	case "immortal":
		_type = "constant"
	case "static":
		_type = "static"
	case "var":
		_type = "var"
	}
	expr := p.parse_decl_expr()
	return &VarDecl{
		left:  expr.left,
		right: expr.right,
		_type: _type,
	}
}

func (p *Parser) eatSemiColon() bool {
	if p.at(0).typ == TokenType["SemiColon"] {
		p.eat()
		return true
	}
	return false
}

func (p *Parser) parse_decl_expr() *AssignmentExpr {
	var left Node
	if p.IsAt(TokenType["OpenBrace"]) {
		left = p.parse_object_destructuring()
	} else {
		left = p.parse_array()
	}
	switch left.(type) {
	case *Identifier, *ArrayLiteral, *ObjectLiteral:
		break
	default:
		pos := getPosFromNode(left)
		p.throwSyntaxError("invalid left hand side in variable declaration" + SourceLog(pos.line, pos.col, pos.count, p.sourcePath, ""))
	}
	op := p.at(0)
	pos := getPosFromNode(left)
	if op.src != "=" {
		return &AssignmentExpr{left, nil, op.src, pos}
	}
	p.nextToken() // eat op
	right := p.parse_globalThis()
	p.eatSemiColon()
	return &AssignmentExpr{left, right, op.src, pos}
}

// #region Expressions
func (p *Parser) parse_globalThis() Node {
	if p.at(0).typ != "globalThis" {
		return p.parse_top_expr()
	}
	tk := p.eat()
	return &globalThis{
		Symbol: tk.src, // globalThis
		Pos:    getPosofToken(tk),
	}
}

// #region getPosFromNode
func getPosFromNode(node Node) Pos {
	var pos Pos
	switch l := node.(type) {
	case *Identifier:
		pos = l.Pos
	case *Number:
		pos = l.Pos
	case *String:
		pos = l.Pos
	case *BinaryExpr:
		pos = l.Pos
	case *AssignmentExpr:
		pos = l.Pos
	case *VarDecl:
		pos = l.Pos
	case *ArrayLiteral:
		pos = l.Pos
	case *AwaitExpr:
		pos = l.Pos
	case *BlockStmt:
		pos = l.Pos
	case *BreakStmt:
		pos = l.Pos
	case *CallExpr:
		pos = l.Pos
	case *ClassDecl:
		pos = l.Pos
	case *ClassMethod:
		pos = l.Pos
	case *ClassProperty:
		pos = l.Pos
	case *ComparisonExpr:
		pos = l.Pos
	case *Constructor:
		pos = l.Pos
	case *ContinueStmt:
		pos = l.Pos
	case *CtorParam:
		pos = l.Pos
	case *DeleteStmt:
		pos = l.Pos
	case *ForIteratorLoop:
		pos = l.Pos
	case *ForLoop:
		pos = l.Pos
	case *FunctionDecl:
		pos = l.Pos
	case *GotoStmt:
		pos = l.Pos
	case *GroupingExpr:
		pos = l.Pos
	case *IfStmt:
		pos = l.Pos
	case *InExpr:
		pos = l.Pos
	case *IncrementExpr:
		pos = l.Pos
	case *Label:
		pos = l.Pos
	case *MemberExpr:
		pos = l.Pos
	case *NewExpr:
		pos = l.Pos
	case *ObjectLiteral:
		pos = l.Pos
	case *Program:
		pos = l.Pos
	case *ReturnStmt:
		pos = l.Pos
	case *SuperExpr:
		pos = l.Pos
	case *ThrowStmt:
		pos = l.Pos
	case *TryCatch:
		pos = l.Pos
	case *TypeOfExpr:
		pos = l.Pos
	case *VoidExpr:
		pos = l.Pos
	case *WhileLoop:
		pos = l.Pos
	case *globalThis:
		pos = l.Pos
	case *globalThisMember:
		pos = l.Pos
	case *globalThisMemberAssignment:
		pos = l.Pos
	case *LogicalExpr:
		pos = l.Pos
	case *FromExpr:
		pos = l.Pos
	case *ImportStmt:
		pos = l.Pos
	case *ExportStmt:
		pos = l.Pos
	case *MatchExpr:
		pos = l.Pos
	case *SwitchStmt:
		pos = l.Pos
	case *RestOrSpreadExpr:
		pos = l.Pos
	case *TemplateString:
		pos = l.Pos
	case *InstanceofExpr:
		pos = l.Pos
	case *TernaryExpr:
		pos = l.Pos
	default:
		panic(fmt.Sprintf("unexpected main.Node: %#v", l))
	}
	return pos
}

// #endregion

func (p *Parser) parse_expr() Node {
	expr := p.parse_assignment_expr()
	p.eatSemiColon()
	return expr
}

func (p *Parser) parse_nested_expr() Node {
	expr := p.parse_assignment_expr()
	return expr
}

func (p *Parser) parse_assignment_expr() Node {
	left := p.parse_top_expr()
	pos := getPosFromNode(left)
	if p.at(0).typ != TokenType["AssignmentOp"] {
		return left
	}
	op := p.expect(TokenType["AssignmentOp"])
	right := p.parse_nested_expr()
	return &AssignmentExpr{left, right, op.src, pos}
}

func (p *Parser) parse_top_expr() Node {
	return p.parse_ternary_expr()
}

func (p *Parser) parse_ternary_expr() Node {
	condition := p.parse_match_expr()
	if p.NotAt("?") {
		return condition
	}
	p.eat()
	then := p.parse_match_expr()
	p.expect(TokenType["Colon"])
	_else := p.parse_match_expr()
	return &TernaryExpr{
		condition: condition,
		then:      then,
		_else:     _else,
		Pos:       getPosFromNode(condition),
	}
}

func (p *Parser) parse_match_expr() Node {
	if p.at(0).typ != "match" {
		return p.parse_from_expr()
	}
	pos := getPosofToken(p.eat())
	cases := []Match{}
	match_against := p.parse_top_expr()
	p.expect(TokenType["OpenBrace"])
	for p.not_eof() && p.at(0).typ != TokenType["CloseBrace"] {
		_case := Match{}
		_case.match = p.parse_nested_expr()
		p.expect(TokenType["Arrow"])
		if p.at(0).typ == TokenType["OpenBrace"] {
			_case.body = p.parse_block_stmt()
		} else {
			_case.body = p.parse_nested_expr()
		}
		cases = append(cases, _case)
	}
	p.expect(TokenType["CloseBrace"])
	return &MatchExpr{
		cases: cases,
		match: match_against,
		Pos:   pos,
	}
}

func (p *Parser) parse_from_expr() Node {
	if p.at(0).typ != "from" {
		return p.parse_logical_expr()
	}
	pos := getPosofToken(p.eat())
	path := p.expect(TokenType["String"]).src
	return &FromExpr{
		path: path,
		Pos:  pos,
	}
}

func (p *Parser) parse_logical_expr() Node {
	if p.at(0).src == "!" {
		tk := p.eat()
		pos := getPosofToken(tk)
		op := tk.src
		left := p.parse_instanceof_expr()
		return &LogicalExpr{
			left:  left,
			right: nil,
			op:    op,
			Pos:   pos,
		}
	}
	left := p.parse_instanceof_expr()
	if p.at(0).typ != TokenType["LogicalOp"] {
		return left
	}
	op := p.eat().src
	right := p.parse_logical_expr()
	pos := getPosFromNode(left)
	return &LogicalExpr{
		left:  left,
		right: right,
		op:    op,
		Pos:   pos,
	}
}

func (p *Parser) parse_instanceof_expr() Node {
	left := p.parse_super_expr()
	if p.NotAt("instanceof") {
		return left
	}
	p.eat()
	right := p.parse_super_expr()
	return &InstanceofExpr{
		left:  left,
		right: right,
		Pos:   getPosFromNode(left),
	}
}

func (p *Parser) parse_super_expr() Node {
	if p.at(0).typ != "super" {
		return p.parse_await_expr()
	}
	pos := getPosofToken(p.eat())
	args := p.parse_args(false)
	return &SuperExpr{
		args: args,
		Pos:  pos,
	}
}

func (p *Parser) parse_await_expr() Node {
	if p.at(0).typ != "await" {
		return p.parse_import_expr()
	}
	pos := getPosofToken(p.eat())
	return &AwaitExpr{
		operand: p.parse_import_expr(),
		Pos:     pos,
	}
}

func (p *Parser) parse_import_expr() Node {
	if p.at(0).typ != "import" {
		return p.parse_new_expr()
	}
	pos := getPosofToken(p.eat())
	p.expect(TokenType["OpenParen"])
	specifier := p.parse_nested_expr()
	p.expect(TokenType["CloseParen"])
	return &DynamicImport{
		specifier: specifier,
		Pos:       pos,
		async:     true,
	}
}

func (p *Parser) parse_new_expr() Node {
	if p.at(0).typ != "new" {
		return p.parse_class_expr()
	}
	pos := getPosofToken(p.eat())
	operand := p.parse_class_expr()
	return &NewExpr{
		operand: operand,
		Pos:     pos,
	}
}

func (p *Parser) parse_class_expr() Node {
	if p.NotAt("class") {
		return p.parse_fn_expr()
	}
	return p.parse_class_decl(true)
}

func (p *Parser) parse_fn_expr() Node {
	if p.at(0).typ != "function" {
		return p.parse_in_expr()
	}
	return p.parse_function_decl(true, false, false)
}

func (p *Parser) parse_in_expr() Node {
	left := p.parse_comparison_expr()
	if p.at(0).typ != "in" {
		return left
	}
	p.eat()
	pos := getPosFromNode(left)
	right := p.parse_comparison_expr()
	return &InExpr{left, right, pos}
}

func (p *Parser) parse_comparison_expr() Node {
	left := p.parse_additive_expr()
	if p.at(0).typ != TokenType["ComparisonOp"] {
		return left
	}
	op := p.expect(TokenType["ComparisonOp"]).src
	right := p.parse_comparison_expr()
	return &ComparisonExpr{
		left:  left,
		right: right,
		op:    op,
		Pos:   getPosFromNode(left),
	}
}

func getPosofToken(token Token) Pos {
	col := token.col
	line := token.line
	count := token.end - col
	return Pos{line, col, count}
}

func (p *Parser) parse_additive_expr() Node {
	left := p.parse_multiplicative_expr()
	for p.at(0).src == "+" ||
		p.at(0).src == "-" {
		op := p.eat().src
		right := p.parse_multiplicative_expr()
		left = &BinaryExpr{
			left,
			right,
			op,
			getPosFromNode(left),
		}
		if p.eatSemiColon() {
			return left
		}
	}
	return left
}

func (p *Parser) parse_multiplicative_expr() Node {
	left := p.parse_member_expr()
	for is_value(p.at(0).src, "*", "/", "%", "**") {
		op := p.eat().src
		right := p.parse_member_expr()
		left = &BinaryExpr{
			left,
			right,
			op,
			getPosFromNode(left),
		}
		if p.eatSemiColon() {
			return left
		}
	}
	return left
}

func (p *Parser) parse_member_expr() Node {
	object := p.parse_call_expr(nil)
	for p.not_eof() && is_value(p.at(0).typ, TokenType["Dot"], TokenType["OpenBracket"]) {
		computed := false
		tk := p.eat() // dot (.) or bracket ([)
		if tk.typ == TokenType["OpenBracket"] {
			computed = true
		}
		var property Node
		if computed {
			property = p.parse_nested_expr()
		} else {
			property = p.parse_call_expr(nil)
		}
		call := false
		switch property := property.(type) {
		case *Identifier:
			break
		case *CallExpr:
			if !computed {
				object = &CallExpr{
					args: property.args,
					caller: &MemberExpr{
						property: property.caller,
						object:   object,
						computed: computed,
						Pos:      getPosFromNode(object),
					},
					Pos: getPosFromNode(property),
				}
				call = true
			}
		default:
			if !computed {
				pos := getPosFromNode(property)
				p.throwSyntaxError("invalid property access, identifier expected: " +
					SourceLog(pos.line, pos.col, pos.count, p.sourcePath, ""))
			}
		}
		pos := getPosFromNode(object)
		if !call {
			object = &MemberExpr{
				object:   object,
				property: property,
				computed: computed,
				Pos:      pos,
			}
		}
		if computed {
			p.expect(TokenType["CloseBracket"])
		}
	}
	return object
}

func (p *Parser) parse_call_expr(caller Node) Node {
	if caller == nil {
		caller = p.parse_object()
	}
	if p.at(0).typ == TokenType["OpenParen"] {
		args := p.parse_args(false)
		pos := getPosFromNode(caller)
		expr := &CallExpr{
			caller: caller,
			args:   args,
			Pos:    pos,
		}
		return expr
	}
	return caller
}

func (p *Parser) parse_object() Node {
	if p.at(0).typ != TokenType["OpenBrace"] {
		return p.parse_array()
	}
	pos := getPosofToken(p.eat()) // open brace
	object := &ObjectLiteral{NewMap[DynamicNode, Node](), pos}
	for p.not_eof() && p.at(0).typ != TokenType["CloseBrace"] {
		var key Node
		dynamic_key := false
		if p.at(0).typ == TokenType["OpenBracket"] {
			p.eat()
			key = p.parse_top_expr()
			p.expect(TokenType["CloseBracket"])
			dynamic_key = true
		} else {
			key = p.parse_primary_expr()
		}
		key_pos := getPosFromNode(key)
		if !dynamic_key {
			switch key.(type) {
			case *Number, *String, *Identifier:
				break
			default:
				p.throwSyntaxError("invalid property key in object literal:" +
					SourceLog(key_pos.line, key_pos.col, key_pos.count, p.sourcePath, ""))
			}
		}
		var value Node
		if p.IsAt(TokenType["OpenParen"]) {
			decl := p.parse_function_decl(false, false, true)
			if dynamic_key {
				decl.name.dynamic = true
			}
			decl.name.node = key
			value = decl
		} else {
			if !p.IsAt(TokenType["Comma"]) &&
				!p.IsAt(TokenType["CloseBrace"]) {
				p.expect(TokenType["Colon"])
				value = p.parse_nested_expr()
			}
		}
		object.properties.set(DynamicNode{
			dynamic: dynamic_key,
			node:    key,
		}, value)
		p.eatComma()
	}
	p.expect(TokenType["CloseBrace"])
	return object
}

func SourceLog(line, col, count int, path, source string) string {
	return SourceWithinRange(path, line, col, count, source) +
		SourceAtPosition(path, line, col)
}

// eats the current token if it is a comma and returns true, else just return false
func (p *Parser) eatComma() bool {
	if p.at(0).typ == TokenType["Comma"] {
		p.eat()
		return true
	}
	return false
}

func (p *Parser) parse_array() Node {
	if p.at(0).typ != TokenType["OpenBracket"] {
		return p.parse_primary_expr()
	}
	pos := getPosofToken(p.expect(TokenType["OpenBracket"]))
	elements := []Node{}
	for p.not_eof() && p.at(0).typ != TokenType["CloseBracket"] {
		elements = append(elements, p.parse_nested_expr())
		if p.at(0).typ != TokenType["CloseBracket"] {
			p.expect(TokenType["Comma"])
		}
	}
	p.expect(TokenType["CloseBracket"])
	return &ArrayLiteral{elements, pos}
}

func (p *Parser) parse_primary_expr() Node {
	pos := getPosofToken(p.at(0))
	switch p.at(0).typ {
	case TokenType["Number"]:
		float, _ := strconv.ParseFloat(p.eat().src, 64)
		return &Number{float, pos}
	case TokenType["String"]:
		return &String{p.eat().src, pos}
	case TokenType["Identifier"]:
		var expr Node = &Identifier{p.eat().src, pos}
		if is_value(p.at(0).typ, TokenType["IncreOp"], TokenType["DecreOp"]) {
			var op string = p.eat().src // (++ | --)
			return &IncrementExpr{
				operand: expr,
				op:      op,
				pre:     false,
				Pos:     pos,
			}
		}
		return expr
	case "typeof":
		pos := getPosofToken(p.eat())
		operand := p.parse_object()
		return &TypeOfExpr{operand, pos}
	case "void":
		pos := getPosofToken(p.eat())
		operand := p.parse_object()
		return &VoidExpr{operand, pos}
	case TokenType["IncreOp"], TokenType["DecreOp"]:
		var op string = p.eat().src // (++ | --)
		var operand Node = p.parse_nested_expr()
		return &IncrementExpr{
			operand: operand,
			op:      op,
			pre:     true,
			Pos:     pos,
		}
	case TokenType["OpenParen"]:
		pos := getPosofToken(p.eat())
		if p.at(0).typ == TokenType["CloseParen"] &&
			p.at(1).typ == TokenType["Arrow"] {
			p.eat() // )
			p.eat() //=>
			body := p.parse_block()
			return &FunctionDecl{
				name: struct {
					dynamic bool
					node    Node
				}{},
				async:     false,
				anonymous: true,
				_type:     "arrow",
				body:      body,
				params:    []Node{},
				Pos:       pos,
			}
		}
		exprs := []Node{p.parse_nested_expr()}
		for p.at(0).typ == TokenType["Comma"] {
			p.eatComma()
			exprs = append(exprs, p.parse_nested_expr())
		}
		p.expect(TokenType["CloseParen"])
		if p.at(0).typ == TokenType["Arrow"] {
			p.eat()
			body := p.parse_block()
			return &FunctionDecl{
				name: struct {
					dynamic bool
					node    Node
				}{},
				async:     false,
				anonymous: true,
				_type:     "arrow",
				body:      body,
				params:    exprs,
				Pos:       pos,
			}
		}
		return &GroupingExpr{
			exprs: exprs,
			Pos:   pos,
		}
	case "globalThis":
		if p.at(1).typ != TokenType["Dot"] {
			p.throwUnexpectedTokenError(p.at(0))
		}
		tk := p.eat()
		pos := getPosofToken(tk)
		p.eat() // dot
		symbol := tk.src
		property := p.expect(TokenType["Identifier"]).src
		member := &globalThisMember{
			Symbol:   symbol,
			property: property,
			Pos:      pos,
		}
		if p.at(0).typ == TokenType["AssignmentOp"] {
			op := p.eat().src
			rhs := p.parse_expr()
			return &globalThisMemberAssignment{
				Symbol:   symbol,
				property: property,
				right:    rhs,
				op:       op,
				Pos:      pos,
			}
		}
		return member
	case TokenType["TString"]:
		str := []Node{}
		tk := p.eat()
		exp := regexp.MustCompile(`(\#\{)`)
		locs := exp.FindAllStringIndex(tk.src, -1)
		substrs := []string{}
		for i := 0; i < len(locs); i++ {
			loc := locs[i]
			exp := regexp.MustCompile(`(\})`)
			to := exp.FindStringIndex(tk.src)[0]
			from := tk.src[loc[0]+2:]
			substrs = append(substrs, from[:to])
		}
		new_string := tk.src
		for i := 0; i < len(substrs); i++ {
			substr := substrs[i]
			new_string = strings.Replace(new_string, substr, "#{...}", 1)
		}
		exp2 := regexp.MustCompile(`(\#\{\.\.\.\})`)
		locs2 := exp2.FindAllStringIndex(tk.src, -1)
		pos2 := 0
		for i := pos2; i < len(locs2); i++ {
			upto := locs2[i][0]
			pos2 = locs2[i][1]
			str = append(str, &String{
				Value: new_string[upto:],
				Pos:   pos,
			})
		}
		println(str)
		return &TemplateString{
			Pos: pos,
			str: str,
		}

	default:
		p.throwUnexpectedTokenError(p.at(0))
	}
	return nil
}

func (p *Parser) throwUnexpectedTokenError(tk Token) {
	line, pos, count := p.getTkPos(tk)
	str := ""
	if tk.typ == TokenType["EOF"] {
		str = TokenType["EOF"]
	}
	p.throwSyntaxError(
		"unexpected token reached: " + str + SourceLog(line, pos, count, p.sourcePath, ""),
	)
}

// get the line column, and length of the token
func (p *Parser) getTkPos(tk Token) (int, int, int) {
	line, pos := tk.line, tk.col
	count := tk.end - pos
	return line, pos, count
}
