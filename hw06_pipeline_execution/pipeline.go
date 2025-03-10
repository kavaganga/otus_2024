package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	output := in
	if done != nil {
		output = ControlPipeline(in, done)
	}

	for _, stage := range stages {
		input := output
		output = stage(input)
	}
	return output
}

func ControlPipeline(done In, in In) Out {
	output := make(Bi)

	go func(done In, in In, out Bi) {
		for {
			select {
			case <-done:
				close(out)
				return
			default:
				val, ok := <-in
				if !ok {
					close(out)
					return
				}
				out <- val
			}
		}
	}(done, in, output)
	return output
}
