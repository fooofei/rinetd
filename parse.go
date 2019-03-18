package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// 20190317 还比较乱 初学者
// $ pigeon ./parse.peg | goimports | less > parse.go

type Unit struct {
	BindAddr     string
	BindPort     int
	BindProto    string
	ConnectAddr  string
	ConnectPort  int
	ConnectProto string
}
type MyPort struct {
	Port  int
	Proto string
}

func (u *Unit) String() string {
	return fmt.Sprintf("{bindAddr:%v bindPort:%v/%v "+
		"connectAddr:%v connectPort:%v/%v}",
		u.BindAddr, u.BindPort, u.BindProto,
		u.ConnectAddr, u.ConnectPort, u.ConnectProto)
}

func newUnit(bindAddr string, bindPort *MyPort, connectAddr string, connectPort *MyPort) *Unit {
	u := &Unit{}
	u.BindAddr = bindAddr
	u.BindPort = bindPort.Port
	u.BindProto = bindPort.Proto
	if u.BindProto == "" {
		u.BindProto = "/tcp"
	}
	u.ConnectAddr = connectAddr
	u.ConnectPort = connectPort.Port
	u.ConnectProto = connectPort.Proto
	if u.ConnectProto == "" {
		u.ConnectProto = "/tcp"
	}

	u.BindProto = strings.TrimPrefix(u.BindProto, "/")
	u.ConnectProto = strings.TrimPrefix(u.ConnectProto, "/")
	return u
}

var g = &grammar{
	rules: []*rule{
		{
			name: "File",
			pos:  position{line: 50, col: 1, offset: 1129},
			expr: &actionExpr{
				pos: position{line: 50, col: 8, offset: 1136},
				run: (*parser).callonFile1,
				expr: &seqExpr{
					pos: position{line: 50, col: 8, offset: 1136},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 50, col: 8, offset: 1136},
							label: "argl",
							expr: &zeroOrMoreExpr{
								pos: position{line: 50, col: 13, offset: 1141},
								expr: &choiceExpr{
									pos: position{line: 50, col: 14, offset: 1142},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 50, col: 14, offset: 1142},
											name: "Line",
										},
										&ruleRefExpr{
											pos:  position{line: 50, col: 21, offset: 1149},
											name: "Comment",
										},
										&ruleRefExpr{
											pos:  position{line: 50, col: 31, offset: 1159},
											name: "EndOfLine",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 50, col: 43, offset: 1171},
							name: "EndOfFile",
						},
					},
				},
			},
		},
		{
			name: "Line",
			pos:  position{line: 63, col: 1, offset: 1440},
			expr: &actionExpr{
				pos: position{line: 63, col: 8, offset: 1447},
				run: (*parser).callonLine1,
				expr: &seqExpr{
					pos: position{line: 63, col: 8, offset: 1447},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 63, col: 8, offset: 1447},
							label: "argu",
							expr: &ruleRefExpr{
								pos:  position{line: 63, col: 13, offset: 1452},
								name: "Unit",
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 63, col: 18, offset: 1457},
							expr: &charClassMatcher{
								pos:        position{line: 63, col: 18, offset: 1457},
								val:        "[ ]",
								chars:      []rune{' '},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 63, col: 23, offset: 1462},
							expr: &seqExpr{
								pos: position{line: 63, col: 24, offset: 1463},
								exprs: []interface{}{
									&choiceExpr{
										pos: position{line: 63, col: 25, offset: 1464},
										alternatives: []interface{}{
											&litMatcher{
												pos:        position{line: 63, col: 25, offset: 1464},
												val:        "//",
												ignoreCase: false,
											},
											&litMatcher{
												pos:        position{line: 63, col: 32, offset: 1471},
												val:        "#",
												ignoreCase: false,
											},
										},
									},
									&zeroOrMoreExpr{
										pos: position{line: 63, col: 37, offset: 1476},
										expr: &seqExpr{
											pos: position{line: 63, col: 38, offset: 1477},
											exprs: []interface{}{
												&notExpr{
													pos: position{line: 63, col: 38, offset: 1477},
													expr: &ruleRefExpr{
														pos:  position{line: 63, col: 39, offset: 1478},
														name: "EndOfLine",
													},
												},
												&anyMatcher{
													line: 63, col: 49, offset: 1488,
												},
											},
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 63, col: 55, offset: 1494},
							name: "EndOfLine",
						},
					},
				},
			},
		},
		{
			name: "Unit",
			pos:  position{line: 71, col: 1, offset: 1682},
			expr: &actionExpr{
				pos: position{line: 71, col: 8, offset: 1689},
				run: (*parser).callonUnit1,
				expr: &seqExpr{
					pos: position{line: 71, col: 8, offset: 1689},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 71, col: 8, offset: 1689},
							label: "bindAddr",
							expr: &ruleRefExpr{
								pos:  position{line: 71, col: 17, offset: 1698},
								name: "IPv4",
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 71, col: 22, offset: 1703},
							expr: &charClassMatcher{
								pos:        position{line: 71, col: 22, offset: 1703},
								val:        "[ ]",
								chars:      []rune{' '},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&labeledExpr{
							pos:   position{line: 71, col: 27, offset: 1708},
							label: "bindPort",
							expr: &ruleRefExpr{
								pos:  position{line: 71, col: 36, offset: 1717},
								name: "Port",
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 71, col: 41, offset: 1722},
							expr: &charClassMatcher{
								pos:        position{line: 71, col: 41, offset: 1722},
								val:        "[ ]",
								chars:      []rune{' '},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&labeledExpr{
							pos:   position{line: 71, col: 46, offset: 1727},
							label: "connectAddr",
							expr: &ruleRefExpr{
								pos:  position{line: 71, col: 58, offset: 1739},
								name: "IPv4",
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 71, col: 63, offset: 1744},
							expr: &charClassMatcher{
								pos:        position{line: 71, col: 63, offset: 1744},
								val:        "[ ]",
								chars:      []rune{' '},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&labeledExpr{
							pos:   position{line: 71, col: 68, offset: 1749},
							label: "connectPort",
							expr: &ruleRefExpr{
								pos:  position{line: 71, col: 80, offset: 1761},
								name: "Port",
							},
						},
					},
				},
			},
		},
		{
			name: "Port",
			pos:  position{line: 78, col: 1, offset: 1898},
			expr: &actionExpr{
				pos: position{line: 78, col: 8, offset: 1905},
				run: (*parser).callonPort1,
				expr: &seqExpr{
					pos: position{line: 78, col: 8, offset: 1905},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 78, col: 8, offset: 1905},
							label: "p",
							expr: &ruleRefExpr{
								pos:  position{line: 78, col: 10, offset: 1907},
								name: "DecimalDigit",
							},
						},
						&labeledExpr{
							pos:   position{line: 78, col: 23, offset: 1920},
							label: "o",
							expr: &choiceExpr{
								pos: position{line: 78, col: 26, offset: 1923},
								alternatives: []interface{}{
									&litMatcher{
										pos:        position{line: 78, col: 26, offset: 1923},
										val:        "/tcp",
										ignoreCase: false,
									},
									&litMatcher{
										pos:        position{line: 78, col: 35, offset: 1932},
										val:        "/udp",
										ignoreCase: false,
									},
									&litMatcher{
										pos:        position{line: 78, col: 44, offset: 1941},
										val:        "",
										ignoreCase: false,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "IPv4",
			pos:  position{line: 91, col: 1, offset: 2149},
			expr: &actionExpr{
				pos: position{line: 91, col: 8, offset: 2156},
				run: (*parser).callonIPv41,
				expr: &labeledExpr{
					pos:   position{line: 91, col: 8, offset: 2156},
					label: "v",
					expr: &choiceExpr{
						pos: position{line: 91, col: 11, offset: 2159},
						alternatives: []interface{}{
							&seqExpr{
								pos: position{line: 91, col: 13, offset: 2161},
								exprs: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 91, col: 13, offset: 2161},
										name: "DecimalDigit",
									},
									&litMatcher{
										pos:        position{line: 91, col: 26, offset: 2174},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 91, col: 30, offset: 2178},
										name: "DecimalDigit",
									},
									&litMatcher{
										pos:        position{line: 91, col: 43, offset: 2191},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 91, col: 47, offset: 2195},
										name: "DecimalDigit",
									},
									&litMatcher{
										pos:        position{line: 91, col: 60, offset: 2208},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 91, col: 64, offset: 2212},
										name: "DecimalDigit",
									},
								},
							},
							&litMatcher{
								pos:        position{line: 91, col: 81, offset: 2229},
								val:        "0",
								ignoreCase: false,
							},
						},
					},
				},
			},
		},
		{
			name: "DecimalDigit",
			pos:  position{line: 103, col: 1, offset: 2453},
			expr: &actionExpr{
				pos: position{line: 103, col: 16, offset: 2468},
				run: (*parser).callonDecimalDigit1,
				expr: &oneOrMoreExpr{
					pos: position{line: 103, col: 16, offset: 2468},
					expr: &charClassMatcher{
						pos:        position{line: 103, col: 16, offset: 2468},
						val:        "[0-9]",
						ranges:     []rune{'0', '9'},
						ignoreCase: false,
						inverted:   false,
					},
				},
			},
		},
		{
			name: "Comment",
			pos:  position{line: 109, col: 1, offset: 2512},
			expr: &seqExpr{
				pos: position{line: 109, col: 11, offset: 2522},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 109, col: 12, offset: 2523},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 109, col: 12, offset: 2523},
								val:        "//",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 109, col: 19, offset: 2530},
								val:        "#",
								ignoreCase: false,
							},
						},
					},
					&zeroOrMoreExpr{
						pos: position{line: 109, col: 24, offset: 2535},
						expr: &seqExpr{
							pos: position{line: 109, col: 25, offset: 2536},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 109, col: 25, offset: 2536},
									expr: &ruleRefExpr{
										pos:  position{line: 109, col: 26, offset: 2537},
										name: "EndOfLine",
									},
								},
								&anyMatcher{
									line: 109, col: 36, offset: 2547,
								},
							},
						},
					},
					&ruleRefExpr{
						pos:  position{line: 109, col: 40, offset: 2551},
						name: "EndOfLine",
					},
				},
			},
		},
		{
			name: "EndOfLine",
			pos:  position{line: 110, col: 1, offset: 2561},
			expr: &choiceExpr{
				pos: position{line: 110, col: 13, offset: 2573},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 110, col: 13, offset: 2573},
						val:        "\r\n",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 110, col: 22, offset: 2582},
						val:        "\n\r",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 110, col: 31, offset: 2591},
						val:        "\r",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 110, col: 38, offset: 2598},
						val:        "\n",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "EndOfFile",
			pos:  position{line: 112, col: 1, offset: 2618},
			expr: &notExpr{
				pos: position{line: 112, col: 13, offset: 2630},
				expr: &anyMatcher{
					line: 112, col: 14, offset: 2631,
				},
			},
		},
	},
}

func (c *current) onFile1(argl interface{}) (interface{}, error) {
	if ar0, ok := argl.([]interface{}); ok {
		r := make([]*Unit, 0)
		for _, ar := range ar0 {
			if u, ok := ar.(*Unit); ok {
				r = append(r, u)
			}
		}
		return r, nil
	}
	return nil, nil
}

func (p *parser) callonFile1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFile1(stack["argl"])
}

func (c *current) onLine1(argu interface{}) (interface{}, error) {
	// fmt.Printf("%s // '%s'\n", c.pos, string(c.text))
	if u, ok := argu.(*Unit); ok {
		return u, nil
	}
	return nil, fmt.Errorf("not valid Unit %v", argu)
}

func (p *parser) callonLine1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLine1(stack["argu"])
}

func (c *current) onUnit1(bindAddr, bindPort, connectAddr, connectPort interface{}) (interface{}, error) {
	u := newUnit(bindAddr.(string), bindPort.(*MyPort),
		connectAddr.(string), connectPort.(*MyPort))
	return u, nil
}

func (p *parser) callonUnit1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnit1(stack["bindAddr"], stack["bindPort"], stack["connectAddr"], stack["connectPort"])
}

func (c *current) onPort1(p, o interface{}) (interface{}, error) {
	r := &MyPort{}
	pi, _ := strconv.Atoi(p.(string))
	r.Port = pi
	if a, ok := o.([]byte); ok {
		r.Proto = string(a)
	} else {
		r.Proto = o.(string)
	}
	return r, nil
}

func (p *parser) callonPort1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPort1(stack["p"], stack["o"])
}

func (c *current) onIPv41(v interface{}) (interface{}, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	if a, ok := v.([]interface{}); ok && len(a) == 7 {
		// a == c.text
		s := string(c.text)
		return s, nil
	}
	return "null", nil
}

func (p *parser) callonIPv41() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIPv41(stack["v"])
}

func (c *current) onDecimalDigit1() (interface{}, error) {
	return string(c.text), nil
}

func (p *parser) callonDecimalDigit1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onDecimalDigit1()
}

var (
	// errNoRule is returned when the grammar to parse has no rule.
	errNoRule = errors.New("grammar has no rule")

	// errInvalidEncoding is returned when the source is not properly
	// utf8-encoded.
	errInvalidEncoding = errors.New("invalid encoding")

	// errNoMatch is returned if no match could be found.
	errNoMatch = errors.New("no match found")
)

// Option is a function that can set an option on the parser. It returns
// the previous setting as an Option.
type Option func(*parser) Option

// Debug creates an Option to set the debug flag to b. When set to true,
// debugging information is printed to stdout while parsing.
//
// The default is false.
func Debug(b bool) Option {
	return func(p *parser) Option {
		old := p.debug
		p.debug = b
		return Debug(old)
	}
}

// Memoize creates an Option to set the memoize flag to b. When set to true,
// the parser will cache all results so each expression is evaluated only
// once. This guarantees linear parsing time even for pathological cases,
// at the expense of more memory and slower times for typical cases.
//
// The default is false.
func Memoize(b bool) Option {
	return func(p *parser) Option {
		old := p.memoize
		p.memoize = b
		return Memoize(old)
	}
}

// Recover creates an Option to set the recover flag to b. When set to
// true, this causes the parser to recover from panics and convert it
// to an error. Setting it to false can be useful while debugging to
// access the full stack trace.
//
// The default is true.
func Recover(b bool) Option {
	return func(p *parser) Option {
		old := p.recover
		p.recover = b
		return Recover(old)
	}
}

// ParseFile parses the file identified by filename.
func ParseFile(filename string, opts ...Option) (interface{}, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseReader(filename, f, opts...)
}

// ParseReader parses the data from r using filename as information in the
// error messages.
func ParseReader(filename string, r io.Reader, opts ...Option) (interface{}, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Parse(filename, b, opts...)
}

// Parse parses the data from b using filename as information in the
// error messages.
func Parse(filename string, b []byte, opts ...Option) (interface{}, error) {
	return newParser(filename, b, opts...).parse(g)
}

// position records a position in the text.
type position struct {
	line, col, offset int
}

func (p position) String() string {
	return fmt.Sprintf("%d:%d [%d]", p.line, p.col, p.offset)
}

// savepoint stores all state required to go back to this point in the
// parser.
type savepoint struct {
	position
	rn rune
	w  int
}

type current struct {
	pos  position // start position of the match
	text []byte   // raw text of the match
}

// the AST types...

type grammar struct {
	pos   position
	rules []*rule
}

type rule struct {
	pos         position
	name        string
	displayName string
	expr        interface{}
}

type choiceExpr struct {
	pos          position
	alternatives []interface{}
}

type actionExpr struct {
	pos  position
	expr interface{}
	run  func(*parser) (interface{}, error)
}

type seqExpr struct {
	pos   position
	exprs []interface{}
}

type labeledExpr struct {
	pos   position
	label string
	expr  interface{}
}

type expr struct {
	pos  position
	expr interface{}
}

type andExpr expr
type notExpr expr
type zeroOrOneExpr expr
type zeroOrMoreExpr expr
type oneOrMoreExpr expr

type ruleRefExpr struct {
	pos  position
	name string
}

type andCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type notCodeExpr struct {
	pos position
	run func(*parser) (bool, error)
}

type litMatcher struct {
	pos        position
	val        string
	ignoreCase bool
}

type charClassMatcher struct {
	pos        position
	val        string
	chars      []rune
	ranges     []rune
	classes    []*unicode.RangeTable
	ignoreCase bool
	inverted   bool
}

type anyMatcher position

// errList cumulates the errors found by the parser.
type errList []error

func (e *errList) add(err error) {
	*e = append(*e, err)
}

func (e errList) err() error {
	if len(e) == 0 {
		return nil
	}
	e.dedupe()
	return e
}

func (e *errList) dedupe() {
	var cleaned []error
	set := make(map[string]bool)
	for _, err := range *e {
		if msg := err.Error(); !set[msg] {
			set[msg] = true
			cleaned = append(cleaned, err)
		}
	}
	*e = cleaned
}

func (e errList) Error() string {
	switch len(e) {
	case 0:
		return ""
	case 1:
		return e[0].Error()
	default:
		var buf bytes.Buffer

		for i, err := range e {
			if i > 0 {
				buf.WriteRune('\n')
			}
			buf.WriteString(err.Error())
		}
		return buf.String()
	}
}

// parserError wraps an error with a prefix indicating the rule in which
// the error occurred. The original error is stored in the Inner field.
type parserError struct {
	Inner  error
	pos    position
	prefix string
}

// Error returns the error message.
func (p *parserError) Error() string {
	return p.prefix + ": " + p.Inner.Error()
}

// newParser creates a parser with the specified input source and options.
func newParser(filename string, b []byte, opts ...Option) *parser {
	p := &parser{
		filename: filename,
		errs:     new(errList),
		data:     b,
		pt:       savepoint{position: position{line: 1}},
		recover:  true,
	}
	p.setOptions(opts)
	return p
}

// setOptions applies the options to the parser.
func (p *parser) setOptions(opts []Option) {
	for _, opt := range opts {
		opt(p)
	}
}

type resultTuple struct {
	v   interface{}
	b   bool
	end savepoint
}

type parser struct {
	filename string
	pt       savepoint
	cur      current

	data []byte
	errs *errList

	recover bool
	debug   bool
	depth   int

	memoize bool
	// memoization table for the packrat algorithm:
	// map[offset in source] map[expression or rule] {value, match}
	memo map[int]map[interface{}]resultTuple

	// rules table, maps the rule identifier to the rule node
	rules map[string]*rule
	// variables stack, map of label to value
	vstack []map[string]interface{}
	// rule stack, allows identification of the current rule in errors
	rstack []*rule

	// stats
	exprCnt int
}

// push a variable set on the vstack.
func (p *parser) pushV() {
	if cap(p.vstack) == len(p.vstack) {
		// create new empty slot in the stack
		p.vstack = append(p.vstack, nil)
	} else {
		// slice to 1 more
		p.vstack = p.vstack[:len(p.vstack)+1]
	}

	// get the last args set
	m := p.vstack[len(p.vstack)-1]
	if m != nil && len(m) == 0 {
		// empty map, all good
		return
	}

	m = make(map[string]interface{})
	p.vstack[len(p.vstack)-1] = m
}

// pop a variable set from the vstack.
func (p *parser) popV() {
	// if the map is not empty, clear it
	m := p.vstack[len(p.vstack)-1]
	if len(m) > 0 {
		// GC that map
		p.vstack[len(p.vstack)-1] = nil
	}
	p.vstack = p.vstack[:len(p.vstack)-1]
}

func (p *parser) print(prefix, s string) string {
	if !p.debug {
		return s
	}

	fmt.Printf("%s %d:%d:%d: %s [%#U]\n",
		prefix, p.pt.line, p.pt.col, p.pt.offset, s, p.pt.rn)
	return s
}

func (p *parser) in(s string) string {
	p.depth++
	return p.print(strings.Repeat(" ", p.depth)+">", s)
}

func (p *parser) out(s string) string {
	p.depth--
	return p.print(strings.Repeat(" ", p.depth)+"<", s)
}

func (p *parser) addErr(err error) {
	p.addErrAt(err, p.pt.position)
}

func (p *parser) addErrAt(err error, pos position) {
	var buf bytes.Buffer
	if p.filename != "" {
		buf.WriteString(p.filename)
	}
	if buf.Len() > 0 {
		buf.WriteString(":")
	}
	buf.WriteString(fmt.Sprintf("%d:%d (%d)", pos.line, pos.col, pos.offset))
	if len(p.rstack) > 0 {
		if buf.Len() > 0 {
			buf.WriteString(": ")
		}
		rule := p.rstack[len(p.rstack)-1]
		if rule.displayName != "" {
			buf.WriteString("rule " + rule.displayName)
		} else {
			buf.WriteString("rule " + rule.name)
		}
	}
	pe := &parserError{Inner: err, pos: pos, prefix: buf.String()}
	p.errs.add(pe)
}

// read advances the parser to the next rune.
func (p *parser) read() {
	p.pt.offset += p.pt.w
	rn, n := utf8.DecodeRune(p.data[p.pt.offset:])
	p.pt.rn = rn
	p.pt.w = n
	p.pt.col++
	if rn == '\n' {
		p.pt.line++
		p.pt.col = 0
	}

	if rn == utf8.RuneError {
		if n == 1 {
			p.addErr(errInvalidEncoding)
		}
	}
}

// restore parser position to the savepoint pt.
func (p *parser) restore(pt savepoint) {
	if p.debug {
		defer p.out(p.in("restore"))
	}
	if pt.offset == p.pt.offset {
		return
	}
	p.pt = pt
}

// get the slice of bytes from the savepoint start to the current position.
func (p *parser) sliceFrom(start savepoint) []byte {
	return p.data[start.position.offset:p.pt.position.offset]
}

func (p *parser) getMemoized(node interface{}) (resultTuple, bool) {
	if len(p.memo) == 0 {
		return resultTuple{}, false
	}
	m := p.memo[p.pt.offset]
	if len(m) == 0 {
		return resultTuple{}, false
	}
	res, ok := m[node]
	return res, ok
}

func (p *parser) setMemoized(pt savepoint, node interface{}, tuple resultTuple) {
	if p.memo == nil {
		p.memo = make(map[int]map[interface{}]resultTuple)
	}
	m := p.memo[pt.offset]
	if m == nil {
		m = make(map[interface{}]resultTuple)
		p.memo[pt.offset] = m
	}
	m[node] = tuple
}

func (p *parser) buildRulesTable(g *grammar) {
	p.rules = make(map[string]*rule, len(g.rules))
	for _, r := range g.rules {
		p.rules[r.name] = r
	}
}

func (p *parser) parse(g *grammar) (val interface{}, err error) {
	if len(g.rules) == 0 {
		p.addErr(errNoRule)
		return nil, p.errs.err()
	}

	// TODO : not super critical but this could be generated
	p.buildRulesTable(g)

	if p.recover {
		// panic can be used in action code to stop parsing immediately
		// and return the panic as an error.
		defer func() {
			if e := recover(); e != nil {
				if p.debug {
					defer p.out(p.in("panic handler"))
				}
				val = nil
				switch e := e.(type) {
				case error:
					p.addErr(e)
				default:
					p.addErr(fmt.Errorf("%v", e))
				}
				err = p.errs.err()
			}
		}()
	}

	// start rule is rule [0]
	p.read() // advance to first rune
	val, ok := p.parseRule(g.rules[0])
	if !ok {
		if len(*p.errs) == 0 {
			// make sure this doesn't go out silently
			p.addErr(errNoMatch)
		}
		return nil, p.errs.err()
	}
	return val, p.errs.err()
}

func (p *parser) parseRule(rule *rule) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRule " + rule.name))
	}

	if p.memoize {
		res, ok := p.getMemoized(rule)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
	}

	start := p.pt
	p.rstack = append(p.rstack, rule)
	p.pushV()
	val, ok := p.parseExpr(rule.expr)
	p.popV()
	p.rstack = p.rstack[:len(p.rstack)-1]
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}

	if p.memoize {
		p.setMemoized(start, rule, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseExpr(expr interface{}) (interface{}, bool) {
	var pt savepoint
	var ok bool

	if p.memoize {
		res, ok := p.getMemoized(expr)
		if ok {
			p.restore(res.end)
			return res.v, res.b
		}
		pt = p.pt
	}

	p.exprCnt++
	var val interface{}
	switch expr := expr.(type) {
	case *actionExpr:
		val, ok = p.parseActionExpr(expr)
	case *andCodeExpr:
		val, ok = p.parseAndCodeExpr(expr)
	case *andExpr:
		val, ok = p.parseAndExpr(expr)
	case *anyMatcher:
		val, ok = p.parseAnyMatcher(expr)
	case *charClassMatcher:
		val, ok = p.parseCharClassMatcher(expr)
	case *choiceExpr:
		val, ok = p.parseChoiceExpr(expr)
	case *labeledExpr:
		val, ok = p.parseLabeledExpr(expr)
	case *litMatcher:
		val, ok = p.parseLitMatcher(expr)
	case *notCodeExpr:
		val, ok = p.parseNotCodeExpr(expr)
	case *notExpr:
		val, ok = p.parseNotExpr(expr)
	case *oneOrMoreExpr:
		val, ok = p.parseOneOrMoreExpr(expr)
	case *ruleRefExpr:
		val, ok = p.parseRuleRefExpr(expr)
	case *seqExpr:
		val, ok = p.parseSeqExpr(expr)
	case *zeroOrMoreExpr:
		val, ok = p.parseZeroOrMoreExpr(expr)
	case *zeroOrOneExpr:
		val, ok = p.parseZeroOrOneExpr(expr)
	default:
		panic(fmt.Sprintf("unknown expression type %T", expr))
	}
	if p.memoize {
		p.setMemoized(pt, expr, resultTuple{val, ok, p.pt})
	}
	return val, ok
}

func (p *parser) parseActionExpr(act *actionExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseActionExpr"))
	}

	start := p.pt
	val, ok := p.parseExpr(act.expr)
	if ok {
		p.cur.pos = start.position
		p.cur.text = p.sliceFrom(start)
		actVal, err := act.run(p)
		if err != nil {
			p.addErrAt(err, start.position)
		}
		val = actVal
	}
	if ok && p.debug {
		p.print(strings.Repeat(" ", p.depth)+"MATCH", string(p.sliceFrom(start)))
	}
	return val, ok
}

func (p *parser) parseAndCodeExpr(and *andCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndCodeExpr"))
	}

	ok, err := and.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, ok
}

func (p *parser) parseAndExpr(and *andExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAndExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(and.expr)
	p.popV()
	p.restore(pt)
	return nil, ok
}

func (p *parser) parseAnyMatcher(any *anyMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseAnyMatcher"))
	}

	if p.pt.rn != utf8.RuneError {
		start := p.pt
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseCharClassMatcher(chr *charClassMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseCharClassMatcher"))
	}

	cur := p.pt.rn
	// can't match EOF
	if cur == utf8.RuneError {
		return nil, false
	}
	start := p.pt
	if chr.ignoreCase {
		cur = unicode.ToLower(cur)
	}

	// try to match in the list of available chars
	for _, rn := range chr.chars {
		if rn == cur {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of ranges
	for i := 0; i < len(chr.ranges); i += 2 {
		if cur >= chr.ranges[i] && cur <= chr.ranges[i+1] {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	// try to match in the list of Unicode classes
	for _, cl := range chr.classes {
		if unicode.Is(cl, cur) {
			if chr.inverted {
				return nil, false
			}
			p.read()
			return p.sliceFrom(start), true
		}
	}

	if chr.inverted {
		p.read()
		return p.sliceFrom(start), true
	}
	return nil, false
}

func (p *parser) parseChoiceExpr(ch *choiceExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseChoiceExpr"))
	}

	for _, alt := range ch.alternatives {
		p.pushV()
		val, ok := p.parseExpr(alt)
		p.popV()
		if ok {
			return val, ok
		}
	}
	return nil, false
}

func (p *parser) parseLabeledExpr(lab *labeledExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLabeledExpr"))
	}

	p.pushV()
	val, ok := p.parseExpr(lab.expr)
	p.popV()
	if ok && lab.label != "" {
		m := p.vstack[len(p.vstack)-1]
		m[lab.label] = val
	}
	return val, ok
}

func (p *parser) parseLitMatcher(lit *litMatcher) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseLitMatcher"))
	}

	start := p.pt
	for _, want := range lit.val {
		cur := p.pt.rn
		if lit.ignoreCase {
			cur = unicode.ToLower(cur)
		}
		if cur != want {
			p.restore(start)
			return nil, false
		}
		p.read()
	}
	return p.sliceFrom(start), true
}

func (p *parser) parseNotCodeExpr(not *notCodeExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotCodeExpr"))
	}

	ok, err := not.run(p)
	if err != nil {
		p.addErr(err)
	}
	return nil, !ok
}

func (p *parser) parseNotExpr(not *notExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseNotExpr"))
	}

	pt := p.pt
	p.pushV()
	_, ok := p.parseExpr(not.expr)
	p.popV()
	p.restore(pt)
	return nil, !ok
}

func (p *parser) parseOneOrMoreExpr(expr *oneOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseOneOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			if len(vals) == 0 {
				// did not match once, no match
				return nil, false
			}
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseRuleRefExpr(ref *ruleRefExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseRuleRefExpr " + ref.name))
	}

	if ref.name == "" {
		panic(fmt.Sprintf("%s: invalid rule: missing name", ref.pos))
	}

	rule := p.rules[ref.name]
	if rule == nil {
		p.addErr(fmt.Errorf("undefined rule: %s", ref.name))
		return nil, false
	}
	return p.parseRule(rule)
}

func (p *parser) parseSeqExpr(seq *seqExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseSeqExpr"))
	}

	var vals []interface{}

	pt := p.pt
	for _, expr := range seq.exprs {
		val, ok := p.parseExpr(expr)
		if !ok {
			p.restore(pt)
			return nil, false
		}
		vals = append(vals, val)
	}
	return vals, true
}

func (p *parser) parseZeroOrMoreExpr(expr *zeroOrMoreExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrMoreExpr"))
	}

	var vals []interface{}

	for {
		p.pushV()
		val, ok := p.parseExpr(expr.expr)
		p.popV()
		if !ok {
			return vals, true
		}
		vals = append(vals, val)
	}
}

func (p *parser) parseZeroOrOneExpr(expr *zeroOrOneExpr) (interface{}, bool) {
	if p.debug {
		defer p.out(p.in("parseZeroOrOneExpr"))
	}

	p.pushV()
	val, _ := p.parseExpr(expr.expr)
	p.popV()
	// whether it matched or not, consider it a match
	return val, true
}

func rangeTable(class string) *unicode.RangeTable {
	if rt, ok := unicode.Categories[class]; ok {
		return rt
	}
	if rt, ok := unicode.Properties[class]; ok {
		return rt
	}
	if rt, ok := unicode.Scripts[class]; ok {
		return rt
	}

	// cannot happen
	panic(fmt.Sprintf("invalid Unicode class: %s", class))
}
