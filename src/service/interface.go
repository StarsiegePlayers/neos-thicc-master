package service

const LocalhostAddress = "127.0.0.1"

type Interface interface {
	Init(services *map[ID]Interface) error
	Status() LifeCycle
}

type Maintainable interface {
	Maintenance()
}

type DailyMaintainable interface {
	DailyMaintenance()
}

type Getable interface {
	Get(string) string
}

type Rehashable interface {
	Rehash()
}

type Runnable interface {
	Run()
	Shutdown()
	Rehashable
}
