package repository

import (
	"context"
	"fmt"

	"wb-tech-backend/internal/core"
	"wb-tech-backend/internal/models"
	"wb-tech-backend/internal/pkg/pgdb"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Deps struct {
	QueryManager       *pgdb.QueryManager
	TransactionManager *pgdb.TransactionManager
	Cash               map[string]models.Order
}

type Repository struct {
	Deps
}

func NewRepository(ctx context.Context, cfg *core.Config) (*Repository, error) {
	pool, err := pgxpool.Connect(ctx, cfg.Storage.URL)
	if err != nil {
		return nil, err
	}
	qm := pgdb.NewQueryManager(pool)
	tm := pgdb.NewTransactionManager(pool)
	r := &Repository{
		Deps{
			QueryManager:       qm,
			TransactionManager: tm,
			Cash:               make(map[string]models.Order),
		},
	}
	orders, err := r.GetOrders(ctx)
	if err != nil {
		return nil, err
	}
	for _, order := range orders {
		r.Cash[order.OrderId] = order
	}
	return r, nil
}

func (r *Repository) addDelivery(ctx context.Context, delivery models.Delivery) (int64, error) {
	query := sq.Insert("deliveries").
		Columns("name", "phone", "zip", "city", "address", "region", "email").
		Values(delivery.Name, delivery.Phone, delivery.Zip, delivery.City, delivery.Address, delivery.Region, delivery.Email).
		PlaceholderFormat(sq.Dollar).Suffix("RETURNING delivery_id")
	rows, err := r.QueryManager.QuerySq(ctx, query)
	if err != nil {
		return 0, err
	}
	if err = rows.Err(); err != nil {
		return 0, err
	}
	defer rows.Close()
	var deliveryId int64
	for rows.Next() {
		err = rows.Scan(&deliveryId)
		if err != nil {
			return 0, err
		}
	}
	return deliveryId, nil
}
func (r *Repository) addPayment(ctx context.Context, payment models.Payment) (int64, error) {
	query := sq.Insert("payments").
		Columns("transaction", "request_id", "currency", "provider", "amount", "payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
		Values(payment.Transaction, payment.RequestId, payment.Currency, payment.Provider, payment.Amount, payment.PaymentDt, payment.Bank, payment.DeliveryCost, payment.GoodsTotal, payment.CustomFee).
		PlaceholderFormat(sq.Dollar).Suffix("RETURNING payment_id")
	rows, err := r.QueryManager.QuerySq(ctx, query)
	if err != nil {
		return 0, err
	}
	if err = rows.Err(); err != nil {
		return 0, err
	}
	defer rows.Close()
	var paymentId int64
	for rows.Next() {
		err = rows.Scan(&paymentId)
		if err != nil {
			return 0, err
		}
	}
	return paymentId, nil
}
func (r *Repository) addItem(ctx context.Context, item models.Item) (int64, error) {
	query := sq.Insert("items").
		Columns("chrt_id", "track_number", "price", "rid", "name", "sale", "size", "total_price", "nm_id", "brand", "status").
		Values(item.ChrtId, item.TrackNumber, item.Price, item.RId, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status).
		PlaceholderFormat(sq.Dollar).Suffix("RETURNING item_id")
	rows, err := r.QueryManager.QuerySq(ctx, query)
	if err != nil {
		return 0, err
	}
	if err = rows.Err(); err != nil {
		return 0, err
	}
	defer rows.Close()
	var itemId int64
	for rows.Next() {
		err = rows.Scan(&itemId)
		if err != nil {
			return 0, err
		}
	}
	return itemId, nil
}

func (r *Repository) AddOrder(ctx context.Context, order models.Order) error {
	return r.TransactionManager.Tx(ctx, func(ctx context.Context) error {
		deliveryId, err := r.addDelivery(ctx, order.Delivery)
		if err != nil {
			return err
		}
		paymentId, err := r.addPayment(ctx, order.Payment)
		if err != nil {
			return err
		}
		itemsIds := make([]int64, 0)
		for _, item := range order.Items {
			itemId, err := r.addItem(ctx, item)
			if err != nil {
				return err
			}
			itemsIds = append(itemsIds, itemId)
		}
		query := sq.Insert("orders").
			Columns("order_uid", "track_number", "entry", "delivery_id", "payment_id", "items_ids", "locale", "internal_signature", "customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard").
			Values(order.OrderId, order.TrackNumber, order.Entry, deliveryId, paymentId, itemsIds, order.Locale, order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, order.DateCreated, order.OofShard).
			PlaceholderFormat(sq.Dollar).Suffix("RETURNING order_uid")
		rows, err := r.QueryManager.QuerySq(ctx, query)
		if err != nil {
			return err
		}
		if err = rows.Err(); err != nil {
			return err
		}
		defer rows.Close()
		var orderId string
		for rows.Next() {
			err = rows.Scan(&orderId)
			if err != nil {
				return err
			}
		}
		if orderId == order.OrderId {
			r.Cash[order.OrderId] = order
			return nil
		}
		return fmt.Errorf("something goes wrong with add order to database")
	})
}

func (r *Repository) GetOrderById(ctx context.Context, orderId string) (models.Order, error) {
	query := sq.Select("o.order_uid", "o.track_number", "o.entry", "o.items_ids", "o.locale", "o.internal_signature", "o.customer_id", "o.delivery_service", "o.shardkey", "o.sm_id", "o.date_created", "o.oof_shard",
		"d.name", "d.phone", "d.zip", "d.city", "d.address", "d.region", "d.email",
		"p.transaction", "p.request_id", "p.currency", "p.provider", "p.amount", "p.payment_dt", "p.bank", "p.delivery_cost", "p.goods_total", "p.custom_fee").
		From("orders o").Join("deliveries d ON o.delivery_id = d.delivery_id").
		Join("payments p ON o.payment_id = p.payment_id").Where(sq.Eq{"o.order_uid": orderId}).PlaceholderFormat(sq.Dollar)
	rows, err := r.QueryManager.QuerySq(ctx, query)
	if err != nil {
		return models.Order{}, err
	}
	if err = rows.Err(); err != nil {
		return models.Order{}, err
	}
	var order models.Order
	var itemsIds []int64
	for rows.Next() {
		err = rows.Scan(
			&order.OrderId, &order.TrackNumber, &order.Entry, &itemsIds, &order.Locale, &order.InternalSignature,
			&order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId, &order.DateCreated,
			&order.OofShard, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
			&order.Payment.Transaction, &order.Payment.RequestId, &order.Payment.Currency, &order.Payment.Provider,
			&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
			&order.Payment.GoodsTotal, &order.Payment.CustomFee,
		)
		if err != nil {
			return models.Order{}, err
		}
	}
	rows.Close()
	defer rows.Close()
	for _, itemId := range itemsIds {
		query = sq.Select("i.chrt_id", "i.track_number", "i.price", "i.rid", "i.name", "i.sale", "i.size", "i.total_price", "i.nm_id", "i.brand", "i.status").
			From("items i").Where(sq.Eq{"item_id": itemId}).PlaceholderFormat(sq.Dollar)
		rows, err = r.QueryManager.QuerySq(ctx, query)
		if err != nil {
			return models.Order{}, err
		}
		if err = rows.Err(); err != nil {
			return models.Order{}, err
		}

		for rows.Next() {
			var item models.Item
			err = rows.Scan(
				&item.ChrtId, &item.TrackNumber, &item.Price, &item.RId, &item.Name, &item.Sale,
				&item.Size, &item.TotalPrice, &item.NmId, &item.Brand, &item.Status,
			)
			if err != nil {
				return models.Order{}, err
			}
			order.Items = append(order.Items, item)
		}
		rows.Close()
	}

	return order, nil
}
func (r *Repository) GetOrders(ctx context.Context) ([]models.Order, error) {
	query := sq.Select("o.order_uid", "o.track_number", "o.entry", "o.items_ids", "o.locale", "o.internal_signature", "o.customer_id", "o.delivery_service", "o.shardkey", "o.sm_id", "o.date_created", "o.oof_shard",
		"d.name", "d.phone", "d.zip", "d.city", "d.address", "d.region", "d.email",
		"p.transaction", "p.request_id", "p.currency", "p.provider", "p.amount", "p.payment_dt", "p.bank", "p.delivery_cost", "p.goods_total", "p.custom_fee").
		From("orders o").Join("deliveries d ON o.delivery_id = d.delivery_id").
		Join("payments p ON o.payment_id = p.payment_id").PlaceholderFormat(sq.Dollar)
	rows, err := r.QueryManager.QuerySq(ctx, query)
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	orders := make([]models.Order, 0)
	var itemsIds []int64
	for rows.Next() {
		var order models.Order
		err = rows.Scan(
			&order.OrderId, &order.TrackNumber, &order.Entry, &itemsIds, &order.Locale, &order.InternalSignature,
			&order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId, &order.DateCreated,
			&order.OofShard, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
			&order.Payment.Transaction, &order.Payment.RequestId, &order.Payment.Currency, &order.Payment.Provider,
			&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
			&order.Payment.GoodsTotal, &order.Payment.CustomFee,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	rows.Close()
	for _, order := range orders {
		for _, itemId := range itemsIds {
			query = sq.Select("i.chrt_id", "i.track_number", "i.price", "i.rid", "i.name", "i.sale", "i.size", "i.total_price", "i.nm_id", "i.brand", "i.status").
				From("items i").Where(sq.Eq{"item_id": itemId}).PlaceholderFormat(sq.Dollar)
			rows, err = r.QueryManager.QuerySq(ctx, query)
			if err != nil {
				return nil, err
			}
			if err = rows.Err(); err != nil {
				return nil, err
			}

			for rows.Next() {
				var item models.Item
				err = rows.Scan(
					&item.ChrtId, &item.TrackNumber, &item.Price, &item.RId, &item.Name, &item.Sale,
					&item.Size, &item.TotalPrice, &item.NmId, &item.Brand, &item.Status,
				)
				if err != nil {
					return nil, err
				}
				order.Items = append(order.Items, item)
			}
			rows.Close()
		}
	}
	return orders, nil
}
func (r *Repository) GetOrders2(ctx context.Context) ([]models.Order, error) {
	query := sq.Select("*").From("orders").PlaceholderFormat(sq.Dollar)
	rows, err := r.QueryManager.QuerySq(ctx, query)
	if err != nil {
		return nil, err
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	orders := make([]models.Order, 0)
	var itemsIDs []int64
	for rows.Next() {
		var order Order
		err = rows.Scan(
			&order.OrderId, &order.TrackNumber, &order.Entry, &order.DeliveryId,
			&order.PaymentId, &itemsIDs, &order.Locale, &order.InternalSignature,
			&order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId,
			&order.DateCreated, &order.OofShard,
		)
		if err != nil {
			return nil, err
		}
		rows.Close()

		query = sq.Select("*").From("deliveries").Where(sq.Eq{"delivery_id": order.DeliveryId}).PlaceholderFormat(sq.Dollar)
		rows, err = r.QueryManager.QuerySq(ctx, query)
		if err != nil {
			return nil, err
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
		var d models.Delivery
		var id int64
		for rows.Next() {
			err = rows.Scan(&id, &d.Name, &d.Phone, &d.Zip, &d.City, &d.Address, &d.Region, &d.Email)
			if err != nil {
				return nil, err
			}
		}
		rows.Close()

		query = sq.Select("*").From("payments").Where(sq.Eq{"payment_id": order.PaymentId}).PlaceholderFormat(sq.Dollar)
		rows, err = r.QueryManager.QuerySq(ctx, query)
		if err != nil {
			return nil, err
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
		var p models.Payment
		for rows.Next() {
			err = rows.Scan(&id, &p.Transaction, &p.RequestId, &p.Currency, &p.Provider, &p.Amount, &p.PaymentDt, &p.Bank, &p.DeliveryCost, &p.GoodsTotal, &p.CustomFee)
			if err != nil {
				return nil, err
			}
		}
		rows.Close()

		items := make([]models.Item, 0)
		for _, itemId := range itemsIDs {
			query = sq.Select("*").From("items").Where(sq.Eq{"item_id": itemId}).PlaceholderFormat(sq.Dollar)
			rows, err = r.QueryManager.QuerySq(ctx, query)
			if err != nil {
				return nil, err
			}
			if err = rows.Err(); err != nil {
				return nil, err
			}
			var i models.Item
			for rows.Next() {
				err = rows.Scan(&id, &i.ChrtId, &i.TrackNumber, &i.Price, &i.RId, &i.Name, &i.Sale, &i.Size, &i.TotalPrice, &i.NmId, &i.Brand, &i.Status)
				if err != nil {
					return nil, err
				}
			}
			items = append(items, i)
			rows.Close()
		}
		ord := models.Order{
			OrderId:           order.OrderId,
			TrackNumber:       order.TrackNumber,
			Entry:             order.Entry,
			Delivery:          d,
			Payment:           p,
			Items:             items,
			Locale:            order.Locale,
			InternalSignature: order.InternalSignature,
			CustomerId:        order.CustomerId,
			DeliveryService:   order.DeliveryService,
			Shardkey:          order.Shardkey,
			SmId:              order.SmId,
			DateCreated:       order.DateCreated,
			OofShard:          order.OofShard,
		}
		orders = append(orders, ord)
	}
	return orders, nil
}
