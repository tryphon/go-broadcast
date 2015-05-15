package command

import (
	"fmt"
	"os"
	"projects.tryphon.eu/go-broadcast/broadcast"
)

type Base struct {
}

func (command *Base) checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		broadcast.Log.Printf("Fatal error : %s", err.Error())
		os.Exit(1)
	}
}
