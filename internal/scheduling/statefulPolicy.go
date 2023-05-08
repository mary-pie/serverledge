package scheduling

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/grussorusso/serverledge/internal/node"
	wr "github.com/mroth/weightedrand"
)

// CloudEdgePolicy supports only Edge-Cloud Offloading
type StatefulPolicy struct {
	totCPUs float64
	totMem  int64
}

func (p *StatefulPolicy) Init() {
	p.totCPUs = node.Resources.AvailableCPUs
	p.totMem = node.Resources.AvailableMemMB
}

func (p *StatefulPolicy) OnCompletion(r *scheduledRequest) {

}

func (p *StatefulPolicy) OnArrival(r *scheduledRequest) {

	isStateless := r.Class != 1 //convert r.Class to bool, assuming that r.Class can only be 0/1

	containerID, err := node.AcquireWarmContainer(r.Fun)               //return 1 with prob
	availableRes := checkAvailableRes(r.Fun.MemoryMB, r.Fun.CPUDemand) //check resources available
	canOffload := false
	if isStateless {
		canOffload = offloadWithProb(p)
		log.Println("randomPicker returns", canOffload)
	}

	if !availableRes || (availableRes && isStateless && canOffload && r.CanDoOffloading) {
		if availableRes {
			log.Println("OFFLOADING with probability p")
		}
		handleCloudOffload(r)
	} else {
		if err == nil {
			log.Println("available warm container --> execute locally")
			execLocally(r, containerID, true)
		} else {
			log.Println("cold start")
			handleColdStart(r)
		}
	}
}

func randomPicker(prob uint) int {
	rand.Seed(time.Now().UTC().UnixNano()) // always seed random!

	chooser, _ := wr.NewChooser(
		wr.Choice{Item: 0, Weight: 10 - prob}, //---> offloading with probability 1-prob
		wr.Choice{Item: 1, Weight: prob},      //---> offloading with probability prob
	)
	result := chooser.Pick().(int)
	log.Println("RESULT RANDOM PICKER", result)
	return chooser.Pick().(int)
}

/*
Offload with probability:

	prob = (L - threshold) / (1 - threshold) if threshold <= L <= 1, or
	prob = 0 if L 0 <= L < threshold
*/
func offloadWithProb(p *StatefulPolicy) bool {
	usedCPUs_frac := 1.0 - (node.Resources.AvailableCPUs / p.totCPUs)
	usedMem_frac := 1.0 - (float64(node.Resources.AvailableMemMB) / float64(p.totMem))
	log.Println("usedCPU_frac", usedCPUs_frac)
	log.Println("usedMem_frac", usedMem_frac)
	log.Println("avail cpu frac:", node.Resources.AvailableCPUs/p.totCPUs)
	log.Println("avail mem frac:", float64(node.Resources.AvailableMemMB)/float64(p.totMem))

	L := math.Max(usedCPUs_frac, float64(usedMem_frac))
	log.Println("L: ", L)
	threshold := 0.5
	prob := 0.0
	canOffload := false
	if threshold <= L && L <= 1.0 { //L exceeds threshold
		prob = (L - threshold) / (1.0 - threshold)
		canOffload = randomPicker(uint(prob*10)) != 0

	}
	log.Println("prob: ", prob)
	return canOffload //true iff randomPicker() returns 1

}

func checkAvailableRes(memDemand int64, cpuDemand float64) bool {
	if node.Resources.AvailableCPUs < cpuDemand || node.Resources.AvailableMemMB < memDemand {
		return false
	}
	return true
}
