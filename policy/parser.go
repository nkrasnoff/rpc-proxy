package policy

import (
	"fmt"
	"regexp"
	"strings"
)

type item struct {
	Ityp itemType
	val  string
}

// itemType represents the type of items returned
type itemType int

const (
	itemError itemType = iota
	itemRule
	itemDirSub
	itemAll
	itemDest
	itemInter
	itemMember
	itemDomUUID
	itemDomID
	itemDomType
	itemSender
	itemStubdom
	itemIfBoolTrue
	itemIfBoolFalse
	itemEOL
)

type stateFn func(*lexer) stateFn

type lexer struct {
	errors string
	input  []string
	pos    int
	items  chan item
}

func lex(errors string, input []string) *lexer {

	l := &lexer{
		errors: errors,
		input:  input,
		items:  make(chan item),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	for state := lexRule; state != nil; {
		state = state(l)
	}
	close(l.items)
	fmt.Println("The thread is done!")

}

// Helper function used to emit items from parser
func (l *lexer) emit(i itemType, optional ...string) {
	switch i {
	case itemRule:
		l.emitRule()
	case itemDirSub:
		l.emitDirSub(optional[0])
	case itemAll:
		l.emitAll()
	case itemIfBoolTrue, itemIfBoolFalse:
		l.emitSpecifier(i)
		l.pos++
	default:
		l.emitSpecifier(i)
	}
}

// Emit item specifying "allow" or "deny"
func (l *lexer) emitRule() {
	l.items <- item{itemRule, l.input[l.pos]}
	l.pos++
}

// Emit item specifying direction and subject
func (l *lexer) emitDirSub(dirSub string) {
	l.items <- item{itemDirSub, dirSub}
	l.pos++
}

// Emit item specifying to match all specifiers
func (l *lexer) emitAll() {
	l.items <- item{itemAll, ""}
}

// Emit item specifying a single specifier
func (l *lexer) emitSpecifier(i itemType) {
	l.items <- item{i, l.input[l.pos]}
	l.pos++
}

func lexTheRestPlaceholder(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return nil
	}

	i := item{itemMember, l.input[l.pos]}
	l.items <- i
	l.pos++
	return lexTheRestPlaceholder

}

func lexRule(l *lexer) stateFn {
	if l.input[l.pos] == "allow" || l.input[l.pos] == "deny" {
		l.emit(itemRule)
		return lexDirSub
	}

	fmt.Println("Not a allow or deny! Oh No!")
	return nil

}

func lexDirSub(l *lexer) stateFn {
	var dir Direction
	var sub Subject

	if l.pos >= len(l.input) {
		return nil
	}

	subject := l.input[l.pos]

	if strings.Index(subject, "inc-") == 0 {
		dir = Incoming
	} else if strings.Index(subject, "out-") == 0 {
		dir = Outgoing
	} else {
		return lexSpecAll
	}

	subSlice := strings.Split(subject, "-")

	if len(subSlice) != 2 && len(subSlice) != 3 {
		fmt.Println("Invalid subject")
		return nil
	}

	switch fst := subSlice[1]; fst {
	case "signal":
		if len(subSlice) != 2 {
			fmt.Println("Invalid subject")
			return nil
		}
		sub = Signal
	case "error":
		if len(subSlice) != 2 {
			fmt.Println("Invalid subject")
			return nil
		}
		sub = Error
	case "any":
		if len(subSlice) != 2 {
			fmt.Println("Invalid subject")
			return nil
		}
		sub = Any
	case "method":
		if len(subSlice) != 3 {
			fmt.Println("Invalid subject")
			return nil
		}
		switch snd := subSlice[2]; snd {
		case "call":
			sub = Call
		case "return":
			sub = Return
		default:
			fmt.Println("Invalid Subject")
			return nil
		}
	default:
		fmt.Println("Invalid Subject")
		return nil
	}

	//fmt.Println("Dir: ", dir, ", Sub: ", sub)
	l.emit(itemDirSub, fmt.Sprintf("%v%v", dir, sub))
	return lexSpecAll
}

func lexSpecAll(l *lexer) stateFn {

	if l.pos >= len(l.input) {
		return nil
	}
	if l.pos == len(l.input)-1 && l.input[l.pos] == "all" {
		//fmt.Println("Found \"all\": Match everything")
		l.emit(itemAll)
		return nil
	}

	return lexSpecifier
}

func lexSpecifier(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return nil
	}

	//fmt.Println(l.input[l.pos])

	switch spec := l.input[l.pos]; spec {
	case "destination":
		l.pos++
		return lexDest
	case "interface":
		l.pos++
		return lexInterface
	case "member":
		l.pos++
		return lexMember
	case "dom-uuid":
		l.pos++
		return lexDomUUID
	case "dom-id":
		l.pos++
		return lexDomID
	case "dom-type":
		l.pos++
		return lexDomType
	case "sender":
		l.pos++
		return lexSender
	case "stubdom":
		l.pos++
		return lexStubdom
	case "if-boolean":
		l.pos++
		return lexIfBool
	}

	fmt.Println("Oh my that's an invalid specifier")
	return nil
}

func lexDest(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if !validStr(l.input[l.pos]) {
		fmt.Println("Invalid Destination")
	}
	l.emit(itemDest)
	return lexSpecifier
}

func lexInterface(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if !validStr(l.input[l.pos]) {
		fmt.Println("Invalid Interface")
		return nil
	}
	l.emit(itemInter)

	return lexSpecifier
}

func lexMember(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if !validStr(l.input[l.pos]) {
		fmt.Println("Invalid Member")
		return nil
	}
	l.emit(itemMember)
	return lexSpecifier
}

func lexDomUUID(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if !validStr(l.input[l.pos]) {
		fmt.Println("Invalid Dom-UUID")
		return nil
	}
	l.emit(itemDomUUID)
	return lexSpecifier
}

func lexDomID(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if !validInt(l.input[l.pos]) {
		fmt.Println("Invalid Dom-ID")
		return nil
	}
	l.emit(itemDomID)
	return lexSpecifier
}

func lexDomType(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if !validStr(l.input[l.pos]) {
		fmt.Println("Invalid Dom-Type")
		return nil
	}
	l.emit(itemDomType)
	return lexSpecifier
}

func lexSender(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if !validStr(l.input[l.pos]) {
		fmt.Println("Invalid Sender")
		return nil
	}
	l.emit(itemSender)
	return lexSpecifier
}

func lexStubdom(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if l.input[l.pos] != "true" && l.input[l.pos] != "false" {
		fmt.Println("Stubdom needs \"true\" or \"false\"")
		return nil
	}
	l.emit(itemStubdom)
	return lexSpecifier
}

func lexIfBool(l *lexer) stateFn {
	if l.pos+1 >= len(l.input) {
		fmt.Println("Specifier not given enough conditions")
		return nil
	}
	if !validStr(l.input[l.pos]) {
		fmt.Println("Invalid If-Boolean condition")
		return nil
	}

	if l.input[l.pos+1] == "true" {
		l.emit(itemIfBoolTrue)
	} else if l.input[l.pos+1] == "false" {
		l.emit(itemIfBoolFalse)
	} else {
		fmt.Println("if-boolean needs \"true\" or \"false\"")
		return nil
	}
	return lexSpecifier
}

func validStr(s string) bool {

	b, _ := regexp.MatchString(`^[0-9a-zA-Z_\.\-]+$`, s)

	return b
}

func validInt(s string) bool {

	b, _ := regexp.MatchString(`^[0-9]+$`, s)

	return b
}
