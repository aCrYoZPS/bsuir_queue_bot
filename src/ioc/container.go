// FUCK IT WE BALL
package ioc

import (
	"fmt"
	"log/slog"
	"os"
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
			slog.Error(fmt.Sprintf("Cirricular dependecy of id %d in a container", providerId))
			os.Exit(-1)
		}
		if _, ok := container[providerId]; !ok {
			isPending[providerId] = true
			container[providerId] = factory()
			isPending[providerId] = false
		}
		return container[providerId].(T)
	}
}

func reset() {
	container = map[int]any{}
	isPending = map[int]bool{}
}