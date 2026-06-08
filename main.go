package main

import (
	// "fmt"

	// "github.com/NEPT-CLOUD/nept-cli-go/internal/app/utls"
)

import (
	"os"

	"github.com/NEPT-CLOUD/nept-cli-go/cmd"
)

func main() {

	// fmt.Println(utls.Backend)

	os.Exit(cmd.Execute())

}
