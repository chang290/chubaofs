package exporter

import (
	"fmt"
	"sync"
	"time"
)

var (
	TPPool = &sync.Pool{New: func() interface{} {
		return new(TimePoint)
	}}
	TPCntPool = &sync.Pool{New: func() interface{} {
		return new(TimePointCount)
	}}
	TPCh = make(chan *TimePoint, ChSize)
)

func collectTP() {
	for {
		m := <- TPCh
		metric := m.Metric()
		metric.Set(float64(m.val.Get()))
		TPPool.Put(m)
	}
}

type TimePoint struct {
	Gauge
	startTime time.Time
}

type TimePointCount struct {
	tp *TimePoint
	cnt *Counter
}

func NewTP(name string) (tp *TimePoint) {
	if ! enabled {
		return
	}
	tp = TPPool.Get().(*TimePoint)
	tp.name = metricsName(name)
	tp.startTime = time.Now()
	return
}

func (tp *TimePoint) Set() {
	if ! enabled {
		return
	}
	val := time.Since(tp.startTime).Nanoseconds()
	tp.val.Set(val)
	tp.publish()
}

func NewTPCnt(name string) (tpc *TimePointCount) {
	if ! enabled {
		return
	}
	tpc = TPCntPool.Get().(*TimePointCount)
	tpc.tp = NewTP(name)
	tpc.cnt = NewCounter(fmt.Sprintf("%s_count", name))
	return
}

func (tpc *TimePointCount) Set() {
	if ! enabled {
		return
	}
	tpc.tp.Set()
	tpc.cnt.Add(1)
	TPCntPool.Put(tpc)
}

func (tp *TimePoint)publish() {
	select {
	case TPCh <- tp:
	default:
	}
}

