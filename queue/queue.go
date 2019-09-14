package queue

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego/logs"
)

// 作业的定义
type Job struct {
	R          *http.Request
	W          http.ResponseWriter
	IsFinished chan bool
}

func NewJob(w http.ResponseWriter, r *http.Request) (*Job, error) {
	if w == nil || r == nil {
		err := fmt.Errorf(`w or r cannot be nil`)
		logs.Error(err)
		return nil, err
	}
	return &Job{
		R:          r,
		W:          w,
		IsFinished: make(chan bool),
	}, nil
}

// 工作者类
type Worker struct {
	WorkerPool chan chan Job
	work       http.Handler
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job, work http.Handler) (*Worker, error) {
	if workerPool == nil {
		err := fmt.Errorf(`workerPool is nil`)
		logs.Error(err)
		return nil, err
	}
	if work == nil {
		err := fmt.Errorf(`work is nil`)
		logs.Error(err)
		return nil, err
	}
	return &Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool),
		work:       work,
	}, nil
}

func (w *Worker) Start() (err error) {
	if w.WorkerPool == nil {
		err = fmt.Errorf(`workerPool is nil`)
		logs.Error(err)
		return
	}
	if w.JobChannel == nil {
		err = fmt.Errorf(`jobChannel is nil`)
		logs.Error(err)
		return
	}
	if w.quit == nil {
		err = fmt.Errorf(`quit is nil`)
		logs.Error(err)
		return
	}
	if w.work == nil {
		err = fmt.Errorf(`work is nil`)
		logs.Error(err)
		return
	}
	go func() {
		for {
			// 将工作人员放到空闲池中
			w.WorkerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				logs.Info(`收到一份工作请求`)
				w.work.ServeHTTP(job.W, job.R)
				job.IsFinished <- true
				//收到一份工作请求
			case <-w.quit:
				return
			}
		}
	}()
	return nil
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

// 调度者类
type Dispatcher struct {
	// 最大的工作者数量
	maxWorkers int
	// 工作者执行的工作
	work http.Handler
	// 工作队列
	jobQueue chan Job
	// 空闲池
	workerPool chan chan Job
}

func NewDispatcher(maxWorkers int, maxQueueSize int, work http.Handler) (*Dispatcher, error) {
	if maxWorkers <= 0 {
		err := fmt.Errorf(`maxWorkers cannot be less than 0`)
		logs.Error(err)
		return nil, err
	}
	if maxQueueSize <= 0 {
		err := fmt.Errorf(`maxQueueSize cannot be less than 0`)
		logs.Error(err)
		return nil, err
	}
	if work == nil {
		err := fmt.Errorf(`work cannot be nil`)
		logs.Error(err)
		return nil, err
	}
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{
		workerPool: pool,
		maxWorkers: maxWorkers,
		jobQueue:   make(chan Job, maxQueueSize),
		work:       work,
	}, nil
}

func (d *Dispatcher) Run() (err error) {
	if d.workerPool == nil {
		err = fmt.Errorf(`workerPool is nil`)
		logs.Error(err)
		return err
	}
	if d.jobQueue == nil {
		err = fmt.Errorf(`jobQueue is nil`)
		logs.Error(err)
		return err
	}
	if d.work == nil {
		err = fmt.Errorf(`work is nil`)
		logs.Error(err)
		return err
	}
	for i := 0; i < d.maxWorkers; i++ {
		worker, err := NewWorker(d.workerPool, d.work)
		if err != nil {
			logs.Error(err)
			return err
		}
		err = worker.Start()
		if err != nil {
			logs.Error(err)
			return err
		}
	}
	go d.dispatch()
	return nil
}

func (d *Dispatcher) dispatch() (err error) {
	if d.workerPool == nil {
		err = fmt.Errorf(`workerPool is nil`)
		logs.Error(err)
		return err
	}
	if d.jobQueue == nil {
		err = fmt.Errorf(`jobQueue is nil`)
		logs.Error(err)
		return err
	}
	if d.work == nil {
		err = fmt.Errorf(`work is nil`)
		logs.Error(err)
		return err
	}
	for {
		select {
		case job := <-d.jobQueue:
			go func(job Job) {
				jobChannel := <-d.workerPool
				jobChannel <- job
			}(job)
		}
	}
}

func (d *Dispatcher) InQueue(job Job) (err error) {
	if d.workerPool == nil {
		err = fmt.Errorf(`workerPool is nil`)
		logs.Error(err)
		return err
	}
	if d.jobQueue == nil {
		err = fmt.Errorf(`jobQueue is nil`)
		logs.Error(err)
		return err
	}
	if d.work == nil {
		err = fmt.Errorf(`work is nil`)
		logs.Error(err)
		return err
	}
	if job.R == nil {
		err = fmt.Errorf(`job.R is nil`)
		logs.Error(err)
		return
	}
	if job.W == nil {
		err = fmt.Errorf(`job.W is nil`)
		logs.Error(err)
		return
	}
	d.jobQueue <- job
	return nil
}

type Queue struct {
	// maxWorkers   int
	// work       http.Handler
	dispatcher *Dispatcher
}

func NewQueue(maxWorkers int, maxQueueSize int, work http.Handler) (*Queue, error) {
	if maxWorkers <= 0 {
		err := fmt.Errorf(`maxWorkers cannot be less than 0`)
		logs.Error(err)
		return nil, err
	}
	if maxQueueSize <= 0 {
		err := fmt.Errorf(`maxQueueSize cannot be less than 0`)
		logs.Error(err)
		return nil, err
	}
	if work == nil {
		err := fmt.Errorf(`work cannot be nil`)
		logs.Error(err)
		return nil, err
	}
	d, err := NewDispatcher(maxWorkers, maxQueueSize, work)
	if err != nil {
		logs.Error(err)
		return nil, err
	}
	err = d.Run()
	if err != nil {
		logs.Error(err)
		return nil, err
	}
	return &Queue{
		dispatcher: d,
		// work:       work,
	}, nil
}

func (q *Queue) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if q.dispatcher == nil {
		err := fmt.Errorf(`q.dispatcher is nil`)
		logs.Error(err)
		return
	}
	job, err := NewJob(w, r)
	if err != nil {
		logs.Error(err)
		return
	}
	err = q.dispatcher.InQueue(*job)
	if err != nil {
		logs.Error(err)
		return
	}
	_ = <-job.IsFinished
	return
}
