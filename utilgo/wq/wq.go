package wq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/google/uuid"
	"k8s.io/client-go/util/workqueue"
)

var (
	dialTimeout    = 2 * time.Second
	requestTimeout = 10 * time.Second
	etcdEndpoint   = "127.0.0.1:2379"
)

/**
IDEAS:
	1. take workqueue package and enhance it: add persistence to etcd storage
	2. store data in Add func and Delete it in Put func (in this case we should get all data from etcd on start and put it in queue)
*/

type WorkQueue interface {
	Add(ctx context.Context, event *Event) error
	Get(ctx context.Context) (event *Event, shutdown bool, err error)
	Len() int
}

type workQueue struct {
	storage *clientv3.Client
	queue   *workqueue.Type
}

type Event struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

func NewWorkQueue(name string) (WorkQueue, error) {
	t := workqueue.NewNamed(name)
	storage, err := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   []string{etcdEndpoint},
	})
	if err != nil {
		return nil, err
	}
	return &workQueue{storage: storage, queue: t}, nil
}

func (w *workQueue) Add(ctx context.Context, event *Event) error {
	jsonData, err := json.Marshal(event)
	if err != nil {
		return err
	}

	if _, err = w.storage.Put(ctx, event.ID, string(jsonData)); err != nil {
		return err
	}
	w.queue.Add(event)
	return nil
}

func (w *workQueue) Get(ctx context.Context) (event *Event, shutdown bool, err error) {
	item, shutdown := w.queue.Get()
	event = item.(*Event)

	if _, err = w.storage.Delete(ctx, event.ID); err != nil {
		return nil, false, err
	}
	w.queue.Done(item)
	return event, shutdown, nil
}

func (w *workQueue) Len() int {
	return w.queue.Len()
}

func example() {
	queue, err := NewWorkQueue("example")
	if err != nil {
		log.Fatal(err)
	}
	data := &Event{
		ID:   uuid.New().String(),
		Data: "data",
	}
	ctx := context.Background()

	if err := queue.Add(ctx, data); err != nil {
		log.Fatal(err)
	}

	res, _, err := queue.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("data = ", res)

	length := queue.Len()
	log.Println("queue length = ", length)
}
