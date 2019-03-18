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
// pigeon ./parse.peg | goimports | less > parse.go

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

func onFile(arg0 interface{}, arg1 interface{}) (interface{}, error) {
	r := make([]*Unit, 0)
	if ar0, ok := arg0.([]interface{}); ok {
		for _, ar := range ar0 {
			if ar1, ok := ar.([]interface{}); ok && len(ar1) > 0 {
				if u, ok := ar1[0].(*Unit); ok {
					r = append(r, u)
				}
			}
		}
	}
	if u, ok := arg1.(*Unit); ok {
		r = append(r, u)
	}
	return r, nil
}

func onLine(arg0 interface{}) (interface{}, error) {
	// fmt.Printf("%s // '%s'\n", c.pos, string(c.text))
	if u, ok := arg0.(*Unit); ok {
		return u, nil
	}
	return nil, fmt.Errorf("not valid Unit %v", arg0)
}

func onPort(arg0 interface{}, arg1 interface{}) (interface{}, error) {
	r := &MyPort{}
	pi, _ := strconv.Atoi(arg0.(string))
	r.Port = pi
	if a, ok := arg1.([]byte); ok {
		r.Proto = string(a)
	} else {
		r.Proto = arg1.(string)
	}
	return r, nil
}

func onIPv4(arg0 interface{}, txt []byte) (interface{}, error) {

	if s, ok := arg0.(string); ok {
		return s, nil
	}
	if a, ok := arg0.([]interface{}); ok && len(a) == 7 {
		// a == c.text
		s := string(txt)
		return s, nil
	}
	return "null", nil
}

var g = &grammar{
	rules: []*rule{
		{
			name: "File",
			pos:  position{line: 106, col: 1, offset: 2446},
			expr: &actionExpr{
				pos: position{line: 106, col: 19, offset: 2464},
				run: (*parser).callonFile1,
				expr: &seqExpr{
					pos: position{line: 106, col: 19, offset: 2464},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 106, col: 19, offset: 2464},
							label: "arg0",
							expr: &zeroOrMoreExpr{
								pos: position{line: 106, col: 24, offset: 2469},
								expr: &choiceExpr{
									pos: position{line: 106, col: 25, offset: 2470},
									alternatives: []interface{}{
										&seqExpr{
											pos: position{line: 106, col: 25, offset: 2470},
											exprs: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 106, col: 25, offset: 2470},
													name: "Line",
												},
												&ruleRefExpr{
													pos:  position{line: 106, col: 30, offset: 2475},
													name: "EndOfLine",
												},
											},
										},
										&seqExpr{
											pos: position{line: 106, col: 42, offset: 2487},
											exprs: []interface{}{
												&ruleRefExpr{
													pos:  position{line: 106, col: 42, offset: 2487},
													name: "Comment",
												},
												&ruleRefExpr{
													pos:  position{line: 106, col: 50, offset: 2495},
													name: "EndOfLine",
												},
											},
										},
										&ruleRefExpr{
											pos:  position{line: 106, col: 62, offset: 2507},
											name: "EndOfLine",
										},
									},
								},
							},
						},
						&labeledExpr{
							pos:   position{line: 106, col: 74, offset: 2519},
							label: "arg1",
							expr: &zeroOrOneExpr{
								pos: position{line: 106, col: 79, offset: 2524},
								expr: &choiceExpr{
									pos: position{line: 106, col: 80, offset: 2525},
									alternatives: []interface{}{
										&ruleRefExpr{
											pos:  position{line: 106, col: 80, offset: 2525},
											name: "Line",
										},
										&ruleRefExpr{
											pos:  position{line: 106, col: 87, offset: 2532},
											name: "Comment",
										},
									},
								},
							},
						},
						&ruleRefExpr{
							pos:  position{line: 106, col: 98, offset: 2543},
							name: "EndOfFile",
						},
					},
				},
			},
		},
		{
			name: "Line",
			pos:  position{line: 108, col: 1, offset: 2602},
			expr: &actionExpr{
				pos: position{line: 108, col: 19, offset: 2620},
				run: (*parser).callonLine1,
				expr: &seqExpr{
					pos: position{line: 108, col: 19, offset: 2620},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 108, col: 19, offset: 2620},
							label: "arg0",
							expr: &ruleRefExpr{
								pos:  position{line: 108, col: 24, offset: 2625},
								name: "Unit",
							},
						},
						&zeroOrMoreExpr{
							pos: position{line: 108, col: 29, offset: 2630},
							expr: &charClassMatcher{
								pos:        position{line: 108, col: 29, offset: 2630},
								val:        "[ ]",
								chars:      []rune{' '},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&zeroOrOneExpr{
							pos: position{line: 108, col: 34, offset: 2635},
							expr: &ruleRefExpr{
								pos:  position{line: 108, col: 34, offset: 2635},
								name: "Comment",
							},
						},
					},
				},
			},
		},
		{
			name: "Unit",
			pos:  position{line: 109, col: 1, offset: 2667},
			expr: &actionExpr{
				pos: position{line: 109, col: 19, offset: 2685},
				run: (*parser).callonUnit1,
				expr: &seqExpr{
					pos: position{line: 109, col: 19, offset: 2685},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 109, col: 19, offset: 2685},
							label: "bindAddr",
							expr: &ruleRefExpr{
								pos:  position{line: 109, col: 28, offset: 2694},
								name: "IPv4",
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 109, col: 33, offset: 2699},
							expr: &charClassMatcher{
								pos:        position{line: 109, col: 33, offset: 2699},
								val:        "[ ]",
								chars:      []rune{' '},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&labeledExpr{
							pos:   position{line: 109, col: 38, offset: 2704},
							label: "bindPort",
							expr: &ruleRefExpr{
								pos:  position{line: 109, col: 47, offset: 2713},
								name: "Port",
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 109, col: 52, offset: 2718},
							expr: &charClassMatcher{
								pos:        position{line: 109, col: 52, offset: 2718},
								val:        "[ ]",
								chars:      []rune{' '},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&labeledExpr{
							pos:   position{line: 109, col: 57, offset: 2723},
							label: "connectAddr",
							expr: &ruleRefExpr{
								pos:  position{line: 109, col: 69, offset: 2735},
								name: "IPv4",
							},
						},
						&oneOrMoreExpr{
							pos: position{line: 109, col: 74, offset: 2740},
							expr: &charClassMatcher{
								pos:        position{line: 109, col: 74, offset: 2740},
								val:        "[ ]",
								chars:      []rune{' '},
								ignoreCase: false,
								inverted:   false,
							},
						},
						&labeledExpr{
							pos:   position{line: 109, col: 79, offset: 2745},
							label: "connectPort",
							expr: &ruleRefExpr{
								pos:  position{line: 109, col: 91, offset: 2757},
								name: "Port",
							},
						},
					},
				},
			},
		},
		{
			name: "Port",
			pos:  position{line: 113, col: 1, offset: 2913},
			expr: &actionExpr{
				pos: position{line: 113, col: 19, offset: 2931},
				run: (*parser).callonPort1,
				expr: &seqExpr{
					pos: position{line: 113, col: 19, offset: 2931},
					exprs: []interface{}{
						&labeledExpr{
							pos:   position{line: 113, col: 19, offset: 2931},
							label: "arg0",
							expr: &ruleRefExpr{
								pos:  position{line: 113, col: 24, offset: 2936},
								name: "DecimalDigit",
							},
						},
						&labeledExpr{
							pos:   position{line: 113, col: 37, offset: 2949},
							label: "arg1",
							expr: &choiceExpr{
								pos: position{line: 113, col: 43, offset: 2955},
								alternatives: []interface{}{
									&litMatcher{
										pos:        position{line: 113, col: 43, offset: 2955},
										val:        "/tcp",
										ignoreCase: false,
									},
									&litMatcher{
										pos:        position{line: 113, col: 52, offset: 2964},
										val:        "/udp",
										ignoreCase: false,
									},
									&litMatcher{
										pos:        position{line: 113, col: 61, offset: 2973},
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
			pos:  position{line: 114, col: 1, offset: 3007},
			expr: &actionExpr{
				pos: position{line: 114, col: 19, offset: 3025},
				run: (*parser).callonIPv41,
				expr: &labeledExpr{
					pos:   position{line: 114, col: 19, offset: 3025},
					label: "arg0",
					expr: &choiceExpr{
						pos: position{line: 114, col: 25, offset: 3031},
						alternatives: []interface{}{
							&seqExpr{
								pos: position{line: 114, col: 27, offset: 3033},
								exprs: []interface{}{
									&ruleRefExpr{
										pos:  position{line: 114, col: 27, offset: 3033},
										name: "DecimalDigit",
									},
									&litMatcher{
										pos:        position{line: 114, col: 40, offset: 3046},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 114, col: 44, offset: 3050},
										name: "DecimalDigit",
									},
									&litMatcher{
										pos:        position{line: 114, col: 57, offset: 3063},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 114, col: 61, offset: 3067},
										name: "DecimalDigit",
									},
									&litMatcher{
										pos:        position{line: 114, col: 74, offset: 3080},
										val:        ".",
										ignoreCase: false,
									},
									&ruleRefExpr{
										pos:  position{line: 114, col: 78, offset: 3084},
										name: "DecimalDigit",
									},
								},
							},
							&litMatcher{
								pos:        position{line: 114, col: 95, offset: 3101},
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
			pos:  position{line: 115, col: 1, offset: 3136},
			expr: &actionExpr{
				pos: position{line: 115, col: 19, offset: 3154},
				run: (*parser).callonDecimalDigit1,
				expr: &oneOrMoreExpr{
					pos: position{line: 115, col: 19, offset: 3154},
					expr: &charClassMatcher{
						pos:        position{line: 115, col: 19, offset: 3154},
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
			pos:  position{line: 116, col: 1, offset: 3190},
			expr: &seqExpr{
				pos: position{line: 116, col: 19, offset: 3208},
				exprs: []interface{}{
					&choiceExpr{
						pos: position{line: 116, col: 20, offset: 3209},
						alternatives: []interface{}{
							&litMatcher{
								pos:        position{line: 116, col: 20, offset: 3209},
								val:        "//",
								ignoreCase: false,
							},
							&litMatcher{
								pos:        position{line: 116, col: 27, offset: 3216},
								val:        "#",
								ignoreCase: false,
							},
						},
					},
					&zeroOrMoreExpr{
						pos: position{line: 116, col: 32, offset: 3221},
						expr: &seqExpr{
							pos: position{line: 116, col: 33, offset: 3222},
							exprs: []interface{}{
								&notExpr{
									pos: position{line: 116, col: 33, offset: 3222},
									expr: &ruleRefExpr{
										pos:  position{line: 116, col: 34, offset: 3223},
										name: "EndOfLine",
									},
								},
								&anyMatcher{
									line: 116, col: 44, offset: 3233,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "EndOfLine",
			pos:  position{line: 117, col: 1, offset: 3237},
			expr: &choiceExpr{
				pos: position{line: 117, col: 19, offset: 3255},
				alternatives: []interface{}{
					&litMatcher{
						pos:        position{line: 117, col: 19, offset: 3255},
						val:        "\r\n",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 117, col: 28, offset: 3264},
						val:        "\n\r",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 117, col: 37, offset: 3273},
						val:        "\r",
						ignoreCase: false,
					},
					&litMatcher{
						pos:        position{line: 117, col: 44, offset: 3280},
						val:        "\n",
						ignoreCase: false,
					},
				},
			},
		},
		{
			name: "EndOfFile",
			pos:  position{line: 118, col: 1, offset: 3285},
			expr: &notExpr{
				pos: position{line: 118, col: 19, offset: 3303},
				expr: &anyMatcher{
					line: 118, col: 20, offset: 3304,
				},
			},
		},
	},
}

func (c *current) onFile1(arg0, arg1 interface{}) (interface{}, error) {
	return onFile(arg0, arg1)
}

func (p *parser) callonFile1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onFile1(stack["arg0"], stack["arg1"])
}

func (c *current) onLine1(arg0 interface{}) (interface{}, error) {
	return onLine(arg0)
}

func (p *parser) callonLine1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onLine1(stack["arg0"])
}

func (c *current) onUnit1(bindAddr, bindPort, connectAddr, connectPort interface{}) (interface{}, error) {
	return newUnit(bindAddr.(string), bindPort.(*MyPort),
		connectAddr.(string), connectPort.(*MyPort)), nil
}

func (p *parser) callonUnit1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onUnit1(stack["bindAddr"], stack["bindPort"], stack["connectAddr"], stack["connectPort"])
}

func (c *current) onPort1(arg0, arg1 interface{}) (interface{}, error) {
	return onPort(arg0, arg1)
}

func (p *parser) callonPort1() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onPort1(stack["arg0"], stack["arg1"])
}

func (c *current) onIPv41(arg0 interface{}) (interface{}, error) {
	return onIPv4(arg0, c.text)
}

func (p *parser) callonIPv41() (interface{}, error) {
	stack := p.vstack[len(p.vstack)-1]
	_ = stack
	return p.cur.onIPv41(stack["arg0"])
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
