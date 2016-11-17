package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleGrammar
	ruleImport
	ruleDefinition
	ruleExpression
	ruleSequence
	rulePrefix
	ruleSuffix
	rulePrimary
	ruleIdentifier
	ruleIdentStart
	ruleIdentCont
	ruleLiteral
	ruleClass
	ruleRanges
	ruleDoubleRanges
	ruleRange
	ruleDoubleRange
	ruleChar
	ruleDoubleChar
	ruleEscape
	ruleLeftArrow
	ruleSlash
	ruleAnd
	ruleNot
	ruleQuestion
	ruleStar
	rulePlus
	ruleOpen
	ruleClose
	ruleDot
	ruleSpaceComment
	ruleSpacing
	ruleMustSpacing
	ruleComment
	ruleSpace
	ruleEndOfLine
	ruleEndOfFile
	ruleAction
	ruleActionBody
	ruleBegin
	ruleEnd
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25
	ruleAction26
	ruleAction27
	ruleAction28
	ruleAction29
	ruleAction30
	ruleAction31
	ruleAction32
	ruleAction33
	ruleAction34
	ruleAction35
	ruleAction36
	ruleAction37
	ruleAction38
	ruleAction39
	ruleAction40
	ruleAction41
	ruleAction42
	ruleAction43
	ruleAction44
	ruleAction45
	ruleAction46
	ruleAction47
	ruleAction48
)

var rul3s = [...]string{
	"Unknown",
	"Grammar",
	"Import",
	"Definition",
	"Expression",
	"Sequence",
	"Prefix",
	"Suffix",
	"Primary",
	"Identifier",
	"IdentStart",
	"IdentCont",
	"Literal",
	"Class",
	"Ranges",
	"DoubleRanges",
	"Range",
	"DoubleRange",
	"Char",
	"DoubleChar",
	"Escape",
	"LeftArrow",
	"Slash",
	"And",
	"Not",
	"Question",
	"Star",
	"Plus",
	"Open",
	"Close",
	"Dot",
	"SpaceComment",
	"Spacing",
	"MustSpacing",
	"Comment",
	"Space",
	"EndOfLine",
	"EndOfFile",
	"Action",
	"ActionBody",
	"Begin",
	"End",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",
	"Action26",
	"Action27",
	"Action28",
	"Action29",
	"Action30",
	"Action31",
	"Action32",
	"Action33",
	"Action34",
	"Action35",
	"Action36",
	"Action37",
	"Action38",
	"Action39",
	"Action40",
	"Action41",
	"Action42",
	"Action43",
	"Action44",
	"Action45",
	"Action46",
	"Action47",
	"Action48",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) Print(buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Printf(" ")
			}
			fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
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
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type Peg struct {
	*Tree

	Buffer string
	buffer []rune
	rules  [92]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokens32
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
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
	p   *Peg
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *Peg) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *Peg) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.AddPackage(text)
		case ruleAction1:
			p.AddPeg(text)
		case ruleAction2:
			p.AddState(text)
		case ruleAction3:
			p.AddImport(text)
		case ruleAction4:
			p.AddRule(text)
		case ruleAction5:
			p.AddExpression()
		case ruleAction6:
			p.AddAlternate()
		case ruleAction7:
			p.AddNil()
			p.AddAlternate()
		case ruleAction8:
			p.AddNil()
		case ruleAction9:
			p.AddSequence()
		case ruleAction10:
			p.AddPredicate(text)
		case ruleAction11:
			p.AddStateChange(text)
		case ruleAction12:
			p.AddPeekFor()
		case ruleAction13:
			p.AddPeekNot()
		case ruleAction14:
			p.AddQuery()
		case ruleAction15:
			p.AddStar()
		case ruleAction16:
			p.AddPlus()
		case ruleAction17:
			p.AddName(text)
		case ruleAction18:
			p.AddDot()
		case ruleAction19:
			p.AddAction(text)
		case ruleAction20:
			p.AddPush()
		case ruleAction21:
			p.AddSequence()
		case ruleAction22:
			p.AddSequence()
		case ruleAction23:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case ruleAction24:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case ruleAction25:
			p.AddAlternate()
		case ruleAction26:
			p.AddAlternate()
		case ruleAction27:
			p.AddRange()
		case ruleAction28:
			p.AddDoubleRange()
		case ruleAction29:
			p.AddCharacter(text)
		case ruleAction30:
			p.AddDoubleCharacter(text)
		case ruleAction31:
			p.AddCharacter(text)
		case ruleAction32:
			p.AddCharacter("\a")
		case ruleAction33:
			p.AddCharacter("\b")
		case ruleAction34:
			p.AddCharacter("\x1B")
		case ruleAction35:
			p.AddCharacter("\f")
		case ruleAction36:
			p.AddCharacter("\n")
		case ruleAction37:
			p.AddCharacter("\r")
		case ruleAction38:
			p.AddCharacter("\t")
		case ruleAction39:
			p.AddCharacter("\v")
		case ruleAction40:
			p.AddCharacter("'")
		case ruleAction41:
			p.AddCharacter("\"")
		case ruleAction42:
			p.AddCharacter("[")
		case ruleAction43:
			p.AddCharacter("]")
		case ruleAction44:
			p.AddCharacter("-")
		case ruleAction45:
			p.AddHexaCharacter(text)
		case ruleAction46:
			p.AddOctalCharacter(text)
		case ruleAction47:
			p.AddOctalCharacter(text)
		case ruleAction48:
			p.AddCharacter("\\")

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Peg) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.Reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.Reset()

	_rules, tree := p.rules, tokens32{tree: make([]token32, math.MaxInt16)}
	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
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
		/* 0 Grammar <- <(Spacing ('p' 'a' 'c' 'k' 'a' 'g' 'e') MustSpacing Identifier Action0 Import* ('t' 'y' 'p' 'e') MustSpacing Identifier Action1 ('P' 'e' 'g') Spacing Action Action2 Definition+ EndOfFile)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[ruleSpacing]() {
					goto l0
				}
				if buffer[position] != rune('p') {
					goto l0
				}
				position++
				if buffer[position] != rune('a') {
					goto l0
				}
				position++
				if buffer[position] != rune('c') {
					goto l0
				}
				position++
				if buffer[position] != rune('k') {
					goto l0
				}
				position++
				if buffer[position] != rune('a') {
					goto l0
				}
				position++
				if buffer[position] != rune('g') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if !_rules[ruleMustSpacing]() {
					goto l0
				}
				if !_rules[ruleIdentifier]() {
					goto l0
				}
				{
					add(ruleAction0, position)
				}
			l3:
				{
					position4, tokenIndex4 := position, tokenIndex
					{
						position5 := position
						if buffer[position] != rune('i') {
							goto l4
						}
						position++
						if buffer[position] != rune('m') {
							goto l4
						}
						position++
						if buffer[position] != rune('p') {
							goto l4
						}
						position++
						if buffer[position] != rune('o') {
							goto l4
						}
						position++
						if buffer[position] != rune('r') {
							goto l4
						}
						position++
						if buffer[position] != rune('t') {
							goto l4
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l4
						}
						if buffer[position] != rune('"') {
							goto l4
						}
						position++
						{
							position6 := position
							{
								switch buffer[position] {
								case '-':
									if buffer[position] != rune('-') {
										goto l4
									}
									position++
									break
								case '.':
									if buffer[position] != rune('.') {
										goto l4
									}
									position++
									break
								case '/':
									if buffer[position] != rune('/') {
										goto l4
									}
									position++
									break
								case '_':
									if buffer[position] != rune('_') {
										goto l4
									}
									position++
									break
								case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
									if c := buffer[position]; c < rune('A') || c > rune('Z') {
										goto l4
									}
									position++
									break
								default:
									if c := buffer[position]; c < rune('a') || c > rune('z') {
										goto l4
									}
									position++
									break
								}
							}

						l7:
							{
								position8, tokenIndex8 := position, tokenIndex
								{
									switch buffer[position] {
									case '-':
										if buffer[position] != rune('-') {
											goto l8
										}
										position++
										break
									case '.':
										if buffer[position] != rune('.') {
											goto l8
										}
										position++
										break
									case '/':
										if buffer[position] != rune('/') {
											goto l8
										}
										position++
										break
									case '_':
										if buffer[position] != rune('_') {
											goto l8
										}
										position++
										break
									case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
										if c := buffer[position]; c < rune('A') || c > rune('Z') {
											goto l8
										}
										position++
										break
									default:
										if c := buffer[position]; c < rune('a') || c > rune('z') {
											goto l8
										}
										position++
										break
									}
								}

								goto l7
							l8:
								position, tokenIndex = position8, tokenIndex8
							}
							add(rulePegText, position6)
						}
						if buffer[position] != rune('"') {
							goto l4
						}
						position++
						if !_rules[ruleSpacing]() {
							goto l4
						}
						{
							add(ruleAction3, position)
						}
						add(ruleImport, position5)
					}
					goto l3
				l4:
					position, tokenIndex = position4, tokenIndex4
				}
				if buffer[position] != rune('t') {
					goto l0
				}
				position++
				if buffer[position] != rune('y') {
					goto l0
				}
				position++
				if buffer[position] != rune('p') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if !_rules[ruleMustSpacing]() {
					goto l0
				}
				if !_rules[ruleIdentifier]() {
					goto l0
				}
				{
					add(ruleAction1, position)
				}
				if buffer[position] != rune('P') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if buffer[position] != rune('g') {
					goto l0
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l0
				}
				if !_rules[ruleAction]() {
					goto l0
				}
				{
					add(ruleAction2, position)
				}
				{
					position16 := position
					if !_rules[ruleIdentifier]() {
						goto l0
					}
					{
						add(ruleAction4, position)
					}
					if !_rules[ruleLeftArrow]() {
						goto l0
					}
					if !_rules[ruleExpression]() {
						goto l0
					}
					{
						add(ruleAction5, position)
					}
					{
						position19, tokenIndex19 := position, tokenIndex
						{
							position20, tokenIndex20 := position, tokenIndex
							if !_rules[ruleIdentifier]() {
								goto l21
							}
							if !_rules[ruleLeftArrow]() {
								goto l21
							}
							goto l20
						l21:
							position, tokenIndex = position20, tokenIndex20
							{
								position22, tokenIndex22 := position, tokenIndex
								if !matchDot() {
									goto l22
								}
								goto l0
							l22:
								position, tokenIndex = position22, tokenIndex22
							}
						}
					l20:
						position, tokenIndex = position19, tokenIndex19
					}
					add(ruleDefinition, position16)
				}
			l14:
				{
					position15, tokenIndex15 := position, tokenIndex
					{
						position23 := position
						if !_rules[ruleIdentifier]() {
							goto l15
						}
						{
							add(ruleAction4, position)
						}
						if !_rules[ruleLeftArrow]() {
							goto l15
						}
						if !_rules[ruleExpression]() {
							goto l15
						}
						{
							add(ruleAction5, position)
						}
						{
							position26, tokenIndex26 := position, tokenIndex
							{
								position27, tokenIndex27 := position, tokenIndex
								if !_rules[ruleIdentifier]() {
									goto l28
								}
								if !_rules[ruleLeftArrow]() {
									goto l28
								}
								goto l27
							l28:
								position, tokenIndex = position27, tokenIndex27
								{
									position29, tokenIndex29 := position, tokenIndex
									if !matchDot() {
										goto l29
									}
									goto l15
								l29:
									position, tokenIndex = position29, tokenIndex29
								}
							}
						l27:
							position, tokenIndex = position26, tokenIndex26
						}
						add(ruleDefinition, position23)
					}
					goto l14
				l15:
					position, tokenIndex = position15, tokenIndex15
				}
				{
					position30 := position
					{
						position31, tokenIndex31 := position, tokenIndex
						if !matchDot() {
							goto l31
						}
						goto l0
					l31:
						position, tokenIndex = position31, tokenIndex31
					}
					add(ruleEndOfFile, position30)
				}
				add(ruleGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 Import <- <('i' 'm' 'p' 'o' 'r' 't' Spacing '"' <((&('-') '-') | (&('.') '.') | (&('/') '/') | (&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))+> '"' Spacing Action3)> */
		nil,
		/* 2 Definition <- <(Identifier Action4 LeftArrow Expression Action5 &((Identifier LeftArrow) / !.))> */
		nil,
		/* 3 Expression <- <((Sequence (Slash Sequence Action6)* (Slash Action7)?) / Action8)> */
		func() bool {
			{
				position35 := position
				{
					position36, tokenIndex36 := position, tokenIndex
					if !_rules[ruleSequence]() {
						goto l37
					}
				l38:
					{
						position39, tokenIndex39 := position, tokenIndex
						if !_rules[ruleSlash]() {
							goto l39
						}
						if !_rules[ruleSequence]() {
							goto l39
						}
						{
							add(ruleAction6, position)
						}
						goto l38
					l39:
						position, tokenIndex = position39, tokenIndex39
					}
					{
						position41, tokenIndex41 := position, tokenIndex
						if !_rules[ruleSlash]() {
							goto l41
						}
						{
							add(ruleAction7, position)
						}
						goto l42
					l41:
						position, tokenIndex = position41, tokenIndex41
					}
				l42:
					goto l36
				l37:
					position, tokenIndex = position36, tokenIndex36
					{
						add(ruleAction8, position)
					}
				}
			l36:
				add(ruleExpression, position35)
			}
			return true
		},
		/* 4 Sequence <- <(Prefix (Prefix Action9)*)> */
		func() bool {
			position45, tokenIndex45 := position, tokenIndex
			{
				position46 := position
				if !_rules[rulePrefix]() {
					goto l45
				}
			l47:
				{
					position48, tokenIndex48 := position, tokenIndex
					if !_rules[rulePrefix]() {
						goto l48
					}
					{
						add(ruleAction9, position)
					}
					goto l47
				l48:
					position, tokenIndex = position48, tokenIndex48
				}
				add(ruleSequence, position46)
			}
			return true
		l45:
			position, tokenIndex = position45, tokenIndex45
			return false
		},
		/* 5 Prefix <- <((And Action Action10) / (Not Action Action11) / ((&('!') (Not Suffix Action13)) | (&('&') (And Suffix Action12)) | (&('"' | '\'' | '(' | '.' | '<' | 'A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '[' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z' | '{') Suffix)))> */
		func() bool {
			position50, tokenIndex50 := position, tokenIndex
			{
				position51 := position
				{
					position52, tokenIndex52 := position, tokenIndex
					if !_rules[ruleAnd]() {
						goto l53
					}
					if !_rules[ruleAction]() {
						goto l53
					}
					{
						add(ruleAction10, position)
					}
					goto l52
				l53:
					position, tokenIndex = position52, tokenIndex52
					if !_rules[ruleNot]() {
						goto l55
					}
					if !_rules[ruleAction]() {
						goto l55
					}
					{
						add(ruleAction11, position)
					}
					goto l52
				l55:
					position, tokenIndex = position52, tokenIndex52
					{
						switch buffer[position] {
						case '!':
							if !_rules[ruleNot]() {
								goto l50
							}
							if !_rules[ruleSuffix]() {
								goto l50
							}
							{
								add(ruleAction13, position)
							}
							break
						case '&':
							if !_rules[ruleAnd]() {
								goto l50
							}
							if !_rules[ruleSuffix]() {
								goto l50
							}
							{
								add(ruleAction12, position)
							}
							break
						default:
							if !_rules[ruleSuffix]() {
								goto l50
							}
							break
						}
					}

				}
			l52:
				add(rulePrefix, position51)
			}
			return true
		l50:
			position, tokenIndex = position50, tokenIndex50
			return false
		},
		/* 6 Suffix <- <(Primary ((&('+') (Plus Action16)) | (&('*') (Star Action15)) | (&('?') (Question Action14)))?)> */
		func() bool {
			position60, tokenIndex60 := position, tokenIndex
			{
				position61 := position
				{
					position62 := position
					{
						switch buffer[position] {
						case '<':
							{
								position64 := position
								if buffer[position] != rune('<') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								add(ruleBegin, position64)
							}
							if !_rules[ruleExpression]() {
								goto l60
							}
							{
								position65 := position
								if buffer[position] != rune('>') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								add(ruleEnd, position65)
							}
							{
								add(ruleAction20, position)
							}
							break
						case '{':
							if !_rules[ruleAction]() {
								goto l60
							}
							{
								add(ruleAction19, position)
							}
							break
						case '.':
							{
								position68 := position
								if buffer[position] != rune('.') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								add(ruleDot, position68)
							}
							{
								add(ruleAction18, position)
							}
							break
						case '[':
							{
								position70 := position
								{
									position71, tokenIndex71 := position, tokenIndex
									if buffer[position] != rune('[') {
										goto l72
									}
									position++
									if buffer[position] != rune('[') {
										goto l72
									}
									position++
									{
										position73, tokenIndex73 := position, tokenIndex
										{
											position75, tokenIndex75 := position, tokenIndex
											if buffer[position] != rune('^') {
												goto l76
											}
											position++
											if !_rules[ruleDoubleRanges]() {
												goto l76
											}
											{
												add(ruleAction23, position)
											}
											goto l75
										l76:
											position, tokenIndex = position75, tokenIndex75
											if !_rules[ruleDoubleRanges]() {
												goto l73
											}
										}
									l75:
										goto l74
									l73:
										position, tokenIndex = position73, tokenIndex73
									}
								l74:
									if buffer[position] != rune(']') {
										goto l72
									}
									position++
									if buffer[position] != rune(']') {
										goto l72
									}
									position++
									goto l71
								l72:
									position, tokenIndex = position71, tokenIndex71
									if buffer[position] != rune('[') {
										goto l60
									}
									position++
									{
										position78, tokenIndex78 := position, tokenIndex
										{
											position80, tokenIndex80 := position, tokenIndex
											if buffer[position] != rune('^') {
												goto l81
											}
											position++
											if !_rules[ruleRanges]() {
												goto l81
											}
											{
												add(ruleAction24, position)
											}
											goto l80
										l81:
											position, tokenIndex = position80, tokenIndex80
											if !_rules[ruleRanges]() {
												goto l78
											}
										}
									l80:
										goto l79
									l78:
										position, tokenIndex = position78, tokenIndex78
									}
								l79:
									if buffer[position] != rune(']') {
										goto l60
									}
									position++
								}
							l71:
								if !_rules[ruleSpacing]() {
									goto l60
								}
								add(ruleClass, position70)
							}
							break
						case '"', '\'':
							{
								position83 := position
								{
									position84, tokenIndex84 := position, tokenIndex
									if buffer[position] != rune('\'') {
										goto l85
									}
									position++
									{
										position86, tokenIndex86 := position, tokenIndex
										{
											position88, tokenIndex88 := position, tokenIndex
											if buffer[position] != rune('\'') {
												goto l88
											}
											position++
											goto l86
										l88:
											position, tokenIndex = position88, tokenIndex88
										}
										if !_rules[ruleChar]() {
											goto l86
										}
										goto l87
									l86:
										position, tokenIndex = position86, tokenIndex86
									}
								l87:
								l89:
									{
										position90, tokenIndex90 := position, tokenIndex
										{
											position91, tokenIndex91 := position, tokenIndex
											if buffer[position] != rune('\'') {
												goto l91
											}
											position++
											goto l90
										l91:
											position, tokenIndex = position91, tokenIndex91
										}
										if !_rules[ruleChar]() {
											goto l90
										}
										{
											add(ruleAction21, position)
										}
										goto l89
									l90:
										position, tokenIndex = position90, tokenIndex90
									}
									if buffer[position] != rune('\'') {
										goto l85
									}
									position++
									if !_rules[ruleSpacing]() {
										goto l85
									}
									goto l84
								l85:
									position, tokenIndex = position84, tokenIndex84
									if buffer[position] != rune('"') {
										goto l60
									}
									position++
									{
										position93, tokenIndex93 := position, tokenIndex
										{
											position95, tokenIndex95 := position, tokenIndex
											if buffer[position] != rune('"') {
												goto l95
											}
											position++
											goto l93
										l95:
											position, tokenIndex = position95, tokenIndex95
										}
										if !_rules[ruleDoubleChar]() {
											goto l93
										}
										goto l94
									l93:
										position, tokenIndex = position93, tokenIndex93
									}
								l94:
								l96:
									{
										position97, tokenIndex97 := position, tokenIndex
										{
											position98, tokenIndex98 := position, tokenIndex
											if buffer[position] != rune('"') {
												goto l98
											}
											position++
											goto l97
										l98:
											position, tokenIndex = position98, tokenIndex98
										}
										if !_rules[ruleDoubleChar]() {
											goto l97
										}
										{
											add(ruleAction22, position)
										}
										goto l96
									l97:
										position, tokenIndex = position97, tokenIndex97
									}
									if buffer[position] != rune('"') {
										goto l60
									}
									position++
									if !_rules[ruleSpacing]() {
										goto l60
									}
								}
							l84:
								add(ruleLiteral, position83)
							}
							break
						case '(':
							{
								position100 := position
								if buffer[position] != rune('(') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								add(ruleOpen, position100)
							}
							if !_rules[ruleExpression]() {
								goto l60
							}
							{
								position101 := position
								if buffer[position] != rune(')') {
									goto l60
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l60
								}
								add(ruleClose, position101)
							}
							break
						default:
							if !_rules[ruleIdentifier]() {
								goto l60
							}
							{
								position102, tokenIndex102 := position, tokenIndex
								if !_rules[ruleLeftArrow]() {
									goto l102
								}
								goto l60
							l102:
								position, tokenIndex = position102, tokenIndex102
							}
							{
								add(ruleAction17, position)
							}
							break
						}
					}

					add(rulePrimary, position62)
				}
				{
					position104, tokenIndex104 := position, tokenIndex
					{
						switch buffer[position] {
						case '+':
							{
								position107 := position
								if buffer[position] != rune('+') {
									goto l104
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l104
								}
								add(rulePlus, position107)
							}
							{
								add(ruleAction16, position)
							}
							break
						case '*':
							{
								position109 := position
								if buffer[position] != rune('*') {
									goto l104
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l104
								}
								add(ruleStar, position109)
							}
							{
								add(ruleAction15, position)
							}
							break
						default:
							{
								position111 := position
								if buffer[position] != rune('?') {
									goto l104
								}
								position++
								if !_rules[ruleSpacing]() {
									goto l104
								}
								add(ruleQuestion, position111)
							}
							{
								add(ruleAction14, position)
							}
							break
						}
					}

					goto l105
				l104:
					position, tokenIndex = position104, tokenIndex104
				}
			l105:
				add(ruleSuffix, position61)
			}
			return true
		l60:
			position, tokenIndex = position60, tokenIndex60
			return false
		},
		/* 7 Primary <- <((&('<') (Begin Expression End Action20)) | (&('{') (Action Action19)) | (&('.') (Dot Action18)) | (&('[') Class) | (&('"' | '\'') Literal) | (&('(') (Open Expression Close)) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z' | '_' | 'a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') (Identifier !LeftArrow Action17)))> */
		nil,
		/* 8 Identifier <- <(<(IdentStart IdentCont*)> Spacing)> */
		func() bool {
			position114, tokenIndex114 := position, tokenIndex
			{
				position115 := position
				{
					position116 := position
					if !_rules[ruleIdentStart]() {
						goto l114
					}
				l117:
					{
						position118, tokenIndex118 := position, tokenIndex
						{
							position119 := position
							{
								position120, tokenIndex120 := position, tokenIndex
								if !_rules[ruleIdentStart]() {
									goto l121
								}
								goto l120
							l121:
								position, tokenIndex = position120, tokenIndex120
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l118
								}
								position++
							}
						l120:
							add(ruleIdentCont, position119)
						}
						goto l117
					l118:
						position, tokenIndex = position118, tokenIndex118
					}
					add(rulePegText, position116)
				}
				if !_rules[ruleSpacing]() {
					goto l114
				}
				add(ruleIdentifier, position115)
			}
			return true
		l114:
			position, tokenIndex = position114, tokenIndex114
			return false
		},
		/* 9 IdentStart <- <((&('_') '_') | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position122, tokenIndex122 := position, tokenIndex
			{
				position123 := position
				{
					switch buffer[position] {
					case '_':
						if buffer[position] != rune('_') {
							goto l122
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l122
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l122
						}
						position++
						break
					}
				}

				add(ruleIdentStart, position123)
			}
			return true
		l122:
			position, tokenIndex = position122, tokenIndex122
			return false
		},
		/* 10 IdentCont <- <(IdentStart / [0-9])> */
		nil,
		/* 11 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char Action21)* '\'' Spacing) / ('"' (!'"' DoubleChar)? (!'"' DoubleChar Action22)* '"' Spacing))> */
		nil,
		/* 12 Class <- <((('[' '[' (('^' DoubleRanges Action23) / DoubleRanges)? (']' ']')) / ('[' (('^' Ranges Action24) / Ranges)? ']')) Spacing)> */
		nil,
		/* 13 Ranges <- <(!']' Range (!']' Range Action25)*)> */
		func() bool {
			position128, tokenIndex128 := position, tokenIndex
			{
				position129 := position
				{
					position130, tokenIndex130 := position, tokenIndex
					if buffer[position] != rune(']') {
						goto l130
					}
					position++
					goto l128
				l130:
					position, tokenIndex = position130, tokenIndex130
				}
				if !_rules[ruleRange]() {
					goto l128
				}
			l131:
				{
					position132, tokenIndex132 := position, tokenIndex
					{
						position133, tokenIndex133 := position, tokenIndex
						if buffer[position] != rune(']') {
							goto l133
						}
						position++
						goto l132
					l133:
						position, tokenIndex = position133, tokenIndex133
					}
					if !_rules[ruleRange]() {
						goto l132
					}
					{
						add(ruleAction25, position)
					}
					goto l131
				l132:
					position, tokenIndex = position132, tokenIndex132
				}
				add(ruleRanges, position129)
			}
			return true
		l128:
			position, tokenIndex = position128, tokenIndex128
			return false
		},
		/* 14 DoubleRanges <- <(!(']' ']') DoubleRange (!(']' ']') DoubleRange Action26)*)> */
		func() bool {
			position135, tokenIndex135 := position, tokenIndex
			{
				position136 := position
				{
					position137, tokenIndex137 := position, tokenIndex
					if buffer[position] != rune(']') {
						goto l137
					}
					position++
					if buffer[position] != rune(']') {
						goto l137
					}
					position++
					goto l135
				l137:
					position, tokenIndex = position137, tokenIndex137
				}
				if !_rules[ruleDoubleRange]() {
					goto l135
				}
			l138:
				{
					position139, tokenIndex139 := position, tokenIndex
					{
						position140, tokenIndex140 := position, tokenIndex
						if buffer[position] != rune(']') {
							goto l140
						}
						position++
						if buffer[position] != rune(']') {
							goto l140
						}
						position++
						goto l139
					l140:
						position, tokenIndex = position140, tokenIndex140
					}
					if !_rules[ruleDoubleRange]() {
						goto l139
					}
					{
						add(ruleAction26, position)
					}
					goto l138
				l139:
					position, tokenIndex = position139, tokenIndex139
				}
				add(ruleDoubleRanges, position136)
			}
			return true
		l135:
			position, tokenIndex = position135, tokenIndex135
			return false
		},
		/* 15 Range <- <((Char '-' Char Action27) / Char)> */
		func() bool {
			position142, tokenIndex142 := position, tokenIndex
			{
				position143 := position
				{
					position144, tokenIndex144 := position, tokenIndex
					if !_rules[ruleChar]() {
						goto l145
					}
					if buffer[position] != rune('-') {
						goto l145
					}
					position++
					if !_rules[ruleChar]() {
						goto l145
					}
					{
						add(ruleAction27, position)
					}
					goto l144
				l145:
					position, tokenIndex = position144, tokenIndex144
					if !_rules[ruleChar]() {
						goto l142
					}
				}
			l144:
				add(ruleRange, position143)
			}
			return true
		l142:
			position, tokenIndex = position142, tokenIndex142
			return false
		},
		/* 16 DoubleRange <- <((Char '-' Char Action28) / DoubleChar)> */
		func() bool {
			position147, tokenIndex147 := position, tokenIndex
			{
				position148 := position
				{
					position149, tokenIndex149 := position, tokenIndex
					if !_rules[ruleChar]() {
						goto l150
					}
					if buffer[position] != rune('-') {
						goto l150
					}
					position++
					if !_rules[ruleChar]() {
						goto l150
					}
					{
						add(ruleAction28, position)
					}
					goto l149
				l150:
					position, tokenIndex = position149, tokenIndex149
					if !_rules[ruleDoubleChar]() {
						goto l147
					}
				}
			l149:
				add(ruleDoubleRange, position148)
			}
			return true
		l147:
			position, tokenIndex = position147, tokenIndex147
			return false
		},
		/* 17 Char <- <(Escape / (!'\\' <.> Action29))> */
		func() bool {
			position152, tokenIndex152 := position, tokenIndex
			{
				position153 := position
				{
					position154, tokenIndex154 := position, tokenIndex
					if !_rules[ruleEscape]() {
						goto l155
					}
					goto l154
				l155:
					position, tokenIndex = position154, tokenIndex154
					{
						position156, tokenIndex156 := position, tokenIndex
						if buffer[position] != rune('\\') {
							goto l156
						}
						position++
						goto l152
					l156:
						position, tokenIndex = position156, tokenIndex156
					}
					{
						position157 := position
						if !matchDot() {
							goto l152
						}
						add(rulePegText, position157)
					}
					{
						add(ruleAction29, position)
					}
				}
			l154:
				add(ruleChar, position153)
			}
			return true
		l152:
			position, tokenIndex = position152, tokenIndex152
			return false
		},
		/* 18 DoubleChar <- <(Escape / (<([a-z] / [A-Z])> Action30) / (!'\\' <.> Action31))> */
		func() bool {
			position159, tokenIndex159 := position, tokenIndex
			{
				position160 := position
				{
					position161, tokenIndex161 := position, tokenIndex
					if !_rules[ruleEscape]() {
						goto l162
					}
					goto l161
				l162:
					position, tokenIndex = position161, tokenIndex161
					{
						position164 := position
						{
							position165, tokenIndex165 := position, tokenIndex
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l166
							}
							position++
							goto l165
						l166:
							position, tokenIndex = position165, tokenIndex165
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l163
							}
							position++
						}
					l165:
						add(rulePegText, position164)
					}
					{
						add(ruleAction30, position)
					}
					goto l161
				l163:
					position, tokenIndex = position161, tokenIndex161
					{
						position168, tokenIndex168 := position, tokenIndex
						if buffer[position] != rune('\\') {
							goto l168
						}
						position++
						goto l159
					l168:
						position, tokenIndex = position168, tokenIndex168
					}
					{
						position169 := position
						if !matchDot() {
							goto l159
						}
						add(rulePegText, position169)
					}
					{
						add(ruleAction31, position)
					}
				}
			l161:
				add(ruleDoubleChar, position160)
			}
			return true
		l159:
			position, tokenIndex = position159, tokenIndex159
			return false
		},
		/* 19 Escape <- <(('\\' ('a' / 'A') Action32) / ('\\' ('b' / 'B') Action33) / ('\\' ('e' / 'E') Action34) / ('\\' ('f' / 'F') Action35) / ('\\' ('n' / 'N') Action36) / ('\\' ('r' / 'R') Action37) / ('\\' ('t' / 'T') Action38) / ('\\' ('v' / 'V') Action39) / ('\\' '\'' Action40) / ('\\' '"' Action41) / ('\\' '[' Action42) / ('\\' ']' Action43) / ('\\' '-' Action44) / ('\\' ('0' ('x' / 'X')) <((&('A' | 'B' | 'C' | 'D' | 'E' | 'F') [A-F]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f') [a-f]) | (&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]))+> Action45) / ('\\' <([0-3] [0-7] [0-7])> Action46) / ('\\' <([0-7] [0-7]?)> Action47) / ('\\' '\\' Action48))> */
		func() bool {
			position171, tokenIndex171 := position, tokenIndex
			{
				position172 := position
				{
					position173, tokenIndex173 := position, tokenIndex
					if buffer[position] != rune('\\') {
						goto l174
					}
					position++
					{
						position175, tokenIndex175 := position, tokenIndex
						if buffer[position] != rune('a') {
							goto l176
						}
						position++
						goto l175
					l176:
						position, tokenIndex = position175, tokenIndex175
						if buffer[position] != rune('A') {
							goto l174
						}
						position++
					}
				l175:
					{
						add(ruleAction32, position)
					}
					goto l173
				l174:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l178
					}
					position++
					{
						position179, tokenIndex179 := position, tokenIndex
						if buffer[position] != rune('b') {
							goto l180
						}
						position++
						goto l179
					l180:
						position, tokenIndex = position179, tokenIndex179
						if buffer[position] != rune('B') {
							goto l178
						}
						position++
					}
				l179:
					{
						add(ruleAction33, position)
					}
					goto l173
				l178:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l182
					}
					position++
					{
						position183, tokenIndex183 := position, tokenIndex
						if buffer[position] != rune('e') {
							goto l184
						}
						position++
						goto l183
					l184:
						position, tokenIndex = position183, tokenIndex183
						if buffer[position] != rune('E') {
							goto l182
						}
						position++
					}
				l183:
					{
						add(ruleAction34, position)
					}
					goto l173
				l182:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l186
					}
					position++
					{
						position187, tokenIndex187 := position, tokenIndex
						if buffer[position] != rune('f') {
							goto l188
						}
						position++
						goto l187
					l188:
						position, tokenIndex = position187, tokenIndex187
						if buffer[position] != rune('F') {
							goto l186
						}
						position++
					}
				l187:
					{
						add(ruleAction35, position)
					}
					goto l173
				l186:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l190
					}
					position++
					{
						position191, tokenIndex191 := position, tokenIndex
						if buffer[position] != rune('n') {
							goto l192
						}
						position++
						goto l191
					l192:
						position, tokenIndex = position191, tokenIndex191
						if buffer[position] != rune('N') {
							goto l190
						}
						position++
					}
				l191:
					{
						add(ruleAction36, position)
					}
					goto l173
				l190:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l194
					}
					position++
					{
						position195, tokenIndex195 := position, tokenIndex
						if buffer[position] != rune('r') {
							goto l196
						}
						position++
						goto l195
					l196:
						position, tokenIndex = position195, tokenIndex195
						if buffer[position] != rune('R') {
							goto l194
						}
						position++
					}
				l195:
					{
						add(ruleAction37, position)
					}
					goto l173
				l194:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l198
					}
					position++
					{
						position199, tokenIndex199 := position, tokenIndex
						if buffer[position] != rune('t') {
							goto l200
						}
						position++
						goto l199
					l200:
						position, tokenIndex = position199, tokenIndex199
						if buffer[position] != rune('T') {
							goto l198
						}
						position++
					}
				l199:
					{
						add(ruleAction38, position)
					}
					goto l173
				l198:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l202
					}
					position++
					{
						position203, tokenIndex203 := position, tokenIndex
						if buffer[position] != rune('v') {
							goto l204
						}
						position++
						goto l203
					l204:
						position, tokenIndex = position203, tokenIndex203
						if buffer[position] != rune('V') {
							goto l202
						}
						position++
					}
				l203:
					{
						add(ruleAction39, position)
					}
					goto l173
				l202:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l206
					}
					position++
					if buffer[position] != rune('\'') {
						goto l206
					}
					position++
					{
						add(ruleAction40, position)
					}
					goto l173
				l206:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l208
					}
					position++
					if buffer[position] != rune('"') {
						goto l208
					}
					position++
					{
						add(ruleAction41, position)
					}
					goto l173
				l208:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l210
					}
					position++
					if buffer[position] != rune('[') {
						goto l210
					}
					position++
					{
						add(ruleAction42, position)
					}
					goto l173
				l210:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l212
					}
					position++
					if buffer[position] != rune(']') {
						goto l212
					}
					position++
					{
						add(ruleAction43, position)
					}
					goto l173
				l212:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l214
					}
					position++
					if buffer[position] != rune('-') {
						goto l214
					}
					position++
					{
						add(ruleAction44, position)
					}
					goto l173
				l214:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l216
					}
					position++
					if buffer[position] != rune('0') {
						goto l216
					}
					position++
					{
						position217, tokenIndex217 := position, tokenIndex
						if buffer[position] != rune('x') {
							goto l218
						}
						position++
						goto l217
					l218:
						position, tokenIndex = position217, tokenIndex217
						if buffer[position] != rune('X') {
							goto l216
						}
						position++
					}
				l217:
					{
						position219 := position
						{
							switch buffer[position] {
							case 'A', 'B', 'C', 'D', 'E', 'F':
								if c := buffer[position]; c < rune('A') || c > rune('F') {
									goto l216
								}
								position++
								break
							case 'a', 'b', 'c', 'd', 'e', 'f':
								if c := buffer[position]; c < rune('a') || c > rune('f') {
									goto l216
								}
								position++
								break
							default:
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l216
								}
								position++
								break
							}
						}

					l220:
						{
							position221, tokenIndex221 := position, tokenIndex
							{
								switch buffer[position] {
								case 'A', 'B', 'C', 'D', 'E', 'F':
									if c := buffer[position]; c < rune('A') || c > rune('F') {
										goto l221
									}
									position++
									break
								case 'a', 'b', 'c', 'd', 'e', 'f':
									if c := buffer[position]; c < rune('a') || c > rune('f') {
										goto l221
									}
									position++
									break
								default:
									if c := buffer[position]; c < rune('0') || c > rune('9') {
										goto l221
									}
									position++
									break
								}
							}

							goto l220
						l221:
							position, tokenIndex = position221, tokenIndex221
						}
						add(rulePegText, position219)
					}
					{
						add(ruleAction45, position)
					}
					goto l173
				l216:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l225
					}
					position++
					{
						position226 := position
						if c := buffer[position]; c < rune('0') || c > rune('3') {
							goto l225
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l225
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l225
						}
						position++
						add(rulePegText, position226)
					}
					{
						add(ruleAction46, position)
					}
					goto l173
				l225:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l228
					}
					position++
					{
						position229 := position
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l228
						}
						position++
						{
							position230, tokenIndex230 := position, tokenIndex
							if c := buffer[position]; c < rune('0') || c > rune('7') {
								goto l230
							}
							position++
							goto l231
						l230:
							position, tokenIndex = position230, tokenIndex230
						}
					l231:
						add(rulePegText, position229)
					}
					{
						add(ruleAction47, position)
					}
					goto l173
				l228:
					position, tokenIndex = position173, tokenIndex173
					if buffer[position] != rune('\\') {
						goto l171
					}
					position++
					if buffer[position] != rune('\\') {
						goto l171
					}
					position++
					{
						add(ruleAction48, position)
					}
				}
			l173:
				add(ruleEscape, position172)
			}
			return true
		l171:
			position, tokenIndex = position171, tokenIndex171
			return false
		},
		/* 20 LeftArrow <- <((('<' '-') / '←') Spacing)> */
		func() bool {
			position234, tokenIndex234 := position, tokenIndex
			{
				position235 := position
				{
					position236, tokenIndex236 := position, tokenIndex
					if buffer[position] != rune('<') {
						goto l237
					}
					position++
					if buffer[position] != rune('-') {
						goto l237
					}
					position++
					goto l236
				l237:
					position, tokenIndex = position236, tokenIndex236
					if buffer[position] != rune('←') {
						goto l234
					}
					position++
				}
			l236:
				if !_rules[ruleSpacing]() {
					goto l234
				}
				add(ruleLeftArrow, position235)
			}
			return true
		l234:
			position, tokenIndex = position234, tokenIndex234
			return false
		},
		/* 21 Slash <- <('/' Spacing)> */
		func() bool {
			position238, tokenIndex238 := position, tokenIndex
			{
				position239 := position
				if buffer[position] != rune('/') {
					goto l238
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l238
				}
				add(ruleSlash, position239)
			}
			return true
		l238:
			position, tokenIndex = position238, tokenIndex238
			return false
		},
		/* 22 And <- <('&' Spacing)> */
		func() bool {
			position240, tokenIndex240 := position, tokenIndex
			{
				position241 := position
				if buffer[position] != rune('&') {
					goto l240
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l240
				}
				add(ruleAnd, position241)
			}
			return true
		l240:
			position, tokenIndex = position240, tokenIndex240
			return false
		},
		/* 23 Not <- <('!' Spacing)> */
		func() bool {
			position242, tokenIndex242 := position, tokenIndex
			{
				position243 := position
				if buffer[position] != rune('!') {
					goto l242
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l242
				}
				add(ruleNot, position243)
			}
			return true
		l242:
			position, tokenIndex = position242, tokenIndex242
			return false
		},
		/* 24 Question <- <('?' Spacing)> */
		nil,
		/* 25 Star <- <('*' Spacing)> */
		nil,
		/* 26 Plus <- <('+' Spacing)> */
		nil,
		/* 27 Open <- <('(' Spacing)> */
		nil,
		/* 28 Close <- <(')' Spacing)> */
		nil,
		/* 29 Dot <- <('.' Spacing)> */
		nil,
		/* 30 SpaceComment <- <(Space / Comment)> */
		func() bool {
			position250, tokenIndex250 := position, tokenIndex
			{
				position251 := position
				{
					position252, tokenIndex252 := position, tokenIndex
					{
						position254 := position
						{
							switch buffer[position] {
							case '\t':
								if buffer[position] != rune('\t') {
									goto l253
								}
								position++
								break
							case ' ':
								if buffer[position] != rune(' ') {
									goto l253
								}
								position++
								break
							default:
								if !_rules[ruleEndOfLine]() {
									goto l253
								}
								break
							}
						}

						add(ruleSpace, position254)
					}
					goto l252
				l253:
					position, tokenIndex = position252, tokenIndex252
					{
						position256 := position
						if buffer[position] != rune('#') {
							goto l250
						}
						position++
					l257:
						{
							position258, tokenIndex258 := position, tokenIndex
							{
								position259, tokenIndex259 := position, tokenIndex
								if !_rules[ruleEndOfLine]() {
									goto l259
								}
								goto l258
							l259:
								position, tokenIndex = position259, tokenIndex259
							}
							if !matchDot() {
								goto l258
							}
							goto l257
						l258:
							position, tokenIndex = position258, tokenIndex258
						}
						if !_rules[ruleEndOfLine]() {
							goto l250
						}
						add(ruleComment, position256)
					}
				}
			l252:
				add(ruleSpaceComment, position251)
			}
			return true
		l250:
			position, tokenIndex = position250, tokenIndex250
			return false
		},
		/* 31 Spacing <- <SpaceComment*> */
		func() bool {
			{
				position261 := position
			l262:
				{
					position263, tokenIndex263 := position, tokenIndex
					if !_rules[ruleSpaceComment]() {
						goto l263
					}
					goto l262
				l263:
					position, tokenIndex = position263, tokenIndex263
				}
				add(ruleSpacing, position261)
			}
			return true
		},
		/* 32 MustSpacing <- <SpaceComment+> */
		func() bool {
			position264, tokenIndex264 := position, tokenIndex
			{
				position265 := position
				if !_rules[ruleSpaceComment]() {
					goto l264
				}
			l266:
				{
					position267, tokenIndex267 := position, tokenIndex
					if !_rules[ruleSpaceComment]() {
						goto l267
					}
					goto l266
				l267:
					position, tokenIndex = position267, tokenIndex267
				}
				add(ruleMustSpacing, position265)
			}
			return true
		l264:
			position, tokenIndex = position264, tokenIndex264
			return false
		},
		/* 33 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		nil,
		/* 34 Space <- <((&('\t') '\t') | (&(' ') ' ') | (&('\n' | '\r') EndOfLine))> */
		nil,
		/* 35 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position270, tokenIndex270 := position, tokenIndex
			{
				position271 := position
				{
					position272, tokenIndex272 := position, tokenIndex
					if buffer[position] != rune('\r') {
						goto l273
					}
					position++
					if buffer[position] != rune('\n') {
						goto l273
					}
					position++
					goto l272
				l273:
					position, tokenIndex = position272, tokenIndex272
					if buffer[position] != rune('\n') {
						goto l274
					}
					position++
					goto l272
				l274:
					position, tokenIndex = position272, tokenIndex272
					if buffer[position] != rune('\r') {
						goto l270
					}
					position++
				}
			l272:
				add(ruleEndOfLine, position271)
			}
			return true
		l270:
			position, tokenIndex = position270, tokenIndex270
			return false
		},
		/* 36 EndOfFile <- <!.> */
		nil,
		/* 37 Action <- <('{' <ActionBody*> '}' Spacing)> */
		func() bool {
			position276, tokenIndex276 := position, tokenIndex
			{
				position277 := position
				if buffer[position] != rune('{') {
					goto l276
				}
				position++
				{
					position278 := position
				l279:
					{
						position280, tokenIndex280 := position, tokenIndex
						if !_rules[ruleActionBody]() {
							goto l280
						}
						goto l279
					l280:
						position, tokenIndex = position280, tokenIndex280
					}
					add(rulePegText, position278)
				}
				if buffer[position] != rune('}') {
					goto l276
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l276
				}
				add(ruleAction, position277)
			}
			return true
		l276:
			position, tokenIndex = position276, tokenIndex276
			return false
		},
		/* 38 ActionBody <- <((!('{' / '}') .) / ('{' ActionBody* '}'))> */
		func() bool {
			position281, tokenIndex281 := position, tokenIndex
			{
				position282 := position
				{
					position283, tokenIndex283 := position, tokenIndex
					{
						position285, tokenIndex285 := position, tokenIndex
						{
							position286, tokenIndex286 := position, tokenIndex
							if buffer[position] != rune('{') {
								goto l287
							}
							position++
							goto l286
						l287:
							position, tokenIndex = position286, tokenIndex286
							if buffer[position] != rune('}') {
								goto l285
							}
							position++
						}
					l286:
						goto l284
					l285:
						position, tokenIndex = position285, tokenIndex285
					}
					if !matchDot() {
						goto l284
					}
					goto l283
				l284:
					position, tokenIndex = position283, tokenIndex283
					if buffer[position] != rune('{') {
						goto l281
					}
					position++
				l288:
					{
						position289, tokenIndex289 := position, tokenIndex
						if !_rules[ruleActionBody]() {
							goto l289
						}
						goto l288
					l289:
						position, tokenIndex = position289, tokenIndex289
					}
					if buffer[position] != rune('}') {
						goto l281
					}
					position++
				}
			l283:
				add(ruleActionBody, position282)
			}
			return true
		l281:
			position, tokenIndex = position281, tokenIndex281
			return false
		},
		/* 39 Begin <- <('<' Spacing)> */
		nil,
		/* 40 End <- <('>' Spacing)> */
		nil,
		/* 42 Action0 <- <{ p.AddPackage(text) }> */
		nil,
		/* 43 Action1 <- <{ p.AddPeg(text) }> */
		nil,
		/* 44 Action2 <- <{ p.AddState(text) }> */
		nil,
		nil,
		/* 46 Action3 <- <{ p.AddImport(text) }> */
		nil,
		/* 47 Action4 <- <{ p.AddRule(text) }> */
		nil,
		/* 48 Action5 <- <{ p.AddExpression() }> */
		nil,
		/* 49 Action6 <- <{ p.AddAlternate() }> */
		nil,
		/* 50 Action7 <- <{ p.AddNil(); p.AddAlternate() }> */
		nil,
		/* 51 Action8 <- <{ p.AddNil() }> */
		nil,
		/* 52 Action9 <- <{ p.AddSequence() }> */
		nil,
		/* 53 Action10 <- <{ p.AddPredicate(text) }> */
		nil,
		/* 54 Action11 <- <{ p.AddStateChange(text) }> */
		nil,
		/* 55 Action12 <- <{ p.AddPeekFor() }> */
		nil,
		/* 56 Action13 <- <{ p.AddPeekNot() }> */
		nil,
		/* 57 Action14 <- <{ p.AddQuery() }> */
		nil,
		/* 58 Action15 <- <{ p.AddStar() }> */
		nil,
		/* 59 Action16 <- <{ p.AddPlus() }> */
		nil,
		/* 60 Action17 <- <{ p.AddName(text) }> */
		nil,
		/* 61 Action18 <- <{ p.AddDot() }> */
		nil,
		/* 62 Action19 <- <{ p.AddAction(text) }> */
		nil,
		/* 63 Action20 <- <{ p.AddPush() }> */
		nil,
		/* 64 Action21 <- <{ p.AddSequence() }> */
		nil,
		/* 65 Action22 <- <{ p.AddSequence() }> */
		nil,
		/* 66 Action23 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 67 Action24 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		nil,
		/* 68 Action25 <- <{ p.AddAlternate() }> */
		nil,
		/* 69 Action26 <- <{ p.AddAlternate() }> */
		nil,
		/* 70 Action27 <- <{ p.AddRange() }> */
		nil,
		/* 71 Action28 <- <{ p.AddDoubleRange() }> */
		nil,
		/* 72 Action29 <- <{ p.AddCharacter(text) }> */
		nil,
		/* 73 Action30 <- <{ p.AddDoubleCharacter(text) }> */
		nil,
		/* 74 Action31 <- <{ p.AddCharacter(text) }> */
		nil,
		/* 75 Action32 <- <{ p.AddCharacter("\a") }> */
		nil,
		/* 76 Action33 <- <{ p.AddCharacter("\b") }> */
		nil,
		/* 77 Action34 <- <{ p.AddCharacter("\x1B") }> */
		nil,
		/* 78 Action35 <- <{ p.AddCharacter("\f") }> */
		nil,
		/* 79 Action36 <- <{ p.AddCharacter("\n") }> */
		nil,
		/* 80 Action37 <- <{ p.AddCharacter("\r") }> */
		nil,
		/* 81 Action38 <- <{ p.AddCharacter("\t") }> */
		nil,
		/* 82 Action39 <- <{ p.AddCharacter("\v") }> */
		nil,
		/* 83 Action40 <- <{ p.AddCharacter("'") }> */
		nil,
		/* 84 Action41 <- <{ p.AddCharacter("\"") }> */
		nil,
		/* 85 Action42 <- <{ p.AddCharacter("[") }> */
		nil,
		/* 86 Action43 <- <{ p.AddCharacter("]") }> */
		nil,
		/* 87 Action44 <- <{ p.AddCharacter("-") }> */
		nil,
		/* 88 Action45 <- <{ p.AddHexaCharacter(text) }> */
		nil,
		/* 89 Action46 <- <{ p.AddOctalCharacter(text) }> */
		nil,
		/* 90 Action47 <- <{ p.AddOctalCharacter(text) }> */
		nil,
		/* 91 Action48 <- <{ p.AddCharacter("\\") }> */
		nil,
	}
	p.rules = _rules
}
