package repositoryorder

import (
	"context"

	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	interfaceorder "3za-digital/internal/interfaces/order"
	repositorygeneric "3za-digital/internal/repositories/generic"
	"3za-digital/pkg/filter"
	"3za-digital/utils"

	"gorm.io/gorm"
)

type repo struct {
	*repositorygeneric.GenericRepository[domainorder.Order]
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) interfaceorder.RepoOrderInterface {
	return &repo{
		GenericRepository: repositorygeneric.New[domainorder.Order](db),
		db:                db,
	}
}

func (r *repo) GetAll(ctx context.Context, params filter.BaseParams) ([]domainorder.Order, int64, error) {
	return r.GenericRepository.GetAll(ctx, params, repositorygeneric.QueryOptions{
		Search: repositorygeneric.BuildSearchFunc("ref_id", "provider_service_id", "target", "customer_no", "customer_name"),
		AllowedFilters: []string{
			"provider",
			"product_type",
			"ref_id",
			"provider_service_id",
			"status",
			"created_by",
		},
		FilterSanitizer: filter.WhitelistStringFilter,
		AllowedOrderColumns: []string{
			"ref_id",
			"status",
			"amount",
			"provider_charge",
			"profit",
			"quantity",
			"created_at",
			"updated_at",
		},
		DefaultOrders: []string{"created_at DESC"},
	})
}

func (r *repo) GetServiceByID(ctx context.Context, id string) (domaincatalog.ProviderService, error) {
	var service domaincatalog.ProviderService
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&service).Error
	return service, err
}

func (r *repo) GetStatusLogs(ctx context.Context, orderID string) ([]domainorder.OrderStatusLog, error) {
	var logs []domainorder.OrderStatusLog
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

func (r *repo) CreateWithStatusLog(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog) (domainorder.Order, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if order.Id == "" {
			order.Id = utils.CreateUUID()
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		if log.Id == "" {
			log.Id = utils.CreateUUID()
		}
		log.OrderID = order.Id
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return domainorder.Order{}, err
	}

	return order, nil
}

func (r *repo) UpdateWithStatusLog(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&order).Error; err != nil {
			return err
		}

		if log.NewStatus != "" {
			if log.Id == "" {
				log.Id = utils.CreateUUID()
			}
			log.OrderID = order.Id
			if err := tx.Create(&log).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
