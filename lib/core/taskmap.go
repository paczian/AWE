package core

import (
	"fmt"
)

type TaskMap struct {
	RWMutex
	_map map[Task_Unique_Identifier]*Task
}

func NewTaskMap() (t *TaskMap) {
	t = &TaskMap{
		_map: make(map[Task_Unique_Identifier]*Task),
	}
	t.RWMutex.Init("TaskMap")
	return t

}

//--------task accessor methods-------

func (tm *TaskMap) Len() (length int, err error) {
	read_lock, err := tm.RLockNamed("Len")
	if err != nil {
		return
	}
	defer tm.RUnlockNamed(read_lock)
	length = len(tm._map)
	return
}

func (tm *TaskMap) Get(taskid Task_Unique_Identifier, lock bool) (task *Task, ok bool, err error) {
	if lock {
		read_lock, xerr := tm.RLockNamed("Get")
		if xerr != nil {
			err = xerr
			return
		}
		defer tm.RUnlockNamed(read_lock)
	}

	task, ok = tm._map[taskid]
	return
}

func (tm *TaskMap) Has(taskid Task_Unique_Identifier, lock bool) (ok bool, err error) {
	if lock {
		read_lock, xerr := tm.RLockNamed("Get")
		if xerr != nil {
			err = xerr
			return
		}
		defer tm.RUnlockNamed(read_lock)
	}

	_, ok = tm._map[taskid]
	return
}

func (tm *TaskMap) GetTasks() (tasks []*Task, err error) {

	tasks = []*Task{}

	read_lock, err := tm.RLockNamed("GetTasks")
	if err != nil {
		return
	}
	defer tm.RUnlockNamed(read_lock)

	for _, task := range tm._map {
		tasks = append(tasks, task)
	}

	return
}

func (tm *TaskMap) Delete(taskid Task_Unique_Identifier) (task *Task, ok bool) {
	tm.LockNamed("Delete")
	defer tm.Unlock()
	delete(tm._map, taskid) // TODO should get write lock on task first
	return
}

func (tm *TaskMap) Add(task *Task) (err error) {
	tm.LockNamed("Add")
	defer tm.Unlock()
	id, err := task.GetId()
	if err != nil {
		return
	}
	_, has_task := tm._map[id]
	if has_task {
		err = fmt.Errorf("(TaskMap/Add) task is already in TaskMap")
		return
	}

	task_state, _ := task.GetState()
	if task_state == TASK_STAT_INIT {
		err = task.SetState(TASK_STAT_PENDING, false)
		if err != nil {
			return
		}
	}

	tm._map[id] = task
	return
}
