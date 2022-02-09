package shipwright

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"pkg.grafana.com/shipwright/v1/docker"
	"pkg.grafana.com/shipwright/v1/fs"
	"pkg.grafana.com/shipwright/v1/git"
	"pkg.grafana.com/shipwright/v1/golang"
	makefile "pkg.grafana.com/shipwright/v1/make"
	"pkg.grafana.com/shipwright/v1/plumbing"
	"pkg.grafana.com/shipwright/v1/plumbing/config"
	"pkg.grafana.com/shipwright/v1/plumbing/pipeline"
	"pkg.grafana.com/shipwright/v1/plumbing/plog"
	"pkg.grafana.com/shipwright/v1/yarn"
)

type Client interface {
	config.Configurer

	// Validate is ran internally before calling Run or Parallel and allows the client to effectively configure per-step requirements
	// For example, Drone steps MUST have an image so the Drone client returns an error in this function when the provided step does not have an image.
	// If the error encountered is not critical but should still be logged, then return a plumbing.ErrorSkipValidation.
	// The error is checked with `errors.Is` so the error can be wrapped with fmt.Errorf.
	Validate(pipeline.Step) error
	// Run allows users to define steps that are ran sequentially. For example, the second step will not run until the first step has completed.
	// This function blocks the goroutine until all of the steps have completed.
	Run(...pipeline.Step)

	// Parallel will run the listed steps at the same time.
	// This function blocks the goroutine until all of the steps have completed.
	Parallel(...pipeline.Step)

	// Go is the equivalent of `go func()`. This function will run a step asynchronously and continue on to the next.
	// Go(...pipeline.Step)

	Cache(pipeline.StepAction, pipeline.Cacher) pipeline.StepAction
	Input(...pipeline.Argument)
	Output(...pipeline.Output)

	// Done must be ran at the end of the pipeline.
	// This is typically what takes the defined pipeline steps, runs them in the order defined, and produces some kind of output.
	Done()
}

type Shipwright struct {
	Client
	Git    git.Client
	FS     fs.Client
	Golang golang.Client
	Make   makefile.Client
	Yarn   yarn.Client
	Docker docker.Client

	// n tracks the ID of a step so that the "shipwright -step=" argument will function independently of the client implementation
	// It ensures that the 11th step in a Drone generated pipeline is also the 11th step in a CLI pipeline
	n int

	Version string
}

func (s *Shipwright) initSteps(steps ...pipeline.Step) []pipeline.Step {
	for i, step := range steps {
		// Set a default image for steps that don't provide one.
		// Most pre-made steps like `yarn`, `node`, `go` steps should provide a separate default image with those utilities installed.
		if step.Image == "" {
			image := plumbing.DefaultImage(s.Version)
			steps[i] = step.WithImage(image)
		}

		// Set a serial / unique identifier for this step so that we can reference it using the '-step' argument consistently.
		steps[i].Serial = s.n
		s.n++
	}

	return steps
}

func formatError(step pipeline.Step, err error) string {
	name := step.Name
	if name == "" {
		name = fmt.Sprintf("unnamed-step-%d", step.Serial)
	}

	return fmt.Sprintf("[name: %s, id: %d] %s", name, step.Serial, err.Error())
}

func (s *Shipwright) validateSteps(steps ...pipeline.Step) {
	for _, v := range steps {
		err := s.Validate(v)
		if err == nil {
			continue
		}

		if errors.Is(err, plumbing.ErrorSkipValidation) {
			plog.Warnln(formatError(v, err))
			continue
		}

		plog.Fatalln(formatError(v, err))
		return
	}
}

func (s *Shipwright) Run(steps ...pipeline.Step) {
	initializedSteps := s.initSteps(steps...)
	s.validateSteps(steps...)

	s.Client.Run(initializedSteps...)
}

func (s *Shipwright) Parallel(steps ...pipeline.Step) {
	initializedSteps := s.initSteps(steps...)
	s.validateSteps(steps...)

	s.Client.Parallel(initializedSteps...)
}

// New creates a new Shipwright client which is used to create pipeline steps.
// This function will panic if the arguments in os.Args do not match what's expected.
// This function, and the type it returns, are only ran inside of a Shipwright pipeline, and so it is okay to treat this like it is the entrypoint of a command.
// Watching for signals, parsing command line arguments, and panics are all things that are OK in this function.
func New(name string, events ...pipeline.Event) Shipwright {
	args, err := plumbing.ParseArguments(os.Args[1:])
	if err != nil {
		plog.Fatalln("Error parsing arguments. Error:", err)
	}

	if args == nil {
		plog.Fatalln("Arguments list must not be nil")
		return Shipwright{}
	}

	sw := NewFromOpts(&pipeline.CommonOpts{
		Name:    name,
		Version: args.Version,
		Output:  os.Stdout,
		Args:    args,
		Log:     plog.New(args.LogLevel, os.Stderr),
	})

	// Ensure that no matter the behavior of the initializer, we still set the version on the shipwright object.
	sw.Version = args.Version

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		<-sigs
		fmt.Print("\033[?25h")
	}()
	return sw
}

func NewFromOpts(opts *pipeline.CommonOpts, events ...pipeline.Event) Shipwright {
	return NewClient(opts)
}
