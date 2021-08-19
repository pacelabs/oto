package pleasantries

// Welcomer welcomes people.
type Welcomer interface {
	// Welcome makes a welcome message for somebody.
	Welcome(WelcomeRequest) WelcomeResponse
}

// WelcomeRequest is the request object for Welcomer.Welcome.
type WelcomeRequest struct {
	// To is the address of the person to send the message to.
	// example: "your@email.com"
	// featured: true
	To string `json:"recipients"`
	// Name is the name of the person to welcome.
	// example: "John Smith"
	Name *string
	// The number of times to send the message.
	// example: 3
	Times int
	// CustomerDetails are the details about the customer.
	CustomerDetails *CustomerDetails
}

// WelcomeResponse is the response object for Welcomer.Welcome.
type WelcomeResponse struct {
	// Message is the welcome message.
	// example: "Welcome John Smith."
	Message string
}

type CustomerDetails struct {
	// NewCustomer indicates whether this is a new customer
	// or not.
	// example: true
	NewCustomer bool
}
