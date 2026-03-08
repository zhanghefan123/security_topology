package decider

type DecideFunction func() int

type ActionDecider struct {
	DecideFunction *DecideFunction
}

func (cd *ActionDecider) ShouldTakeAction() bool {
	return (*cd.DecideFunction)() == 1
}
