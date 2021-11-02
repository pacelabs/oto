package maps

type GreeterService interface {
	Greet(GreetRequest) GreetResponse
	GreetMultiple(GreetMultipleRequest) GreetMultipleResponse
}

type GreetRequest struct {
	GreetingMap map[string]int
}

type GreetResponse struct {
	Greeting map[string]string `json:"greeting,omitempty"`
}

type GreetMultipleRequest struct {
	GreetingMap map[string][]GreetRequest
}

type GreetMultipleResponse struct {
	Greeting map[string][]GreetResponse `json:"greeting,omitempty"`
}
