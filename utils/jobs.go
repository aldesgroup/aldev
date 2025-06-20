package utils

import (
	"fmt"
	"sync"
	"time"
)

func RunJobs(aldevCtx CancelableContext, parallel bool) {
	if parallel {
		runJobsInParallel(aldevCtx, Config().Jobs)
	} else {
		runJobsInSequence(aldevCtx, Config().Jobs)
	}
}

func runJobsInParallel(aldevCtx CancelableContext, jobs []*JobConfig) {
	startTime := time.Now()
	wg := new(sync.WaitGroup)
	for _, job := range jobs {
		wg.Add(1)
		go runJob(aldevCtx, job, wg)
	}

	wg.Wait()
	Debug("All jobs in parallel have been run in %dms\n", time.Since(startTime).Milliseconds())
}

func runJobsInSequence(aldevCtx CancelableContext, jobs []*JobConfig) {
	startTime := time.Now()
	for _, job := range jobs {
		runJob(aldevCtx, job, nil)
		Debug("Job '%s' has been run in %dms", job.Description, time.Since(startTime).Milliseconds())
	}
	Debug("All jobs in sequence have been run in %dms\n", time.Since(startTime).Milliseconds())
}

func runJob(aldevCtx CancelableContext, job *JobConfig, wg *sync.WaitGroup) {
	defer wg.Done()
	defer Recover(aldevCtx, "Job '%s' has failed", job.Description)

	Debug("Running job '%s'", job.Description)

	for i, cmd := range job.Cmds {
		Run(fmt.Sprintf("Command %d", i+1), aldevCtx.NewChildContext().WithExecDir(cmd.From).WithAllowFailure(cmd.FailOK), false, "%s", cmd.Exec)
	}
}
