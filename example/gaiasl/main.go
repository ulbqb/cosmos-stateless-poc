package main

import (
	"flag"
	"fmt"

	"github.com/ulbqb/cosmos-stateless-poc/example/gaiasl/exec"
)

// It is assumed that Hash of block to execute is correct.
func main() {
	var basedir string
	var trustHeight int
	var trustBlockHash string
	var rpcAddr string

	flag.StringVar(&basedir, "basedir", "/tmp/stateless", "Directory to cache oracle data.")
	flag.IntVar(&trustHeight, "height", 1, "Height of block to execute")
	flag.StringVar(&trustBlockHash, "hash", "", "Hash of block to execute")
	flag.StringVar(&rpcAddr, "rpc", "http://localhost", "RPC host.")
	flag.Parse()

	appHash, _, err := exec.Execute(basedir, trustHeight, trustBlockHash, rpcAddr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%X\n", appHash)
}
