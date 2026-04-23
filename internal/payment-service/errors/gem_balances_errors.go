package errors

type InsufficientGems struct {
	Message string
}

func NewInsufficientGems(message string) error {
	return &InsufficientGems{Message: message}
}

func (e *InsufficientGems) Error() string {
	return e.Message
}

type CosmeticNotFound struct {
	Message string
}

func NewCosmeticNotFound(message string) error {
	return &CosmeticNotFound{Message: message}
}

func (e *CosmeticNotFound) Error() string {
	return e.Message
}

type CosmeticAlreadyPurchased struct {
	Message string
}

func NewCosmeticAlreadyPurchased(message string) error {
	return &CosmeticAlreadyPurchased{Message: message}
}

func (e *CosmeticAlreadyPurchased) Error() string {
	return e.Message
}
