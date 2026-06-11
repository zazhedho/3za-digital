package repositoryorder

import (
	"context"
	"strings"
	"time"

	domaincatalog "3za-digital/internal/domain/catalog"
	domainorder "3za-digital/internal/domain/order"
	domainwallet "3za-digital/internal/domain/wallet"
	interfaceorder "3za-digital/internal/interfaces/order"
	repositorygeneric "3za-digital/internal/repositories/generic"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/money"
	"3za-digital/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (r *repo) GetByID(ctx context.Context, id string) (domainorder.Order, error) {
	var order domainorder.Order
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&order).
		Error
	if err != nil {
		return domainorder.Order{}, err
	}
	if order.ServiceID == nil || *order.ServiceID == "" {
		return order, nil
	}
	var service domaincatalog.ProviderService
	if serviceErr := r.db.WithContext(ctx).
		Select("name", "category", "platform").
		Where("id = ?", *order.ServiceID).
		First(&service).Error; serviceErr == nil {
		order.ServiceName = service.Name
		order.ServiceCategory = service.Category
		order.ServicePlatform = service.Platform
	}
	return order, nil
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

func (r *repo) CreateWithStatusLogAndWalletDebit(ctx context.Context, order domainorder.Order, log domainorder.OrderStatusLog, userID string) (domainorder.Order, error) {
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

		return mutateWallet(tx, walletMutation{
			UserID:      userID,
			OrderID:     &order.Id,
			Type:        domainwallet.TransactionTypeDebitOrder,
			Direction:   domainwallet.DirectionDebit,
			Amount:      order.Amount,
			Reference:   "debit_order:" + order.Id,
			Description: "order debit " + order.RefID,
			CreatedBy:   order.CreatedBy,
		})
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

func (r *repo) RefundWalletForOrder(ctx context.Context, order domainorder.Order, amount string, description string) (bool, error) {
	refunded := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		reference := "refund_order:" + order.Id
		if err := tx.Model(&domainwallet.WalletTransaction{}).Where("reference = ?", reference).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		userID := ""
		if order.CreatedBy != nil {
			userID = *order.CreatedBy
		}
		if err := mutateWallet(tx, walletMutation{
			UserID:      userID,
			OrderID:     &order.Id,
			Type:        domainwallet.TransactionTypeRefundOrder,
			Direction:   domainwallet.DirectionCredit,
			Amount:      amount,
			Reference:   reference,
			Description: strings.TrimSpace(description),
			CreatedBy:   order.CreatedBy,
		}); err != nil {
			return err
		}
		refunded = true
		return nil
	})
	return refunded, err
}

type walletMutation struct {
	UserID      string
	OrderID     *string
	Type        string
	Direction   string
	Amount      string
	Reference   string
	Description string
	CreatedBy   *string
}

func mutateWallet(tx *gorm.DB, req walletMutation) error {
	if strings.TrimSpace(req.UserID) == "" {
		return domainwallet.ErrInvalidAmount
	}
	wallet, err := ensureWallet(tx, req.UserID)
	if err != nil {
		return err
	}
	if !wallet.IsActive {
		return domainwallet.ErrInactiveWallet
	}

	var locked domainwallet.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", wallet.Id).First(&locked).Error; err != nil {
		return err
	}

	current, err := money.ParseCents(locked.Balance)
	if err != nil {
		return err
	}
	amount, err := money.ParseCents(req.Amount)
	if err != nil || amount <= 0 {
		return domainwallet.ErrInvalidAmount
	}

	var next int64
	switch req.Direction {
	case domainwallet.DirectionCredit:
		next = current + amount
	case domainwallet.DirectionDebit:
		if current < amount {
			return domainwallet.ErrInsufficientBalance
		}
		next = current - amount
	default:
		return domainwallet.ErrInvalidDirection
	}

	locked.Balance = money.FormatCents(next)
	locked.UpdatedAt = new(time.Now())
	if err := tx.Save(&locked).Error; err != nil {
		return err
	}

	walletTx := domainwallet.WalletTransaction{
		Id:            utils.CreateUUID(),
		WalletID:      locked.Id,
		UserID:        locked.UserID,
		OrderID:       req.OrderID,
		Type:          req.Type,
		Direction:     req.Direction,
		Amount:        money.FormatCents(amount),
		BalanceBefore: money.FormatCents(current),
		BalanceAfter:  money.FormatCents(next),
		Reference:     strings.TrimSpace(req.Reference),
		Description:   strings.TrimSpace(req.Description),
		Metadata:      utils.EmptyJSON(),
		CreatedBy:     req.CreatedBy,
	}
	return tx.Create(&walletTx).Error
}

func ensureWallet(tx *gorm.DB, userID string) (domainwallet.Wallet, error) {
	wallet := domainwallet.Wallet{
		Id:            utils.CreateUUID(),
		UserID:        userID,
		Balance:       "0",
		LockedBalance: "0",
		Currency:      "IDR",
		IsActive:      true,
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoNothing: true,
	}).Create(&wallet).Error; err != nil {
		return domainwallet.Wallet{}, err
	}
	// Query into a fresh struct. The insert candidate has a generated primary key,
	// and GORM would include it in First(&wallet), hiding existing rows after conflict.
	var existing domainwallet.Wallet
	if err := tx.Where("user_id = ?", userID).First(&existing).Error; err != nil {
		return domainwallet.Wallet{}, err
	}
	return existing, nil
}
