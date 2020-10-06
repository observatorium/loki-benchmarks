package docker

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/client"
)

func WaitForReadyDeployment(c client.Client, cid string, timeout time.Duration) error {
	rc := make(chan struct{}, 1)
	go func() {
		for {
			s, _ := c.ContainerInspect(context.Background(), cid)
			if s.State.Running {
				rc <- struct{}{}
			}
		}
	}()

	select {
	case <-rc:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout waiting for container running")
	}
}
