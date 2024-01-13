/*
 * Bootstrap utility library. Abstracts shutdown handlers.
 *    TODO add setup/destroy bootstrap timeouts.
 */

package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type Setup func() error

type Daemon func(done context.CancelFunc)

type Shutdown func()

// run a Daemon process given start/run/stop functions.
func RunDaemon(doSetup Setup, doRun Daemon, doShutdown Shutdown) {
	ctx, done := context.WithCancel(context.Background())

	// setup is optional, but if provided it must succeed
	if doSetup != nil {
		if err := doSetup(); err != nil {
			os.Exit(1)
		}
	}

	// shutdown is optional, and will block exit
	if doShutdown != nil {
		defer doShutdown()
	}

	// signal handler is removed before shutdown
	term, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	// daemon signals on context when (or if) finished
	go doRun(done)

	// a termination signal, or daemon exited
	<-term.Done()
}
