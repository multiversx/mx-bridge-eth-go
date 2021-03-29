package bridge

import "context"

type Startable interface {
	Start(context.Context)

	Monitor()

	Stop()
}
