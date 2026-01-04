package main

import (
	"fmt"
	"regexp"
)

var TokenType = map[string]string{
	"EOF":         "<eof>",
	"White Space": "white-space",
	// "EOF": "End of File",
	// literals
	"Number":     "number",
	"String":     "string",
	"TString":    "template-string",
	"Identifier": "identifier",
	"Label":      "label",
	// operators
	"BinaryOp":     "binary-operator",
	"Arrow":        "arrow",               // =>
	"AssignmentOp": "assignment-operator", // =, +=, -= *= /= ??=
	"ComparisonOp": "comparison-operator", // ==, ===, <=, >= < >
	"LogicalOp":    "logical-operator",    // ==, ===, <=, >= < >
	"IncreOp":      "increment-operator",  // ++
	"DecreOp":      "decrement-operator",  // --
	"OpenParen":    "open-parenthesis",    // (
	"CloseParen":   "close-parenthesis",   // )
	"OpenBracket":  "open-bracket",        // [
	"CloseBracket": "close-bracket",       // ]
	"OpenBrace":    "open-brace",          // {
	"CloseBrace":   "close-brace",         // }
	"Colon":        "colon",               // :
	"SemiColon":    "semi-colon",          // ;
	"Dot":          "dot",                 // .
	"Comma":        "comma",               // ,
	"DQuote":       "double-quote",        // "
	"SQuote":       "single-quote",        // '
	"BTick":        "back-tick",           // `
	//
	"Comment": "comment", // $ ...

	"Spawn":    "spawn",
	"Immortal": "immortal",
	"Static":   "static",
	"Var":      "var",
}

// const (
// 	CLASS
// 	VAR
// 	FUNC
// 	IF
// 	ELSE
// 	WHILE
// 	PRINT
// 	NEW
// )

type Token struct {
	src  string
	typ  string
	line int
	col  int
	end  int
}

func (t Token) String() string {
	return fmt.Sprintf("\x1b[32mToken\x1b[0m {\ntype: \x1b[32m%s\x1b[0m,\nsrc: \x1b[32m%s\x1b[0m,\nline: \x1b[33m%d\x1b[0m,\ncolumn: \x1b[33m%d\x1b[0m,\nend: \x1b[33m%d\x1b[0m\n}", t.typ, t.src, t.line, t.col, t.end)
}

type Spec struct {
	exp *regexp.Regexp
	typ string
}

var TokenSpecs []Spec = []Spec{
	// keywords
	// {regexp.MustCompile(`^print`), TokenType["Print"]},

	// literals
	{regexp.MustCompile(`^[-]?\d+(\.\d+)?\b`), TokenType["Number"]},
	{regexp.MustCompile(`"`), TokenType["DQuote"]},
	{regexp.MustCompile(`'`), TokenType["SQuote"]},
	{regexp.MustCompile("```"), "```"},
	{regexp.MustCompile(`\$\{`), "${"},
	{regexp.MustCompile(`\}\$`), "}$"},
	{regexp.MustCompile("`"), TokenType["BTick"]},
	{regexp.MustCompile(`0(b|B)([01]+|([01_][01])*)`), TokenType["Number"]},
	{regexp.MustCompile(`0(o|O)[0-7]+[0-7_]*`), TokenType["Number"]},
	{regexp.MustCompile(`0(x|X)[0-9a-fA-F]+[0-9a-fA-F_]*`), TokenType["Number"]},
	{regexp.MustCompile(`^([a-zA-Z_#]+[a-zA-Z0-9_#]*)(\ )*(:\>)`), TokenType["Label"]},
	{regexp.MustCompile(`^[a-zA-Z_#]+[a-zA-Z0-9_#]*`), TokenType["Identifier"]},
	//
	{regexp.MustCompile(`\s`), TokenType["White Space"]},

	// operators
	{regexp.MustCompile(`\=\>`), TokenType["Arrow"]},
	{regexp.MustCompile(`^(===|==|!==|!=|\>=|\<=|\>|\<)`), TokenType["ComparisonOp"]},
	{regexp.MustCompile(`^(\=|\+=|\-=|\*=|\/=|\%=|\?\?=)`), TokenType["AssignmentOp"]},
	{regexp.MustCompile(`^(\+\+)`), TokenType["IncreOp"]},
	{regexp.MustCompile(`^(\-\-)`), TokenType["DecreOp"]},
	{regexp.MustCompile(`^(\?)`), "?"},
	{regexp.MustCompile(`^(\.\.\.)`), "..."},
	{regexp.MustCompile(`^(\+|\-|/|\%|\*\*|\*)`), TokenType["BinaryOp"]},
	{regexp.MustCompile(`^(\&\&|\|\||\!)`), TokenType["LogicalOp"]},
	{regexp.MustCompile(`^\(`), TokenType["OpenParen"]},
	{regexp.MustCompile(`^\)`), TokenType["CloseParen"]},
	{regexp.MustCompile(`^\{`), TokenType["OpenBrace"]},
	{regexp.MustCompile(`^\}`), TokenType["CloseBrace"]},
	{regexp.MustCompile(`^\[`), TokenType["OpenBracket"]},
	{regexp.MustCompile(`^\]`), TokenType["CloseBracket"]},
	{regexp.MustCompile(`^:`), TokenType["Colon"]},
	{regexp.MustCompile(`^;`), TokenType["SemiColon"]},
	{regexp.MustCompile(`^\.`), TokenType["Dot"]},
	{regexp.MustCompile(`^,`), TokenType["Comma"]},
	// {regexp.MustCompile(`^\$.*(?:\$|$)`), TokenType["Comment"]},
}

func Tokenize(source string, path string) *TokenArray {
	tokens := tokenArray()
	position, line, column := 0, 1, 1

top:
	for position < len(source) {
		matched := false
		remaining := source[position:]
		var match struct {
			length int
			typ    string
			src    string
		}
		// loop over token types to find a match
		for i := 0; i < len(TokenSpecs); i++ {
			pattern := TokenSpecs[i]
			loc := pattern.exp.FindStringIndex(remaining)
			if loc != nil && loc[0] == 0 {
				match.length = loc[1]
				match.typ = pattern.typ
				match.src = remaining[:loc[1]]
				matched = true
				if is_value(match.typ, TokenType["DQuote"], TokenType["SQuote"]) {
					quote := source[position]
					position++ // eat character
					// remaining = source[position:]
					_string := ""
					for position < len(source) && source[position] != quote {
						char := source[position]
						position++ // eat character
						_string += string(char)
						if char == '\\' {
							_string += string(source[position])
							position++ // eat character
						}
					}
					if position >= len(source) || source[position] != quote {
						throwMessage(SyntaxError("unclosed string literal:" + SourceLog(line, column, 1, path, "")))
					}
					// hex := regexp.MustCompile(`\\x[0-9A-Fa-f]{2}`)
					// unicode1 := regexp.MustCompile(`\\u[0-9A-Fa-f]{4}`)
					// unicode2 := regexp.MustCompile(`\\u\\{[0-9A-Fa-f]+\\}`)
					// c_ := regexp.MustCompile(`\\[0-2][0-7]{0,2}`)
					// esc1 := regexp.MustCompile(`\\3[0-6][0-7]?`)
					// esc2 := regexp.MustCompile(`\\37[0-7]?`)
					// esc3 := regexp.MustCompile(`\\[4-7][0-7]?`)
					// esc := regexp.MustCompile(`\\.`)
					// dollar := regexp.MustCompile(`\\$`)
					match.src = _string
					match.typ = TokenType["String"]
				} else if match.typ == TokenType["BTick"] {
					quote := source[position]
					position++ // eat character
					template_string := ""
					for position < len(source) && source[position] != quote {
						char := source[position]
						position++ // eat character
						template_string += string(char)
						if char == '\\' {
							template_string += string(source[position])
							position++ // eat character
						}
					}
					if position >= len(source) || source[position] != quote {
						throwMessage(SyntaxError("unclosed template literal:" + SourceLog(line, column, 1, path, "")))
					}
					match.src = template_string
					match.typ = TokenType["TString"]
				} else if match.typ == "comment" {
					continue top
				} else if match.typ == TokenType["Identifier"] {
					keywords := []string{
						// -- statements --
						"var", // declarations ...
						"spawn",
						"immortal",
						"static",
						"using",
						"function",
						"class",
						"constructor",
						// "route",
						// "component",
						"if", // control flow ...
						"else",
						"break",
						"continue",
						"switch",
						"case",
						"default",
						"delete", // operators ...
						"do",     // loops ...
						"while",
						"for",
						"throw", // ...
						"return",
						"goto",
						"try", // error handling ...
						"catch",
						"finally",
						// -- modifiers --
						"private", // properties and methods ...
						"public",
						"default",
						"static",
						"extends", // classes
						"async",   // functions
						// -- modules --
						"import",
						"export",
						"from",
						"as",
						// -- variables --
						"globalThis",
						// "this",
						// -- expressions --
						"in",
						"of",
						"instanceof",
						"typeof",
						"void",
						"super",
						"new",
						"await",
						"go",
						"match",
					}
					for _, keyword := range keywords {
						if keyword == match.src {
							match.typ = keyword
							break
						}
					}
				} else if is_value(match.typ, TokenType["White Space"], TokenType["Comment"]) {
					position += match.length
					column += match.length
					if match.src == "\n" {
						column = 1
						line++
					}
					continue top
				}
				break
			}
		}
		if !matched {
			throwMessage(
				SyntaxError(
					// "unrecognised character found in source: " + string(source[position]) + SourceAtPosition(path, line, column)),
					"unrecognised character found in source: " + SourceLog(line, column, 1, path, "")),
			)
		}
		tokens.push(Token{
			src:  match.src,
			typ:  match.typ,
			line: line,
			col:  column,
			end:  column + match.length,
		})
		position += match.length
		column += match.length
	}
	tokens.push(Token{src: "EOF", typ: TokenType["EOF"], line: line, col: column, end: column})
	return tokens
}
