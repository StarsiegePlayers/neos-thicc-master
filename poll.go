package main

type PollService struct {
	Config   *Configuration
	Services *map[string]Service

	Service
	Component
}

func (p *PollService) Init(args map[string]interface{}) (err error) {
	p.Component = Component{
		Name:   "Server Poll",
		LogTag: "poll",
	}
	var ok bool
	p.Config, ok = args["config"].(*Configuration)
	if !ok {
		return ErrorInvalidArgument
	}

	p.Services, ok = args["services"].(*map[string]Service)
	if !ok {
		return ErrorInvalidArgument
	}

	return
}

func (p *PollService) Rehash() {

}

func (p *PollService) Run() {

}

func (p *PollService) Shutdown() {

}
