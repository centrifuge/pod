package transaction

import (
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

type Submitter interface {
	SubmitAndWatch(method interface{}, params ...interface{}) func(accountID identity.DID, jobID jobs.JobID, jobMan jobs.Manager, errOut chan<- error)
}
