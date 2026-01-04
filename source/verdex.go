package main

import "regexp"

// type Lexer struct {
// 	position, column, line int
// 	source                 string
// 	path                   string
// }

// func (l *Lexer) Tokenize(path, source string) *TokenArray {
// 	tokens := tokenArray()
// 	var TokenSpecs []Spec = []Spec{}
// 	return tokens
// }

// var routeTable = NewMap[string, *Component]()
var routeTable = NewMap[string, string]()

// func resetRouteTable() {
// 	routeTable.forEach(func(key string, _ *Component) {
// 		routeTable.delete(key)
// 	})
// }

type ASXNode interface {
	asx()
}

type Attr struct {
	name  string
	value Node
}

type Element struct {
	tagName  string
	attr     *Map[string, Attr]
	children []ASXNode
}

// asx implements ASXNode.
func (*Element) asx() {}

type Text struct {
	text string
}

// asx implements ASXNode.
func (*Text) asx() {}

type Interpolation struct {
	expr Node
}

// asx implements ASXNode.
func (*Interpolation) asx() {}

// export component Name (props, children)
// {
//   $ as code
//   spawn cool = true
// }
// ```
// <!-- template -->
// <div>${cool}$</div>
// ${children()}$
// ```

// Component (AST)
type Component struct {
	export   bool
	name     string
	as_code  []Node
	template ASXNode
	props    string
	children string
	Pos
}

// // asx implements ASXNode.
// func (*Component) asx() {}

// <Name prop=${0}$>children...</Name>
// <Name prop=${0}$/>

// ComponentCall (AST)
type ComponentCall struct {
	name     string
	props    []Attr
	children []ASXNode
	Pos
}

// asx implements ASXNode.
func (*ComponentCall) asx() {}

type ASXModule struct {
	imports    []*ImportStmt
	components *Map[string, string]
	main       bool
	path       string
	parser     *Parser
}

type ASXParser struct{}

func (p *ASXParser) Parse(path string, main bool) *ASXModule {
	parser := NewParser(path, "module", "")
	module := &ASXModule{
		imports:    []*ImportStmt{},
		components: NewMap[string, string](),
		main:       main,
		path:       parser.sourcePath,
		parser:     parser,
	}
	for parser.not_eof() {
		module.parse_stmt(p)
	}
	return module
}

func NewASXParser() *ASXParser {
	return &ASXParser{}
}

func (m *ASXModule) parse_stmt(p *ASXParser) {
	switch m.parser.at(0).src {
	case "import":
		m.imports = append(m.imports, m.parser.parse_import_stmt())
	case "export":
		m.parser.eat()
		comp := m.parse_component(p)
		comp.export = true
	case "route":
		m.parse_route()
	case "component":
		m.parse_component(p)
	default:
		m.parser.throwUnexpectedTokenError(m.parser.at(0))
	}
}

func (m *ASXModule) parse_route() {
	p := m.parser
	m.expect("route")
	route := p.expect(TokenType["String"]).src
	tk := p.expect(TokenType["Identifier"])
	name := tk.src
	component := m.components.get(name)
	if len(component) == 0 {
		line, col, count := p.getTkPos(tk)
		p.throwSyntaxError("component " + name + " does not exist" + SourceLog(line, col, count, m.path, ""))
	}
	routeTable.set(route, sprintf("\r\nif (window.location.pathname == \"%s\") { $('body').html(%s().render()); }", route, name))
}

func (m *ASXModule) parse_component(asxp *ASXParser) *Component {
	p := m.parser
	pos := getPosofToken(m.expect("component"))
	name := p.expect(TokenType["Identifier"]).src
	props, children := "", ""
	if p.IsAt(TokenType["OpenParen"]) {
		props = p.expect(TokenType["Identifier"]).src
		if p.NotAt(TokenType["CloseParen"]) {
			p.expect(TokenType["Comma"])
			children = p.expect(TokenType["Identifier"]).src
		}
		p.expect(TokenType["CloseParen"])
	}
	as_code := []Node{}
	if p.IsAt(TokenType["OpenBrace"]) {
		as_code = m.parse_block()
	}
	template := m.parse_template()
	comp := &Component{
		export:   false,
		name:     name,
		as_code:  as_code,
		template: template,
		props:    props,
		children: children,
		Pos:      pos,
	}
	m.components.set(comp.name, asxp.CompileComponent(comp))
	return comp
}

func (m *ASXModule) parse_template() ASXNode {
	p := m.parser
	p.expect("```")
	node := m.parse_asx_node()
	p.expect("```")
	return node
}

func (m *ASXModule) parse_asx_node() ASXNode {
	p := m.parser
	_1st := p.at(0)
	_2nd := p.at(1)
	isTag := _1st.src == "<" && _2nd.typ == TokenType["Identifier"]
	if p.at(0).typ == "${" {
		p.eat()
		expr := p.parse_expr()
		p.expect("}$")
		return &Interpolation{
			expr: expr,
		}
	} else if isTag && _2nd.src[0] >= 65 && _2nd.src[0] <= 90 {
		pos := getPosofToken(p.eat()) // '<'
		name := p.eat().src
		props := []Attr{}
		children := []ASXNode{}
		if p.at(0).src == "/" {
			p.eat()
			m.expect(">")
		} else {
			m.expect(">")
			for !(p.at(0).src == "<" && p.at(1).src == "/") {
				children = append(children, m.parse_asx_node())
			}
			m.parse_close_tag(name)
		}
		return &ComponentCall{
			name:     name,
			props:    props,
			children: children,
			Pos:      pos,
		}
	} else if isTag && _2nd.src[0] >= 97 && _2nd.src[0] <= 122 {
		tagName, attributes := m.parse_open_tag()
		children := []ASXNode{}
		for !(p.at(0).src == "<" && p.at(1).src == "/") {
			children = append(children, m.parse_asx_node())
		}
		m.parse_close_tag(tagName)
		return &Element{
			tagName:  tagName,
			attr:     attributes,
			children: children,
		}
	} else if m.isText(0) {
		text := ""
		for p.not_eof() && m.isText(0) {
			text += p.eat().src
		}
		return &Text{
			text: text,
		}
	}
	panic("unhandled node")
}

func (m *ASXModule) isText(i uint) bool {
	src := m.parser.at(i).src
	single_tk_exp := regexp.MustCompile("(\\`\\`\\`|\\$\\{|\\}\\$)")
	if single_tk_exp.MatchString(src) {
		return false
	}
	if m.parser.tokens.length > (i + 1) {
		src += m.parser.at(i + 1).src
	}
	exp := regexp.MustCompile(`(\<\w+|\<\/\w+)`)
	match := exp.MatchString(src)
	return !match
}

func (m *ASXModule) parse_open_tag() (string, *Map[string, Attr]) {
	p := m.parser
	m.expect("<")
	tagName := "div"
	tagName = p.expect(TokenType["Identifier"]).src
	attributes := &Map[string, Attr]{}
	m.expect(">")
	return tagName, attributes
}

func (m *ASXModule) parse_close_tag(tagName string) {
	p := m.parser
	m.expect("<")
	m.expect("/")
	close_tag := p.expect(TokenType["Identifier"])
	if tagName != close_tag.src {
		line, col, count := p.getTkPos(close_tag)
		p.throwSyntaxError("unclosed " + tagName + " tag" + SourceLog(line, col, count, m.path, ""))
	}
	for p.at(0).src != ">" && p.not_eof() {
		p.eat()
	}
	m.expect(">")
}

func (m *ASXModule) expect(s string) Token {
	p := m.parser
	tk := p.at(0)
	line, col, count := p.getTkPos(tk)
	src := tk.src
	if src != s {
		p.throwSyntaxError("expected a token " + s + ", but got " + src +
			SourceLog(line, col, count, p.sourcePath, ""))
	}
	return p.eat()
}

func (m *ASXModule) parse_block() []Node {
	return m.parser.parse_block()
}

func (p *ASXParser) CompileASX(m *ASXModule) string {
	compiled := ""
	for _, stmt := range m.imports {
		compiled += p.CompileImport(stmt)
	}
	m.components.forEach(func(_ string, value string) {
		compiled += value
	})
	routeTable.forEach(func(_, value string) {
		compiled += value
	})
	return compiled
}

func (p *ASXParser) CompileComponent(comp *Component) string {
	compiled := sprintf("\r\nfunction %s(props, children) {\r\n", comp.name)
	for i := 0; i < len(comp.as_code); i++ {
		code := comp.as_code[i]
		compiled += p.CompileAS(code)
	}
	template, bindings := p.CompileTemplate(comp.template)
	render := sprintf("\r\nreturn { render() {\r\nreturn `%s`; }\r\n}\r\n}", template)
	return sprintf("%s%s%s", compiled, bindings, render)
}

func (p *ASXParser) CompileTemplate(template ASXNode) (string, string) {
	compiled := ""
	bindings := ""
	switch t := template.(type) {
	case *ComponentCall:
		compiled = "${" + t.name + "().render()}"
	case *Element:
		compiled = "<" + t.tagName + ">"
		for i := 0; i < len(t.children); i++ {
			template, b := p.CompileTemplate(t.children[i])
			compiled += template
			bindings += b
		}
		compiled += "</" + t.tagName + ">"
	case *Text:
		compiled = t.text
	case *Interpolation:
		id := GenerateRadix(10)
		bindings += sprintf("\r\n__vdx_bind(\"%s\", () => %s)", id, p.CompileAS(t.expr))
		compiled = sprintf(`<reactive data-vx="%s"></reactive>`, id)
	default:
		panic(sprintf("unexpected main.ASXNode: %#v", t))
	}
	return compiled, bindings
}

func (p *ASXParser) CompileAS(code Node) string {
	compiled := ""

	switch code := code.(type) {

	case *ArrayLiteral:
	case *AssignmentExpr:
	case *AwaitExpr:
	case *BinaryExpr:
	case *BlockStmt:
	case *BreakStmt:
	case *CallExpr:
	case *ClassDecl:
	case *ClassMethod:
	case *ClassProperty:
	case *ComparisonExpr:
	case *Constructor:
	case *ContinueStmt:
	case *CtorParam:
	case *DeleteStmt:
	case *DynamicImport:
	case *ExportStmt:
	case *ForIteratorLoop:
	case *ForLoop:
	case *FromExpr:
	case *FunctionDecl:
	case *GotoStmt:
	case *GroupingExpr:
	case *Identifier:
		compiled = code.Symbol
	case *IfStmt:
	case *ImportStmt:
	case *InExpr:
	case *IncrementExpr:
	case *InstanceofExpr:
	case *Label:
	case *LogicalExpr:
	case *MatchExpr:
	case *MemberExpr:
	case *NewExpr:
	case *Number:
		compiled = sprint(code.Value)
	case *ObjectLiteral:
	case *Program:
	case *RestOrSpreadExpr:
	case *ReturnStmt:
	case *String:
		compiled = "\"" + code.Value + "\""
	case *SuperExpr:
	case *SwitchStmt:
	case *TemplateString:
	case *TernaryExpr:
	case *ThrowStmt:
	case *TryCatch:
	case *TypeOfExpr:
	case *VarDecl:
		keyword := map[string]string{
			"mutable":  "let",
			"constant": "const",
			"var":      "var",
			"static":   "const",
		}
		compiled += sprintf("%s %s = %s;", keyword[code._type], p.CompileAS(code.left), p.CompileAS(code.right))
	case *VoidExpr:
	case *WhileLoop:
	case *globalThis:
		compiled += code.Symbol
	case *globalThisMember:
	case *globalThisMemberAssignment:
	default:
		panic(sprintf("unexpected main.Node: %#v", code))
	}
	return compiled
}

func (p *ASXParser) CompileImport(stmt *ImportStmt) string {
	panic("unimplemented")
}
