package schedulers

import (
	"github.com/hashicorp/hcl/v2"
)

type Scheduler interface {
	Run() hcl.Diagnostics
}
