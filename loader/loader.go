package loader

import (
	"context"
	"fmt"
	"github.com/ProjectAthenaa/sonic-core/sonic/core"
	"github.com/ProjectAthenaa/task-loader-service/helpers"
	"github.com/go-redis/redis/v8"
	"github.com/prometheus/common/log"
	"github.com/scylladb/go-set/strset"
	"os"
	"sync"
	"time"
)

var rdb redis.UniversalClient

func init() {
	rdb = core.Base.GetRedis("cache")
}

type Loader struct {
	ctx         context.Context
	cancelFunc  context.CancelFunc
	loadedTasks *strset.Set
	locker      *sync.Mutex
}

func NewLoader() *Loader {
	ctx, cancelFund := context.WithCancel(context.Background())

	return &Loader{
		ctx:         ctx,
		cancelFunc:  cancelFund,
		loadedTasks: strset.New(),
		locker:      &sync.Mutex{},
	}
}

func (l *Loader) Start() {
	if os.Getenv("DEBUG") == "1" {
		go l.debug()
	}

	log.Info("Starting Loader")
	go l.deleteListener()
	l.loader()
}

func (l *Loader) loader() {
	for range time.Tick(time.Millisecond * 50) {
		tasks := l.fetchTasks()
		rdb.SAdd(l.ctx, "scheduler:scheduled", tasks...)
		l.loadedTasks.Add(convertToStringSlice(tasks)...)
	}
}

func (l *Loader) deleteListener() {
	pubSub := rdb.Subscribe(l.ctx, "scheduler:tasks-deleted")
	defer pubSub.Close()
	for deletedTask := range pubSub.Channel() {
		deletedTask := deletedTask
		go func() {
			rdb.SRem(l.ctx, deletedTask.Payload)
			rdb.Publish(l.ctx, fmt.Sprintf("tasks:commands:%s", helpers.SHA1(deletedTask.Payload)), "STOP")
			log.Info("Deleted Task ", deletedTask.Payload)
		}()
	}
}

func (l *Loader) debug() {
	for range time.Tick(time.Second) {
		fmt.Println(l.loadedTasks)
		fmt.Println("-")
	}
}
