package repository

import (
	"context"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
	"wb-tech-backend/internal/core"
	"wb-tech-backend/internal/models"
	"wb-tech-backend/internal/pkg/pgdb"
)

type Deps struct {
	Config             *core.Config
	QueryManager       *pgdb.QueryManager
	TransactionManager *pgdb.TransactionManager
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
			Config:             cfg,
			QueryManager:       qm,
			TransactionManager: tm,
		},
	}
	return r, nil
}

func (r *Repository) AddOrder(ctx context.Context, order models.Order) error {
	return r.TransactionManager.Tx(ctx, func(ctx context.Context) error {
		query := sq.Insert("deliveries").
			Columns("name", "phone", "zip", "city", "address", "region", "email").
			Values(order.Delivery.Name, &order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email).
			PlaceholderFormat(sq.Dollar).Suffix("RETURNING delivery_id")
		rows, err := r.QueryManager.QuerySq(ctx, query)
		if err != nil {
			return err
		}
		if err = rows.Err(); err != nil {
			return err
		}
		var deliveryId int64
		err = rows.Scan(&deliveryId)
		if err != nil {
			return err
		}
		rows.Close()
		query = sq.Insert("payments").
			Columns("transaction", "request_id", "currency", "provider", "amount", "payment_dt", "bank", "delivery_cost", "goods_total", "custom_fee").
			Values(order.Payment.Transaction, order.Payment.RequestId, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee).
			PlaceholderFormat(sq.Dollar).Suffix("RETURNING payment_id")
		rows, err = r.QueryManager.QuerySq(ctx, query)
		if err != nil {
			return err
		}
		if err = rows.Err(); err != nil {
			return err
		}
		var paymentId int64
		err = rows.Scan(&paymentId)
		if err != nil {
			return err
		}
		rows.Close()
		itemsIds := make([]int64, 0)
		for _, item := range order.Items {
			query = sq.Insert("items").
				Columns("chrt_id", "track_number", "price", "rid", "name", "sale", "size", "total_price", "nm_id", "brand", "status").
				Values(item.ChrtId, item.TrackNumber, item.Price, item.RId, item.Name, item.Sale, item.Size, item.TotalPrice, item.NmId, item.Brand, item.Status).
				PlaceholderFormat(sq.Dollar).Suffix("RETURNING item_id")
			rows, err = r.QueryManager.QuerySq(ctx, query)
			if err != nil {
				return err
			}
			if err = rows.Err(); err != nil {
				return err
			}
			var itemId int64
			err = rows.Scan(&itemId)
			if err != nil {
				return err
			}
			rows.Close()
			itemsIds = append(itemsIds, itemId)
		}
		query = sq.Insert("orders").
			Columns("order_uid", "track_number", "entry", "delivery_id", "payment_id", "items_ids", "locale", "internal_signature", "customer_id", "delivery_service", "shardkey", "sm_id", "date_created", "oof_shard").
			Values(order.OrderId, order.TrackNumber, order.Entry, deliveryId, paymentId, itemsIds, order.Locale, order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, order.DateCreated, order.OofShard).
			PlaceholderFormat(sq.Dollar).Suffix("RETURNING order_uid")
		rows, err = r.QueryManager.QuerySq(ctx, query)
		if err != nil {
			return err
		}
		if err = rows.Err(); err != nil {
			return err
		}
		defer rows.Close()
		var orderId string
		err = rows.Scan(&orderId)
		if err != nil {
			return err
		}
		if orderId == order.OrderId {
			return nil
		}
		return fmt.Errorf("something goes wrong with add order to database")
	})
}

func (r *Repository) GetOrderById(ctx context.Context, orderId string) (models.Order, error) {
	query := sq.Select("o.order_uid", "o.track_number", "o.entry", "o.locale", "o.internal_signature", "o.customer_id", "o.delivery_service", "o.shardkey", "o.sm_id", "o.date_created", "o.oof_shard",
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
	err = rows.Scan(
		&order.OrderId, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId, &order.DateCreated,
		&order.OofShard, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
		&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email,
		&order.Payment.Transaction, &order.Payment.RequestId, &order.Payment.Currency, &order.Payment.Provider,
		&order.Payment.Amount, &order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	rows.Close()
	if err != nil {
		return models.Order{}, err
	}
	query = sq.Select("i.chrt_id", "i.track_number", "i.price", "i.rid", "i.name", "i.sale", "i.size", "i.total_price", "i.nm_id", "i.brand", "i.status").
		From("items i").Join("orders o ON o.item_id = i.item_id").Where(sq.Eq{"o.order_uid": orderId}).PlaceholderFormat(sq.Dollar)
	rows, err = r.QueryManager.QuerySq(ctx, query)
	if err != nil {
		return models.Order{}, err
	}
	if err = rows.Err(); err != nil {
		return models.Order{}, err
	}
	defer rows.Close()

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
	return order, nil
}
func (r *Repository) GetOrders(ctx context.Context) ([]models.Order, error) {
	query := sq.Select("o.order_uid", "o.track_number", "o.entry", "o.locale", "o.internal_signature", "o.customer_id", "o.delivery_service", "o.shardkey", "o.sm_id", "o.date_created", "o.oof_shard",
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
	for rows.Next() {
		var order models.Order
		err = rows.Scan(
			&order.OrderId, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
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
		query = sq.Select("i.chrt_id", "i.track_number", "i.price", "i.rid", "i.name", "i.sale", "i.size", "i.total_price", "i.nm_id", "i.brand", "i.status").
			From("items i").Join("orders o ON o.item_id = i.item_id").Where(sq.Eq{"o.order_uid": order.OrderId}).PlaceholderFormat(sq.Dollar)
		rows, err = r.QueryManager.QuerySq(ctx, query)
		if err != nil {
			return nil, err
		}
		if err = rows.Err(); err != nil {
			return nil, err
		}
		defer rows.Close()

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
	}
	return orders, nil
}
