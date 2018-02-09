package main

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/cmd"
	_ "github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
)


func main() {
	cmd.Execute()
}
