package service

type LifeCycle byte

const (
	Static = LifeCycle(iota)
	Stopped
	Starting
	Running
	Rehashing
	Restarting
	Stopping
)

var LifeCycleStrings = map[LifeCycle]string{
	Static:     "Static",
	Stopped:    "Stopped",
	Starting:   "Starting",
	Running:    "Running",
	Rehashing:  "Rehashing",
	Restarting: "Restarting",
	Stopping:   "Stopping",
}

func (l LifeCycle) String() string {
	return LifeCycleStrings[l]
}
