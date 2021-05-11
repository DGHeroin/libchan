package common

import (
    "container/list"
    "sync"
    "time"
)
// 带超时的消息队列
type MQ struct {
    mu   sync.Mutex
    cond *sync.Cond
    list *list.List
}

func NewMQ() *MQ {
    m := &MQ{}
    m.cond = sync.NewCond(&m.mu)
    m.list = list.New()
    return m
}

func (m *MQ) Add(value ...interface{}) {
    m.mu.Lock()
    for _, val := range value {
        m.list.PushBack(val)
    }
    m.cond.Broadcast()
    m.mu.Unlock()
}

func (m *MQ) Wait(timeout time.Duration, min int, max int) (result []interface{}) {
    if min <= 0 {
        min = 1
    }
    isTimeout := false
    var timer *time.Timer
    if timeout > 0 {
        timer = time.AfterFunc(timeout, func() {
            m.mu.Lock()
            defer m.mu.Unlock()
            timer = nil
            isTimeout = true
            m.cond.Broadcast()
        })
    }
    m.mu.Lock()
add:
    if isTimeout {
        m.mu.Unlock()
        return
    }
    if m.list.Len() < min {
        m.cond.Wait()
        goto add
    }
    if max > 0 && m.list.Len() > max {
        for e := m.list.Front(); e != nil && max > 0; e = m.list.Front() {
            max--
            result = append(result, e.Value)
            m.list.Remove(e)
        }
    } else {
        for e := m.list.Front(); e != nil; e = e.Next() {
            result = append(result, e.Value)
        }
        m.list.Init()
    }
    if timer != nil {
        timer.Stop()
    }
    m.mu.Unlock()
    return
}
