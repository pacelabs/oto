package pleasantries

import (
	"github.com/meitner-se/oto/testdata/services"
)

// GreeterService is a polite API.
// You will love it.
// strapline: "A lovely greeter service"
type GreeterService interface {
	// Greet creates a Greeting for one or more people.
	// featured: true
	Greet(GreetRequest) GreetResponse
	// GetGreetings gets a range of saved Greetings.
	// featured: false
	GetGreetings(GetGreetingsRequest) GetGreetingsResponse
}

// GreetRequest is the request object for GreeterService.Greet.
type GreetRequest struct {
	// Names are the names of the people to greet.
	// example: ["Mat", "David"]
	Names []string
}

// GreetResponse is the response object containing a
// person's greeting.
type GreetResponse struct {
	// Greeting is the generated Greeting.
	Greeting Greeting
}

// GetGreetingsRequest is the request object for GreeterService.GetGreetings.
// featured: true
type GetGreetingsRequest struct {
	// Page describes which page of data to get.
	Page services.Page `tagtest:"value,option1,option2"`
}

// GetGreetingsResponse is the respponse object for GreeterService.GetGreetings.
// featured: false
type GetGreetingsResponse struct {
	Greetings []Greeting
}

// Greeting contains the pleasentry.
type Greeting struct {
	// Text is the message.
	// example: "Hello there"
	Text string
}
