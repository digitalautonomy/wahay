package hosting

import (
	"fmt"
	"log"

	"github.com/digitalautonomy/grumble/server"
)

func TestHosting() {
	s, err := server.NewServer(1)
	if err != nil {
		log.Fatalf("Couldn't start server: %s", err.Error())
	}
	fmt.Printf("NewServer: %v\n", s)
}
