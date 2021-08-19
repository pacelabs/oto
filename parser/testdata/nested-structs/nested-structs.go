package nested

type GreeterService interface {
	Greet(GreetRequest) GreetResponse
}

type GreetRequest struct {
	Person struct {
		Title string `json:"person_title,omitempty"`
		Name  string `json:"person_name,omitempty"`
	} `json:"person,omitempty"`
	Formats []string `json:"formats,omitempty"`
}

type GreetResponse struct {
	Greeting string `json:"greeting,omitempty"`
}
