package workerpool

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Worker interface {
	StartWorkerPool(func(interface{}) (ctrl.Result, error)) error
	SubmitJob(job interface{})
	SubmitJobAfter(job interface{}, submitAfter time.Duration)
}

type worker struct {
	// workersStarted is the flag to prevent starting duplicate set of workers
	workersStarted bool
	// workerFunc is the function that will be invoked with the job by the worker routine
	workerFunc func(interface{}) (ctrl.Result, error)
	// maxRetries is the number of times to retry item in case of failure
	maxRetriesOnErr int
	// maxWorkerCount represents the maximum number of workers that will be started
	maxWorkerCount int
	// ctx is the background context to close the chanel on termination signal
	ctx context.Context
	// Log is the structured logger set to log with resource name
	Log logr.Logger

	// queue is the k8s rate limiting queue to store the submitted jobs
	queue workqueue.TypedRateLimitingInterface[interface{}] // interface{} allows any type of job to be submitted
}

func NewDefaultWorkerPool(workerCount int, maxRequeue int,
	logger logr.Logger, ctx context.Context) Worker {

	return &worker{

		maxRetriesOnErr: maxRequeue,
		maxWorkerCount:  workerCount,
		Log:             logger,
		queue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[interface{}](), "default-worker-queue"),
		ctx:             ctx,
	}
}

/*
Worker related function to process from queue
*/
func (w *worker) SetWorkerFunc(workerFunc func(interface{}) (ctrl.Result, error)) {
	w.workerFunc = workerFunc
}

// runWorker runs a worker that listens on new item on the worker queue
func (w *worker) runWorker() {
	for w.processNextItem() {
	}
}

// StartWorkerPool starts the worker pool that starts the worker routines that concurrently listen on the channel
func (w *worker) StartWorkerPool(workerFunc func(interface{}) (ctrl.Result, error)) error {

	return nil
}

// processNextItem returns false if the queue is shut down, otherwise processes the job and returns true
func (w *worker) processNextItem() (cont bool) {
	job, quit := w.queue.Get()
	if quit {
		return
	}
	defer w.queue.Done(job)
	log := w.Log.WithValues("job", job)

	cont = true
	// Handles any error with completing job - if job fails, re-queue it again
	if result, err := w.workerFunc(job); err != nil {
		if w.queue.NumRequeues(job) >= w.maxRetriesOnErr {
			log.Error(err, "exceeded maximum retries", "max retries", w.maxRetriesOnErr)
			w.queue.Forget(job)
			return
		}
		log.Error(err, "re-queuing job", "retry count", w.queue.NumRequeues(job))
		w.queue.AddRateLimited(job)
		return
	} else if result.RequeueAfter > 7*time.Second {
		log.V(1).Info("timed retry", "retry after", result.RequeueAfter)
		w.queue.AddAfter(job, result.RequeueAfter)
		return
	}

	log.V(1).Info("completed job successfully")

	w.queue.Forget(job)
	return
}

/*
	Submitting jobs/pushing to the queue
*/
// SubmitJob adds the job to the rate limited queue
func (w *worker) SubmitJob(job interface{}) {
}

// SubmitJobAfter submits the job to the work queue after the given time period
func (w *worker) SubmitJobAfter(job interface{}, submitAfter time.Duration) {
}
