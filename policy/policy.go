package policy

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

//Rule for rpc-proxy Firewall
type Rule struct {
	Allow       bool
	Direction   Direction
	Subject     Subject
	Destination string
	Interface   string
	Member      string
	DomUUID     string
	DomID       string
	DomType     string
	Sender      string
	SpecStubdom bool
	Stubdom     bool
	IfBool      IfBool
	Int         int
}

type IfBool struct {
	Condition  bool
	Identifier string
}

//Direction is whether a message is sent or recieved by rpc-proxy
type Direction int

//Directions
const (
	Incoming Direction = iota
	Outgoing
)

//Subject is the type of dbus message
type Subject int

//Subjects
const (
	Any Subject = iota
	Call
	Signal
	Return
	Error
)

func ReadPolicy(path string) []Rule {
	var rules []Rule

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ruleStr := scanner.Text()
		fmt.Println("Read line: ", ruleStr)
		if ruleStr == "" || ruleStr[0] == '#' {
			continue
		}

		r := createRule(strings.Fields(ruleStr))
		rules = append(rules, r)
		fmt.Println("Rule: ", r)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return rules
}

func createRule(ruleSlc []string) Rule {
	l := lex("Testing", ruleSlc)
	var newRule Rule
	for aItem := range l.items {
		switch aItem.Ityp {
		case itemRule:
			newRule.Allow = aItem.val == "allow"
		case itemDirSub:
			//As item values are passed as a string Direction and Subject are passed as
			//a string where the first character represents the Direcion and the second
			//represents the Subject
			dirsub, _ := strconv.Atoi(aItem.val)
			newRule.Direction = Direction(dirsub / 10)
			newRule.Subject = Subject(dirsub % 10)
		case itemDest:
			newRule.Destination = aItem.val
		case itemInter:
			newRule.Interface = aItem.val
		case itemMember:
			newRule.Member = aItem.val
		case itemDomUUID:
			newRule.DomUUID = aItem.val
		case itemDomID:
			newRule.DomID = aItem.val
		case itemDomType:
			newRule.DomType = aItem.val
		case itemSender:
			newRule.Sender = aItem.val
		case itemStubdom:
			newRule.SpecStubdom = true
			newRule.Stubdom = aItem.val == "true"
		case itemIfBoolTrue:
			newRule.IfBool.Condition = true
			newRule.IfBool.Identifier = aItem.val
		case itemIfBoolFalse:
			newRule.IfBool.Condition = false
			newRule.IfBool.Identifier = aItem.val
		case itemError:
			log.Fatal(aItem.val)
			os.Exit(1)

		}

	}

	return newRule
}
