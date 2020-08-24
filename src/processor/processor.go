package processor

type processor struct {
	drive *ReptileDriver
}

func New(d ReptileDriver) *processor {
	return &processor{drive: &d}
}

func (p *processor) Run() {

}
