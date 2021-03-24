package policy

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

//Rule for rpc-proxy Firewall
type Rule struct {
	RDirection   Direction
	RSubject     Subject
	RDestination string
	RInt         int
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

//ReadPolicy from config files and return rules
func ReadPolicy(path string) {
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
			fmt.Println("TODO: SKIP")
		} else {
			ruleSlc := strings.Fields(ruleStr)

			r := createRule(ruleSlc)
			fmt.Println("Rule: ", r)
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func createRule(ruleSlc []string) Rule {
	l := lex("Testing", ruleSlc)
	for aItem := range l.items {
		fmt.Println("Item: ", aItem)

	}
	//fmt.Println("The thread got closed!")

	var newRule Rule
	newRule.RDirection = Incoming
	newRule.RSubject = Error

	return newRule
}
