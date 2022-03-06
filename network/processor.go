package network

type Processor interface {
	// Route must goroutine safe
	Route(msg interface{}, userData interface{}) error
	// Unmarshal must goroutine safe
	Unmarshal(data []byte) (interface{}, error)
	// Marshal must goroutine safe
	Marshal(msg interface{}) ([][]byte, error)
}
