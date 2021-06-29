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

// Return an error item and terminate parsing
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, fmt.Sprintf(format, args...)}
	return nil
}

func lexRule(l *lexer) stateFn {
	if l.input[l.pos] == "allow" || l.input[l.pos] == "deny" {
		l.emit(itemRule)
		return lexDirSub
	}

	return l.errorf("\"%v\" is not allow or deny", l.input[l.pos])

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
		return l.errorf("Invalid subject")
	}

	switch fst := subSlice[1]; fst {
	case "signal":
		if len(subSlice) != 2 {
			return l.errorf("Invalid subject")
		}
		sub = Signal
	case "error":
		if len(subSlice) != 2 {
			return l.errorf("Invalid subject")
		}
		sub = Error
	case "any":
		if len(subSlice) != 2 {
			return l.errorf("Invalid subject")
		}
		sub = Any
	case "method":
		if len(subSlice) != 3 {
			return l.errorf("Invalid subject")
		}
		switch snd := subSlice[2]; snd {
		case "call":
			sub = Call
		case "return":
			sub = Return
		default:
			return l.errorf("Invalid subject")
		}
	default:
		return l.errorf("Invalid subject")
	}

	l.emit(itemDirSub, fmt.Sprintf("%v%v", int(dir), int(sub)))
	return lexSpecAll
}

func lexSpecAll(l *lexer) stateFn {

	if l.pos >= len(l.input) {
		return nil
	}
	if l.pos == len(l.input)-1 && l.input[l.pos] == "all" {
		l.emit(itemAll)
		return nil
	}

	return lexSpecifier
}

func lexSpecifier(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return nil
	}

	var specifier stateFn
	switch spec := l.input[l.pos]; spec {
	case "destination":
		specifier = lexDest
	case "interface":
		specifier = lexInterface
	case "member":
		specifier = lexMember
	case "dom-uuid":
		specifier = lexDomUUID
	case "dom-id":
		specifier = lexDomID
	case "dom-type":
		specifier = lexDomType
	case "sender":
		specifier = lexSender
	case "stubdom":
		specifier = lexStubdom
	case "if-boolean":
		specifier = lexIfBool
	default:
		return l.errorf("Invalid specifier: %v", l.input[l.pos])
	}

	l.pos++
	return specifier
}

func lexDest(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return l.errorf("Specifier \"destination\" not given enough arguments")
	}
	if !validStr(l.input[l.pos]) {
		return l.errorf("Invalid Destination: %v", l.input[l.pos])
	}
	l.emit(itemDest)
	return lexSpecifier
}

func lexInterface(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return l.errorf("Specifier \"interface\" not given enough arguments")
	}
	if !validStr(l.input[l.pos]) {
		return l.errorf("Invalid Interface: %v", l.input[l.pos])
	}
	l.emit(itemInter)

	return lexSpecifier
}

func lexMember(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return l.errorf("Specifier \"member\" not given enough arguments")
	}
	if !validStr(l.input[l.pos]) {
		return l.errorf("Invalid Member: %v", l.input[l.pos])
	}
	l.emit(itemMember)
	return lexSpecifier
}

func lexDomUUID(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return l.errorf("Specifier \"dom-uuid\" not given enough arguments")
	}
	if !validStr(l.input[l.pos]) {
		return l.errorf("Invalid Dom-UUID: %v", l.input[l.pos])
	}
	l.emit(itemDomUUID)
	return lexSpecifier
}

func lexDomID(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return l.errorf("Specifier \"dom-id\" not given enough arguments")
	}
	if !validInt(l.input[l.pos]) {
		return l.errorf("Invalid Dom-ID: %v", l.input[l.pos])
	}
	l.emit(itemDomID)
	return lexSpecifier
}

func lexDomType(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return l.errorf("Specifier \"dom-type\" not given enough arguments")
	}
	if !validStr(l.input[l.pos]) {
		return l.errorf("Invalid Dom-Type: %v", l.input[l.pos])
	}
	l.emit(itemDomType)
	return lexSpecifier
}

func lexSender(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return l.errorf("Specifier \"sender\" not given enough arguments")
	}
	if !validStr(l.input[l.pos]) {
		return l.errorf("Invalid Sender: %v", l.input[l.pos])
	}
	l.emit(itemSender)
	return lexSpecifier
}

func lexStubdom(l *lexer) stateFn {
	if l.pos >= len(l.input) {
		return l.errorf("Specifier \"stubdom\" not given enough arguments")
	}
	if l.input[l.pos] != "true" && l.input[l.pos] != "false" {
		return l.errorf("Stubdom given \"%v\" instead of boolean", l.input[l.pos])
	}
	l.emit(itemStubdom)
	return lexSpecifier
}

func lexIfBool(l *lexer) stateFn {
	if l.pos+1 >= len(l.input) {
		return l.errorf("Specifier \"if-boolean\" not given enough arguments")
	}
	if !validStr(l.input[l.pos]) {
		return l.errorf("Invalid If-Boolean: %v", l.input[l.pos])
	}

	if l.input[l.pos+1] == "true" {
		l.emit(itemIfBoolTrue)
	} else if l.input[l.pos+1] == "false" {
		l.emit(itemIfBoolFalse)
	} else {
		return l.errorf("Second argument of if-boolean must be boolean")
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
