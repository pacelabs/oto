package unexported

type CalculatorService interface {
	Calculate(CalculateRequest) CalculateResponse
}

type CalculateRequest struct{}

type CalculateResponse struct {
	result int
}
