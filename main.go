package main

import (
	"time"

	"github.com/realDragonium/Albedo/status"
)

func main() {
	status.StartSpam(10, 10*time.Millisecond)
	// status.SendStatus()
	// status.PrintServerStatus()
	// status.StatusSomething()

}
