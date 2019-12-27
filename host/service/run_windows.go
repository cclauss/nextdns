package service

import (
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
)

type windowService struct {
	Runner
	lastErr error
}

func (s windowService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	if err := s.Start(s.log); err != nil {
		s.lastErr = err
		return true, 1
	}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

loop:
	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			changes <- svc.Status{State: svc.StopPending}
			if err := s.Stop(s.log); err != nil {
				s.lastErr = err
				return true, 2
			}
			break loop
		}
	}

	return false, 0
}

func runService(r Runner) error {
	runner := svc.Run
	if isDebug {
		runner = debug.Run
	}
	s := &windowService{r}
	err := runner(name, s)
	if s.lastErr != nil {
		return s.lastErr
	}
	return err
}
