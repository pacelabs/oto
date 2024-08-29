package unexporteddeeper

import "math/big"

type CalculatorService interface {
	Calculate(CalculateRequest) CalculateResponse
}

type CalculateRequest struct{}

type CalculateResponse struct {
	// big.Int has unexported fields
	Result big.Int
}
