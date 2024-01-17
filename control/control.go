package control

type Control interface {
	SetContext() error
}

type ControllerInstance func() Control

type Controller struct {
}

func (c *Controller) SetContext() error {
	return nil
}
