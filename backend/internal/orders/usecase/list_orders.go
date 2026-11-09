package usecase

import (
	"context"
	"time"
	"vibe_market/backend/internal/orders/domain"
)

type ListOrdersOutDTO struct {
	Orders []OrderDTO `json:"orders"`
}

type OrderNamedEntityDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type OrderPickupPointDTO struct {
	ID        int64   `json:"id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
}

type OrderPromocodeDTO struct {
	ID                  int64     `json:"id"`
	Code                string    `json:"code"`
	Discount            int       `json:"discount"`
	AvailableCategoryID *int64    `json:"availableCategoryId"`
	StartDatetime       time.Time `json:"startDatetime"`
	EndDatetime         time.Time `json:"endDatetime"`
}

type OrderProductDTO struct {
	ID          int64               `json:"id"`
	ProductID   int64               `json:"productId"`
	Name        string              `json:"name"`
	Description *string             `json:"description"`
	Category    OrderNamedEntityDTO `json:"category"`
	Seller      OrderNamedEntityDTO `json:"seller"`
	PictureURLs []string            `json:"pictureUrls"`
	Count       int                 `json:"count"`
	UnitPrice   int64               `json:"unitPrice"`
	TotalPrice  int64               `json:"totalPrice"`
}

type OrderNotificationDTO struct {
	ID        int64               `json:"id"`
	Status    OrderNamedEntityDTO `json:"status"`
	IsRead    bool                `json:"isRead"`
	CreatedAt time.Time           `json:"createdAt"`
}

type OrderDTO struct {
	ID            int64                  `json:"id"`
	Status        OrderNamedEntityDTO    `json:"status"`
	CreatedAt     time.Time              `json:"createdAt"`
	UpdatedAt     time.Time              `json:"updatedAt"`
	PickupPoint   OrderPickupPointDTO    `json:"pickupPoint"`
	Promocode     *OrderPromocodeDTO     `json:"promocode"`
	Products      []OrderProductDTO      `json:"products"`
	Notifications []OrderNotificationDTO `json:"notifications"`
	TotalCount    int64                  `json:"totalCount"`
	TotalPrice    int64                  `json:"totalPrice"`
}

func ListOrdersUsecase(ctx context.Context, buyerID int64, repo domain.OrderRepo) (ListOrdersOutDTO, error) {
	orders, err := repo.ListByBuyer(ctx, buyerID)
	if err != nil {
		return ListOrdersOutDTO{}, err
	}
	out := ListOrdersOutDTO{Orders: make([]OrderDTO, 0, len(orders))}
	for _, order := range orders {
		dto := OrderDTO{
			ID: order.ID, Status: OrderNamedEntityDTO(order.Status),
			CreatedAt: order.CreatedAt.UTC(), UpdatedAt: order.UpdatedAt.UTC(),
			PickupPoint:   OrderPickupPointDTO(order.PickupPoint),
			Products:      make([]OrderProductDTO, 0, len(order.Products)),
			Notifications: make([]OrderNotificationDTO, 0, len(order.Notifications)),
		}
		if p := order.Promocode; p != nil {
			dto.Promocode = &OrderPromocodeDTO{ID: p.ID, Code: p.Code, Discount: p.Discount, AvailableCategoryID: p.AvailableCategoryID, StartDatetime: p.StartDatetime.UTC(), EndDatetime: p.EndDatetime.UTC()}
		}
		for _, p := range order.Products {
			pictures := append([]string{}, p.PictureURLs...)
			total := int64(p.Count) * p.Price
			dto.Products = append(dto.Products, OrderProductDTO{
				ID: p.ID, ProductID: p.ProductID, Name: p.Name, Description: p.Description,
				Category: OrderNamedEntityDTO(p.Category), Seller: OrderNamedEntityDTO(p.Seller),
				PictureURLs: pictures, Count: p.Count, UnitPrice: p.Price, TotalPrice: total,
			})
			dto.TotalCount += int64(p.Count)
			dto.TotalPrice += total
		}
		for _, n := range order.Notifications {
			dto.Notifications = append(dto.Notifications, OrderNotificationDTO{ID: n.ID, Status: OrderNamedEntityDTO(n.Status), IsRead: n.IsRead, CreatedAt: n.CreatedAt.UTC()})
		}
		out.Orders = append(out.Orders, dto)
	}
	return out, nil
}
