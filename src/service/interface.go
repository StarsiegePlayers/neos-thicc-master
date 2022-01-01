package service

const LocalhostAddress = "127.0.0.1"

type Interface interface {
	Init(services *map[ID]Interface) error
	Run()
	Rehash()
	Shutdown()

	Acquirable
}

type Maintainable interface {
	Maintenance()
}

type DailyMaintainable interface {
	DailyMaintenance()
}

type Acquirable interface {
	Get() string
}
