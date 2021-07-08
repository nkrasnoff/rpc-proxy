package policy

import (
	"bufio"
	"fmt"
	"github.com/godbus/dbus"
	"log"
	"os"
	"strconv"
	"strings"
)

//Rules contains global rules and per domain rules
type Rules struct {
	Global []Rule
	PerVM  map[string][]Rule
}

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

//IfBool contains an Identifier that must match a bool condition to be accepted
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

//ReadPolicy from config files and return rules
func ReadPolicy(path string) Rules {
	var rules Rules

	//Read in Global rules from config files
	rules.Global = readGlobalRules(path)
	rules.PerVM = readPerVMRules()

	return rules
}

func readGlobalRules(path string) []Rule {
	var globalRules []Rule

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
		globalRules = append(globalRules, r)
		fmt.Println("Rule: ", r)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return globalRules
}

func readPerVMRules() map[string][]Rule {
	perVM := make(map[string][]Rule)

	conn, err := dbus.SystemBus()
	if err != nil {
		fmt.Println("Error finding system bus: ", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("I connected to the bus")

	obj := conn.Object("com.citrix.xenclient.db", "/")
	call := obj.Call("com.citrix.xenclient.db.list", 0, "/vm/")
	if call.Err != nil {
		fmt.Println("Error calling db.list: ", call.Err)
	}
	var vms []string
	call.Store(&vms)

	for _, uuid := range vms {
		call = obj.Call("com.citrix.xenclient.db.list", 0, "/vm/"+uuid+"/rpc-firewall-rules")
		if call.Err != nil {
			fmt.Println("Error calling db.list: ", call.Err)
		}
		var ruleIds []string
		call.Store(&ruleIds)

		var vmRules []Rule
		for _, id := range ruleIds {
			//fmt.Println("/vm/" + uuid + "/rpc-firewall-rules/" + id)
			call = obj.Call("com.citrix.xenclient.db.read", 0, "/vm/"+uuid+"/rpc-firewall-rules/"+id)
			if call.Err != nil {
				fmt.Println("Error calling db.read: ", call.Err)
			}
			var ruleStr string
			call.Store(&ruleStr)
			//fmt.Println(ruleStr)

			r := createRule(strings.Fields(ruleStr))
			vmRules = append(vmRules, r)

		}
		perVM[uuid] = vmRules
	}

	return perVM
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
