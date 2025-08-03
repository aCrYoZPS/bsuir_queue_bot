package ioc

import (
	"fmt"
	
	"github.com/aCrYoZPS/bsuir_queue_bot/src/logging"
)

var currentId = 0

func getNextId() int {
	currentId += 1
	return currentId
}

var container = map[int]any{}
var isPending = map[int]bool{}

func provider[T any](factory func() T) func() T {
	providerId := getNextId()
	return func() T {
		if pending, ok := isPending[providerId]; ok && pending {
			logging.FatalLog(fmt.Sprintf("cirricular dependecy of id %d in a container", providerId))
		}
		if _, ok := container[providerId]; !ok {
			isPending[providerId] = true
			container[providerId] = factory()
			isPending[providerId] = false
		}
		return container[providerId].(T)
	}
}

func Reset() {
	container = map[int]any{}
	isPending = map[int]bool{}
}