package lib

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
)

type EventListener interface {
	Build(ctx context.Context)
}

type eventListener struct {
	Contract   *bind.BoundContract
	Events     []string
	signerOpts *bind.TransactOpts
}

func (e *eventListener) Build(ctx context.Context) {
	WatchOpts := bind.WatchOpts{
		Start:   nil,
		Context: ctx,
	}
	for _, event := range e.Events {
		logs, sub, err := e.Contract.WatchLogs(&WatchOpts, event)
		if err != nil {
			err = fmt.Errorf("error watching logs: %v", err)
			panic(err)
		}

		go func(event string, logs chan types.Log, sub ethereum.Subscription) {
			defer sub.Unsubscribe()
			for {
				select {
				case err := <-sub.Err():
					if err != nil {
						err = fmt.Errorf("subscription error: %v", err)
						panic(err)
					}
					return
				case vLog := <-logs:
					fmt.Printf("Received log for event %s: %v\n", event, vLog)
				case <-ctx.Done():
					log.Println("context cancelled, stopping listener")
					return
				}
			}
		}(event, logs, sub)
	}
}
