package usecase

import (
	"github.com/LucasBelusso1/GoCleanArchChallange/internal/entity"
)

type ListOrderUseCase struct {
	OrderRepository entity.OrderRepositoryInterface
}

func NewListOrderUseCase(OrderRepository entity.OrderRepositoryInterface) *ListOrderUseCase {
	return &ListOrderUseCase{OrderRepository: OrderRepository}
}

func (c *ListOrderUseCase) Execute() ([]OrderOutputDTO, error) {
	orders, err := c.OrderRepository.List()

	if err != nil {
		return []OrderOutputDTO{}, err
	}

	var ordersOutputDTO []OrderOutputDTO
	for _, order := range orders {
		ordersOutputDTO = append(ordersOutputDTO, OrderOutputDTO{
			ID:         order.ID,
			Price:      order.Price,
			Tax:        order.Tax,
			FinalPrice: order.FinalPrice,
		})
	}

	return ordersOutputDTO, nil
}
