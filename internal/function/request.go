package function

import (
	"fmt"
	"time"
)

// Request represents a single function invocation.
type Request struct {
	ReqId      string
	Fun        *Function
	Params     map[string]interface{}
	Arrival    time.Time
	ExecReport ExecutionReport
	RequestQoS
	CanDoOffloading bool
	Async           bool
	NetLatencies    map[string]interface{}
}

type RequestQoS struct {
	Class    ServiceClass
	MaxRespT float64
}

type ExecutionReport struct {
	Result         string
	ResponseTime   float64
	IsWarmStart    bool
	InitTime       float64
	OffloadLatency float64
	Duration       float64
	SchedAction    string
}

type Response struct {
	Success bool
	ExecutionReport
}

type AsyncResponse struct {
	ReqId string
}

func (r *Request) String() string {
	return fmt.Sprintf("Rq-%s", r.Fun.Name, r.ReqId)
}

type ServiceClass int64

const (
	LOW               ServiceClass = 0
	HIGH_PERFORMANCE               = 1
	HIGH_AVAILABILITY              = 2
)
