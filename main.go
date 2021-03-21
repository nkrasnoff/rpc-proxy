/*
 Go implementation of RPC-Proxy for Openxt
*/

package main

import (
	"fmt"
	"rpc-proxy/policy"
)

func main() {
	fmt.Println("vim-go")
	policy.ReadPolicy("rpc-proxy1.rules")
	fmt.Println(policy.Rule{})
}
