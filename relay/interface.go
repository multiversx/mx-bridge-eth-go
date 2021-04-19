package relay

import "context"

type Startable interface {
	Start(context.Context) error

	Stop() error
}
