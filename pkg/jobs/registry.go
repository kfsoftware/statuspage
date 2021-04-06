package jobs

import (
	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"sync"
	"time"
)

type SchedulerRegistry struct {
	jobs map[string]cron.EntryID
	cron *cron.Cron
	m    sync.Mutex
}

func NewSchedulerRegistry(tz *time.Location) *SchedulerRegistry {
	c := cron.New(
		cron.WithSeconds(),
		cron.WithLocation(tz),
	)
	c.Start()
	return &SchedulerRegistry{
		jobs: map[string]cron.EntryID{},
		cron: c,
	}
}

func (r *SchedulerRegistry) Register(id string, duration time.Duration, cmd func()) error {
	r.m.Lock()
	defer r.m.Unlock()
	sched := cron.Every(duration)
	_, exists := r.jobs[id]
	if exists {
		return errors.Errorf("Job already exists for %s", id)
	}
	entryId := r.cron.Schedule(sched, cron.FuncJob(cmd))
	r.jobs[id] = entryId
	return nil
}

func (r *SchedulerRegistry) Unregister(id string) error {
	r.m.Lock()
	defer r.m.Unlock()
	entryId, exists := r.jobs[id]
	if !exists {
		return errors.Errorf("Job doesn't exist for %s", id)
	}
	r.cron.Remove(entryId)
	return nil
}
