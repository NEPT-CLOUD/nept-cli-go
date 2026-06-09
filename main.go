package main

import (
	"fmt"
	"os"

	"github.com/NEPT-CLOUD/nept-cli-go/cmd"
	"github.com/NEPT-CLOUD/nept-cli-go/internal/app/utils"
)

func main() {
	fmt.Println(utils.BackendUrl)
	os.Exit(cmd.Execute())

}
