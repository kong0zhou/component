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

// 启动调度器
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
			// 这里是串行
			jobChannel := <-d.workerPool
			jobChannel <- job
		}
	}
}

// 进行入队操作，ok是否入队成功
func (d *Dispatcher) InQueue(job Job) (ok bool, err error) {
	if d.workerPool == nil {
		err = fmt.Errorf(`workerPool is nil`)
		logs.Error(err)
		return false, err
	}
	if d.jobQueue == nil {
		err = fmt.Errorf(`jobQueue is nil`)
		logs.Error(err)
		return false, err
	}
	if d.work == nil {
		err = fmt.Errorf(`work is nil`)
		logs.Error(err)
		return false, err
	}
	if job.R == nil {
		err = fmt.Errorf(`job.R is nil`)
		logs.Error(err)
		return false, err
	}
	if job.W == nil {
		err = fmt.Errorf(`job.W is nil`)
		logs.Error(err)
		return false, err
	}
	// d.jobQueue <- job
	select {
	case d.jobQueue <- job:
		return true, nil
	default:
		return false, nil
	}
}

type Queue struct {
	dispatcher *Dispatcher
}

// 新建一个处理器，如果新建不成功直接panic
func NewQueue(maxWorkers int, maxQueueSize int, work http.Handler) *Queue {
	if maxWorkers <= 0 {
		err := fmt.Errorf(`maxWorkers cannot be less than 0`)
		logs.Error(err)
		panic(err)
		// return nil
	}
	if maxQueueSize <= 0 {
		err := fmt.Errorf(`maxQueueSize cannot be less than 0`)
		logs.Error(err)
		panic(err)
		// return nil
	}
	if work == nil {
		err := fmt.Errorf(`work cannot be nil`)
		logs.Error(err)
		panic(err)
		// return nil
	}
	d, err := NewDispatcher(maxWorkers, maxQueueSize, work)
	if err != nil {
		logs.Error(err)
		panic(err)
		// return nil
	}
	err = d.Run()
	if err != nil {
		logs.Error(err)
		panic(err)
		// return nil
	}
	return &Queue{
		dispatcher: d,
	}
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
	ok, err := q.dispatcher.InQueue(*job)
	if err != nil {
		logs.Error(err)
		return
	}
	if !ok {
		logs.Warn(`服务器繁忙，入队失败`)
		return
	}
	_ = <-job.IsFinished
	return
}
