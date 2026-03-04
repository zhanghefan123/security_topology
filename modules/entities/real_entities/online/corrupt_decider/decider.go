package corrupt_decider

type DecideFunction func() int

type CorruptDecider struct {
	DecideFunction *DecideFunction
}

func (cd *CorruptDecider) ShouldCorrupt() bool {
	return (*cd.DecideFunction)() == 1
}
