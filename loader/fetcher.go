package loader

import (
	"fmt"
	"github.com/ProjectAthenaa/sonic-core/sonic/core"
	"github.com/ProjectAthenaa/sonic-core/sonic/database/ent/task"
	"time"
)

func (l *Loader) fetchTasks() []interface{} {
	tasks, err := core.
		Base.
		GetPg("pg").
		Task.Query().
		Where(
			task.StartTimeGTE(
				time.Now(),
			),
			task.StartTimeLTE(time.Now().Add(time.Minute*30)),
		).
		WithProduct().
		WithProfileGroup().
		WithProxyList().
		All(l.ctx)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	ids := make([]interface{}, len(tasks))
	processingTasks := rdb.SMembers(l.ctx, "scheduler:processing").Val()

	for i, tk := range tasks {
		var processing bool
		for _, processingTask := range processingTasks {
			if processingTask == tk.ID.String() {
				processing = true
				break
			}
		}
		if !processing {
			ids[i] = tk.ID
		}

	}

	return ids
}

func convertToStringSlice(data []interface{}) []string {
	var out = make([]string, len(data))

	for i := range data {
		out[i] = fmt.Sprint(data[i])
	}
	return out
}
