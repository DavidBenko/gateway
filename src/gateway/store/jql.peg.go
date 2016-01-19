package store

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	rulee
	ruleorder
	rulecast
	ruleasc
	ruledesc
	rulelimit
	ruleoffset
	rulee1
	rulee2
	rulee3
	ruleexpression
	ruleop
	rulepath
	ruleword
	rulevalue1
	rulevalue2
	ruleplaceholder
	rulestring
	rulenumber
	ruleboolean
	rulenull
	rulewhole
	ruleand
	ruleor
	ruleopen
	ruleclose
	rulesp

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"e",
	"order",
	"cast",
	"asc",
	"desc",
	"limit",
	"offset",
	"e1",
	"e2",
	"e3",
	"expression",
	"op",
	"path",
	"word",
	"value1",
	"value2",
	"placeholder",
	"string",
	"number",
	"boolean",
	"null",
	"whole",
	"and",
	"or",
	"open",
	"close",
	"sp",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next uint32, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/*func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2 * len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}*/

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type JQL struct {
	Buffer string
	buffer []rune
	rules  [28]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range []rune(buffer) {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *JQL
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *JQL) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *JQL) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *JQL) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens32{tree: make([]token32, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 e <- <(sp e1 (order / limit / offset)* !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[rulesp]() {
					goto l0
				}
				if !_rules[rulee1]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					{
						position4, tokenIndex4, depth4 := position, tokenIndex, depth
						if !_rules[ruleorder]() {
							goto l5
						}
						goto l4
					l5:
						position, tokenIndex, depth = position4, tokenIndex4, depth4
						if !_rules[rulelimit]() {
							goto l6
						}
						goto l4
					l6:
						position, tokenIndex, depth = position4, tokenIndex4, depth4
						if !_rules[ruleoffset]() {
							goto l3
						}
					}
				l4:
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
				}
				{
					position7, tokenIndex7, depth7 := position, tokenIndex, depth
					if !matchDot() {
						goto l7
					}
					goto l0
				l7:
					position, tokenIndex, depth = position7, tokenIndex7, depth7
				}
				depth--
				add(rulee, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 order <- <(('o' / 'O') ('r' / 'R') ('d' / 'D') ('e' / 'E') ('r' / 'R') sp (cast / path) (asc / desc) sp)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				{
					position10, tokenIndex10, depth10 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l11
					}
					position++
					goto l10
				l11:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if buffer[position] != rune('O') {
						goto l8
					}
					position++
				}
			l10:
				{
					position12, tokenIndex12, depth12 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l13
					}
					position++
					goto l12
				l13:
					position, tokenIndex, depth = position12, tokenIndex12, depth12
					if buffer[position] != rune('R') {
						goto l8
					}
					position++
				}
			l12:
				{
					position14, tokenIndex14, depth14 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l15
					}
					position++
					goto l14
				l15:
					position, tokenIndex, depth = position14, tokenIndex14, depth14
					if buffer[position] != rune('D') {
						goto l8
					}
					position++
				}
			l14:
				{
					position16, tokenIndex16, depth16 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l17
					}
					position++
					goto l16
				l17:
					position, tokenIndex, depth = position16, tokenIndex16, depth16
					if buffer[position] != rune('E') {
						goto l8
					}
					position++
				}
			l16:
				{
					position18, tokenIndex18, depth18 := position, tokenIndex, depth
					if buffer[position] != rune('r') {
						goto l19
					}
					position++
					goto l18
				l19:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
					if buffer[position] != rune('R') {
						goto l8
					}
					position++
				}
			l18:
				if !_rules[rulesp]() {
					goto l8
				}
				{
					position20, tokenIndex20, depth20 := position, tokenIndex, depth
					if !_rules[rulecast]() {
						goto l21
					}
					goto l20
				l21:
					position, tokenIndex, depth = position20, tokenIndex20, depth20
					if !_rules[rulepath]() {
						goto l8
					}
				}
			l20:
				{
					position22, tokenIndex22, depth22 := position, tokenIndex, depth
					if !_rules[ruleasc]() {
						goto l23
					}
					goto l22
				l23:
					position, tokenIndex, depth = position22, tokenIndex22, depth22
					if !_rules[ruledesc]() {
						goto l8
					}
				}
			l22:
				if !_rules[rulesp]() {
					goto l8
				}
				depth--
				add(ruleorder, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 2 cast <- <('n' 'u' 'm' 'e' 'r' 'i' 'c' sp open path close)> */
		func() bool {
			position24, tokenIndex24, depth24 := position, tokenIndex, depth
			{
				position25 := position
				depth++
				if buffer[position] != rune('n') {
					goto l24
				}
				position++
				if buffer[position] != rune('u') {
					goto l24
				}
				position++
				if buffer[position] != rune('m') {
					goto l24
				}
				position++
				if buffer[position] != rune('e') {
					goto l24
				}
				position++
				if buffer[position] != rune('r') {
					goto l24
				}
				position++
				if buffer[position] != rune('i') {
					goto l24
				}
				position++
				if buffer[position] != rune('c') {
					goto l24
				}
				position++
				if !_rules[rulesp]() {
					goto l24
				}
				if !_rules[ruleopen]() {
					goto l24
				}
				if !_rules[rulepath]() {
					goto l24
				}
				if !_rules[ruleclose]() {
					goto l24
				}
				depth--
				add(rulecast, position25)
			}
			return true
		l24:
			position, tokenIndex, depth = position24, tokenIndex24, depth24
			return false
		},
		/* 3 asc <- <(('a' / 'A') ('s' / 'S') ('c' / 'C'))> */
		func() bool {
			position26, tokenIndex26, depth26 := position, tokenIndex, depth
			{
				position27 := position
				depth++
				{
					position28, tokenIndex28, depth28 := position, tokenIndex, depth
					if buffer[position] != rune('a') {
						goto l29
					}
					position++
					goto l28
				l29:
					position, tokenIndex, depth = position28, tokenIndex28, depth28
					if buffer[position] != rune('A') {
						goto l26
					}
					position++
				}
			l28:
				{
					position30, tokenIndex30, depth30 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l31
					}
					position++
					goto l30
				l31:
					position, tokenIndex, depth = position30, tokenIndex30, depth30
					if buffer[position] != rune('S') {
						goto l26
					}
					position++
				}
			l30:
				{
					position32, tokenIndex32, depth32 := position, tokenIndex, depth
					if buffer[position] != rune('c') {
						goto l33
					}
					position++
					goto l32
				l33:
					position, tokenIndex, depth = position32, tokenIndex32, depth32
					if buffer[position] != rune('C') {
						goto l26
					}
					position++
				}
			l32:
				depth--
				add(ruleasc, position27)
			}
			return true
		l26:
			position, tokenIndex, depth = position26, tokenIndex26, depth26
			return false
		},
		/* 4 desc <- <(('d' / 'D') ('e' / 'E') ('s' / 'S') ('c' / 'C'))> */
		func() bool {
			position34, tokenIndex34, depth34 := position, tokenIndex, depth
			{
				position35 := position
				depth++
				{
					position36, tokenIndex36, depth36 := position, tokenIndex, depth
					if buffer[position] != rune('d') {
						goto l37
					}
					position++
					goto l36
				l37:
					position, tokenIndex, depth = position36, tokenIndex36, depth36
					if buffer[position] != rune('D') {
						goto l34
					}
					position++
				}
			l36:
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l39
					}
					position++
					goto l38
				l39:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
					if buffer[position] != rune('E') {
						goto l34
					}
					position++
				}
			l38:
				{
					position40, tokenIndex40, depth40 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l41
					}
					position++
					goto l40
				l41:
					position, tokenIndex, depth = position40, tokenIndex40, depth40
					if buffer[position] != rune('S') {
						goto l34
					}
					position++
				}
			l40:
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if buffer[position] != rune('c') {
						goto l43
					}
					position++
					goto l42
				l43:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if buffer[position] != rune('C') {
						goto l34
					}
					position++
				}
			l42:
				depth--
				add(ruledesc, position35)
			}
			return true
		l34:
			position, tokenIndex, depth = position34, tokenIndex34, depth34
			return false
		},
		/* 5 limit <- <(('l' / 'L') ('i' / 'I') ('m' / 'M') ('i' / 'I') ('t' / 'T') sp value1)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				{
					position46, tokenIndex46, depth46 := position, tokenIndex, depth
					if buffer[position] != rune('l') {
						goto l47
					}
					position++
					goto l46
				l47:
					position, tokenIndex, depth = position46, tokenIndex46, depth46
					if buffer[position] != rune('L') {
						goto l44
					}
					position++
				}
			l46:
				{
					position48, tokenIndex48, depth48 := position, tokenIndex, depth
					if buffer[position] != rune('i') {
						goto l49
					}
					position++
					goto l48
				l49:
					position, tokenIndex, depth = position48, tokenIndex48, depth48
					if buffer[position] != rune('I') {
						goto l44
					}
					position++
				}
			l48:
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if buffer[position] != rune('m') {
						goto l51
					}
					position++
					goto l50
				l51:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
					if buffer[position] != rune('M') {
						goto l44
					}
					position++
				}
			l50:
				{
					position52, tokenIndex52, depth52 := position, tokenIndex, depth
					if buffer[position] != rune('i') {
						goto l53
					}
					position++
					goto l52
				l53:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
					if buffer[position] != rune('I') {
						goto l44
					}
					position++
				}
			l52:
				{
					position54, tokenIndex54, depth54 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l55
					}
					position++
					goto l54
				l55:
					position, tokenIndex, depth = position54, tokenIndex54, depth54
					if buffer[position] != rune('T') {
						goto l44
					}
					position++
				}
			l54:
				if !_rules[rulesp]() {
					goto l44
				}
				if !_rules[rulevalue1]() {
					goto l44
				}
				depth--
				add(rulelimit, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 6 offset <- <(('o' / 'O') ('f' / 'F') ('f' / 'F') ('s' / 'S') ('e' / 'E') ('t' / 'T') sp value1)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				{
					position58, tokenIndex58, depth58 := position, tokenIndex, depth
					if buffer[position] != rune('o') {
						goto l59
					}
					position++
					goto l58
				l59:
					position, tokenIndex, depth = position58, tokenIndex58, depth58
					if buffer[position] != rune('O') {
						goto l56
					}
					position++
				}
			l58:
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l61
					}
					position++
					goto l60
				l61:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if buffer[position] != rune('F') {
						goto l56
					}
					position++
				}
			l60:
				{
					position62, tokenIndex62, depth62 := position, tokenIndex, depth
					if buffer[position] != rune('f') {
						goto l63
					}
					position++
					goto l62
				l63:
					position, tokenIndex, depth = position62, tokenIndex62, depth62
					if buffer[position] != rune('F') {
						goto l56
					}
					position++
				}
			l62:
				{
					position64, tokenIndex64, depth64 := position, tokenIndex, depth
					if buffer[position] != rune('s') {
						goto l65
					}
					position++
					goto l64
				l65:
					position, tokenIndex, depth = position64, tokenIndex64, depth64
					if buffer[position] != rune('S') {
						goto l56
					}
					position++
				}
			l64:
				{
					position66, tokenIndex66, depth66 := position, tokenIndex, depth
					if buffer[position] != rune('e') {
						goto l67
					}
					position++
					goto l66
				l67:
					position, tokenIndex, depth = position66, tokenIndex66, depth66
					if buffer[position] != rune('E') {
						goto l56
					}
					position++
				}
			l66:
				{
					position68, tokenIndex68, depth68 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l69
					}
					position++
					goto l68
				l69:
					position, tokenIndex, depth = position68, tokenIndex68, depth68
					if buffer[position] != rune('T') {
						goto l56
					}
					position++
				}
			l68:
				if !_rules[rulesp]() {
					goto l56
				}
				if !_rules[rulevalue1]() {
					goto l56
				}
				depth--
				add(ruleoffset, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 7 e1 <- <(e2 (or e2)*)> */
		func() bool {
			position70, tokenIndex70, depth70 := position, tokenIndex, depth
			{
				position71 := position
				depth++
				if !_rules[rulee2]() {
					goto l70
				}
			l72:
				{
					position73, tokenIndex73, depth73 := position, tokenIndex, depth
					if !_rules[ruleor]() {
						goto l73
					}
					if !_rules[rulee2]() {
						goto l73
					}
					goto l72
				l73:
					position, tokenIndex, depth = position73, tokenIndex73, depth73
				}
				depth--
				add(rulee1, position71)
			}
			return true
		l70:
			position, tokenIndex, depth = position70, tokenIndex70, depth70
			return false
		},
		/* 8 e2 <- <(e3 (and e3)*)> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				if !_rules[rulee3]() {
					goto l74
				}
			l76:
				{
					position77, tokenIndex77, depth77 := position, tokenIndex, depth
					if !_rules[ruleand]() {
						goto l77
					}
					if !_rules[rulee3]() {
						goto l77
					}
					goto l76
				l77:
					position, tokenIndex, depth = position77, tokenIndex77, depth77
				}
				depth--
				add(rulee2, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 9 e3 <- <(expression / (open e1 close))> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if !_rules[ruleexpression]() {
						goto l81
					}
					goto l80
				l81:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
					if !_rules[ruleopen]() {
						goto l78
					}
					if !_rules[rulee1]() {
						goto l78
					}
					if !_rules[ruleclose]() {
						goto l78
					}
				}
			l80:
				depth--
				add(rulee3, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 10 expression <- <((path op value2) / (boolean sp))> */
		func() bool {
			position82, tokenIndex82, depth82 := position, tokenIndex, depth
			{
				position83 := position
				depth++
				{
					position84, tokenIndex84, depth84 := position, tokenIndex, depth
					if !_rules[rulepath]() {
						goto l85
					}
					if !_rules[ruleop]() {
						goto l85
					}
					if !_rules[rulevalue2]() {
						goto l85
					}
					goto l84
				l85:
					position, tokenIndex, depth = position84, tokenIndex84, depth84
					if !_rules[ruleboolean]() {
						goto l82
					}
					if !_rules[rulesp]() {
						goto l82
					}
				}
			l84:
				depth--
				add(ruleexpression, position83)
			}
			return true
		l82:
			position, tokenIndex, depth = position82, tokenIndex82, depth82
			return false
		},
		/* 11 op <- <(('=' / ('!' '=') / ('>' '=') / ('<' '=') / '>' / '<') sp)> */
		func() bool {
			position86, tokenIndex86, depth86 := position, tokenIndex, depth
			{
				position87 := position
				depth++
				{
					position88, tokenIndex88, depth88 := position, tokenIndex, depth
					if buffer[position] != rune('=') {
						goto l89
					}
					position++
					goto l88
				l89:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
					if buffer[position] != rune('!') {
						goto l90
					}
					position++
					if buffer[position] != rune('=') {
						goto l90
					}
					position++
					goto l88
				l90:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
					if buffer[position] != rune('>') {
						goto l91
					}
					position++
					if buffer[position] != rune('=') {
						goto l91
					}
					position++
					goto l88
				l91:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
					if buffer[position] != rune('<') {
						goto l92
					}
					position++
					if buffer[position] != rune('=') {
						goto l92
					}
					position++
					goto l88
				l92:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
					if buffer[position] != rune('>') {
						goto l93
					}
					position++
					goto l88
				l93:
					position, tokenIndex, depth = position88, tokenIndex88, depth88
					if buffer[position] != rune('<') {
						goto l86
					}
					position++
				}
			l88:
				if !_rules[rulesp]() {
					goto l86
				}
				depth--
				add(ruleop, position87)
			}
			return true
		l86:
			position, tokenIndex, depth = position86, tokenIndex86, depth86
			return false
		},
		/* 12 path <- <(word ('.' word)* sp)> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				if !_rules[ruleword]() {
					goto l94
				}
			l96:
				{
					position97, tokenIndex97, depth97 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l97
					}
					position++
					if !_rules[ruleword]() {
						goto l97
					}
					goto l96
				l97:
					position, tokenIndex, depth = position97, tokenIndex97, depth97
				}
				if !_rules[rulesp]() {
					goto l94
				}
				depth--
				add(rulepath, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 13 word <- <([a-z] / [A-Z] / ([0-9] / [0-9]))+> */
		func() bool {
			position98, tokenIndex98, depth98 := position, tokenIndex, depth
			{
				position99 := position
				depth++
				{
					position102, tokenIndex102, depth102 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l103
					}
					position++
					goto l102
				l103:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l104
					}
					position++
					goto l102
				l104:
					position, tokenIndex, depth = position102, tokenIndex102, depth102
					{
						position105, tokenIndex105, depth105 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l106
						}
						position++
						goto l105
					l106:
						position, tokenIndex, depth = position105, tokenIndex105, depth105
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l98
						}
						position++
					}
				l105:
				}
			l102:
			l100:
				{
					position101, tokenIndex101, depth101 := position, tokenIndex, depth
					{
						position107, tokenIndex107, depth107 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l108
						}
						position++
						goto l107
					l108:
						position, tokenIndex, depth = position107, tokenIndex107, depth107
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l109
						}
						position++
						goto l107
					l109:
						position, tokenIndex, depth = position107, tokenIndex107, depth107
						{
							position110, tokenIndex110, depth110 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l111
							}
							position++
							goto l110
						l111:
							position, tokenIndex, depth = position110, tokenIndex110, depth110
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l101
							}
							position++
						}
					l110:
					}
				l107:
					goto l100
				l101:
					position, tokenIndex, depth = position101, tokenIndex101, depth101
				}
				depth--
				add(ruleword, position99)
			}
			return true
		l98:
			position, tokenIndex, depth = position98, tokenIndex98, depth98
			return false
		},
		/* 14 value1 <- <((placeholder / whole) sp)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if !_rules[ruleplaceholder]() {
						goto l115
					}
					goto l114
				l115:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
					if !_rules[rulewhole]() {
						goto l112
					}
				}
			l114:
				if !_rules[rulesp]() {
					goto l112
				}
				depth--
				add(rulevalue1, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 15 value2 <- <((placeholder / string / number / boolean / null) sp)> */
		func() bool {
			position116, tokenIndex116, depth116 := position, tokenIndex, depth
			{
				position117 := position
				depth++
				{
					position118, tokenIndex118, depth118 := position, tokenIndex, depth
					if !_rules[ruleplaceholder]() {
						goto l119
					}
					goto l118
				l119:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[rulestring]() {
						goto l120
					}
					goto l118
				l120:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[rulenumber]() {
						goto l121
					}
					goto l118
				l121:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[ruleboolean]() {
						goto l122
					}
					goto l118
				l122:
					position, tokenIndex, depth = position118, tokenIndex118, depth118
					if !_rules[rulenull]() {
						goto l116
					}
				}
			l118:
				if !_rules[rulesp]() {
					goto l116
				}
				depth--
				add(rulevalue2, position117)
			}
			return true
		l116:
			position, tokenIndex, depth = position116, tokenIndex116, depth116
			return false
		},
		/* 16 placeholder <- <('$' [0-9]+)> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				if buffer[position] != rune('$') {
					goto l123
				}
				position++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l123
				}
				position++
			l125:
				{
					position126, tokenIndex126, depth126 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l126
					}
					position++
					goto l125
				l126:
					position, tokenIndex, depth = position126, tokenIndex126, depth126
				}
				depth--
				add(ruleplaceholder, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 17 string <- <('\'' (!'\'' ([a-z] / [A-Z] / ([0-9] / [0-9])))* '\'')> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				if buffer[position] != rune('\'') {
					goto l127
				}
				position++
			l129:
				{
					position130, tokenIndex130, depth130 := position, tokenIndex, depth
					{
						position131, tokenIndex131, depth131 := position, tokenIndex, depth
						if buffer[position] != rune('\'') {
							goto l131
						}
						position++
						goto l130
					l131:
						position, tokenIndex, depth = position131, tokenIndex131, depth131
					}
					{
						position132, tokenIndex132, depth132 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l133
						}
						position++
						goto l132
					l133:
						position, tokenIndex, depth = position132, tokenIndex132, depth132
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l134
						}
						position++
						goto l132
					l134:
						position, tokenIndex, depth = position132, tokenIndex132, depth132
						{
							position135, tokenIndex135, depth135 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l136
							}
							position++
							goto l135
						l136:
							position, tokenIndex, depth = position135, tokenIndex135, depth135
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l130
							}
							position++
						}
					l135:
					}
				l132:
					goto l129
				l130:
					position, tokenIndex, depth = position130, tokenIndex130, depth130
				}
				if buffer[position] != rune('\'') {
					goto l127
				}
				position++
				depth--
				add(rulestring, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 18 number <- <([0-9]+ ('.' [0-9]+)?)> */
		func() bool {
			position137, tokenIndex137, depth137 := position, tokenIndex, depth
			{
				position138 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l137
				}
				position++
			l139:
				{
					position140, tokenIndex140, depth140 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l140
					}
					position++
					goto l139
				l140:
					position, tokenIndex, depth = position140, tokenIndex140, depth140
				}
				{
					position141, tokenIndex141, depth141 := position, tokenIndex, depth
					if buffer[position] != rune('.') {
						goto l141
					}
					position++
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l141
					}
					position++
				l143:
					{
						position144, tokenIndex144, depth144 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l144
						}
						position++
						goto l143
					l144:
						position, tokenIndex, depth = position144, tokenIndex144, depth144
					}
					goto l142
				l141:
					position, tokenIndex, depth = position141, tokenIndex141, depth141
				}
			l142:
				depth--
				add(rulenumber, position138)
			}
			return true
		l137:
			position, tokenIndex, depth = position137, tokenIndex137, depth137
			return false
		},
		/* 19 boolean <- <(('t' 'r' 'u' 'e') / ('f' 'a' 'l' 's' 'e'))> */
		func() bool {
			position145, tokenIndex145, depth145 := position, tokenIndex, depth
			{
				position146 := position
				depth++
				{
					position147, tokenIndex147, depth147 := position, tokenIndex, depth
					if buffer[position] != rune('t') {
						goto l148
					}
					position++
					if buffer[position] != rune('r') {
						goto l148
					}
					position++
					if buffer[position] != rune('u') {
						goto l148
					}
					position++
					if buffer[position] != rune('e') {
						goto l148
					}
					position++
					goto l147
				l148:
					position, tokenIndex, depth = position147, tokenIndex147, depth147
					if buffer[position] != rune('f') {
						goto l145
					}
					position++
					if buffer[position] != rune('a') {
						goto l145
					}
					position++
					if buffer[position] != rune('l') {
						goto l145
					}
					position++
					if buffer[position] != rune('s') {
						goto l145
					}
					position++
					if buffer[position] != rune('e') {
						goto l145
					}
					position++
				}
			l147:
				depth--
				add(ruleboolean, position146)
			}
			return true
		l145:
			position, tokenIndex, depth = position145, tokenIndex145, depth145
			return false
		},
		/* 20 null <- <('n' 'u' 'l' 'l')> */
		func() bool {
			position149, tokenIndex149, depth149 := position, tokenIndex, depth
			{
				position150 := position
				depth++
				if buffer[position] != rune('n') {
					goto l149
				}
				position++
				if buffer[position] != rune('u') {
					goto l149
				}
				position++
				if buffer[position] != rune('l') {
					goto l149
				}
				position++
				if buffer[position] != rune('l') {
					goto l149
				}
				position++
				depth--
				add(rulenull, position150)
			}
			return true
		l149:
			position, tokenIndex, depth = position149, tokenIndex149, depth149
			return false
		},
		/* 21 whole <- <[0-9]+> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l151
				}
				position++
			l153:
				{
					position154, tokenIndex154, depth154 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l154
					}
					position++
					goto l153
				l154:
					position, tokenIndex, depth = position154, tokenIndex154, depth154
				}
				depth--
				add(rulewhole, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 22 and <- <('a' 'n' 'd' sp)> */
		func() bool {
			position155, tokenIndex155, depth155 := position, tokenIndex, depth
			{
				position156 := position
				depth++
				if buffer[position] != rune('a') {
					goto l155
				}
				position++
				if buffer[position] != rune('n') {
					goto l155
				}
				position++
				if buffer[position] != rune('d') {
					goto l155
				}
				position++
				if !_rules[rulesp]() {
					goto l155
				}
				depth--
				add(ruleand, position156)
			}
			return true
		l155:
			position, tokenIndex, depth = position155, tokenIndex155, depth155
			return false
		},
		/* 23 or <- <('o' 'r' sp)> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if buffer[position] != rune('o') {
					goto l157
				}
				position++
				if buffer[position] != rune('r') {
					goto l157
				}
				position++
				if !_rules[rulesp]() {
					goto l157
				}
				depth--
				add(ruleor, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 24 open <- <('(' sp)> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				if buffer[position] != rune('(') {
					goto l159
				}
				position++
				if !_rules[rulesp]() {
					goto l159
				}
				depth--
				add(ruleopen, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 25 close <- <(')' sp)> */
		func() bool {
			position161, tokenIndex161, depth161 := position, tokenIndex, depth
			{
				position162 := position
				depth++
				if buffer[position] != rune(')') {
					goto l161
				}
				position++
				if !_rules[rulesp]() {
					goto l161
				}
				depth--
				add(ruleclose, position162)
			}
			return true
		l161:
			position, tokenIndex, depth = position161, tokenIndex161, depth161
			return false
		},
		/* 26 sp <- <(' ' / '\t')*> */
		func() bool {
			{
				position164 := position
				depth++
			l165:
				{
					position166, tokenIndex166, depth166 := position, tokenIndex, depth
					{
						position167, tokenIndex167, depth167 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l168
						}
						position++
						goto l167
					l168:
						position, tokenIndex, depth = position167, tokenIndex167, depth167
						if buffer[position] != rune('\t') {
							goto l166
						}
						position++
					}
				l167:
					goto l165
				l166:
					position, tokenIndex, depth = position166, tokenIndex166, depth166
				}
				depth--
				add(rulesp, position164)
			}
			return true
		},
	}
	p.rules = _rules
}
