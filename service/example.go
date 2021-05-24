package service

type Example struct{}

func init() {
	registerMicroService(microServiceName("example"), newExample)
}

func newExample() MicroService {
	return &Example{}
}

func (e *Example) Rpc() error {
	return nil
}
