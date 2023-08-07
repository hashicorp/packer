package opts

type SchedulerOptions struct {
	Only, Except []string

	// Build-specific options
	Debug, Force                        bool
	Color, TimestampUi, MachineReadable bool
	ParallelBuilds                      int64
	OnError                             string

	SkipDatasourcesExecution bool
}
