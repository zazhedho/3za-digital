package repositorywallet

import (
	"context"
	"fmt"
	"strings"
	"time"

	domainwallet "3za-digital/internal/domain/wallet"
	interfacewallet "3za-digital/internal/interfaces/wallet"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/money"
	"3za-digital/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repo struct {
	db *gorm.DB
}

func NewWalletRepo(db *gorm.DB) interfacewallet.RepoWalletInterface {
	return &repo{db: db}
}

func (r *repo) GetWalletByUserID(ctx context.Context, userID string) (domainwallet.Wallet, error) {
	var wallet domainwallet.Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&wallet).Error
	return wallet, err
}

func (r *repo) EnsureWallet(ctx context.Context, userID string) (domainwallet.Wallet, error) {
	var wallet domainwallet.Wallet
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		wallet, err = r.ensureWalletTx(tx.WithContext(ctx), userID)
		return err
	})
	return wallet, err
}

func (r *repo) GetTransactions(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.WalletTransaction, int64, error) {
	query := r.db.WithContext(ctx).Model(&domainwallet.WalletTransaction{}).Where("user_id = ?", userID)
	query = applyStringFilters(query, filter.WhitelistStringFilter(params.Filters, []string{"type", "direction", "reference"}))
	return getPaged[domainwallet.WalletTransaction](query, params, map[string]bool{
		"created_at": true,
		"amount":     true,
		"type":       true,
		"direction":  true,
	}, "created_at DESC")
}

func (r *repo) GetWallets(ctx context.Context, params filter.BaseParams) ([]domainwallet.Wallet, int64, error) {
	query := r.db.WithContext(ctx).Model(&domainwallet.Wallet{})
	query = applyStringFilters(query, filter.WhitelistStringFilter(params.Filters, []string{"user_id", "currency", "is_active"}))
	return getPaged[domainwallet.Wallet](query, params, map[string]bool{
		"created_at": true,
		"updated_at": true,
		"balance":    true,
		"user_id":    true,
	}, "created_at DESC")
}

func (r *repo) GetDeposits(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	query := r.db.WithContext(ctx).Model(&domainwallet.DepositRequest{})
	if strings.TrimSpace(userID) != "" {
		query = query.Where("user_id = ?", userID)
	}
	query = applyStringFilters(query, filter.WhitelistStringFilter(params.Filters, []string{"status", "method", "provider", "payment_reference"}))
	return getPaged[domainwallet.DepositRequest](query, params, map[string]bool{
		"created_at": true,
		"updated_at": true,
		"amount":     true,
		"status":     true,
	}, "created_at DESC")
}

func (r *repo) GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	var deposit domainwallet.DepositRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&deposit).Error
	return deposit, err
}

func (r *repo) CreatePaymentGatewayDeposit(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	if deposit.Id == "" {
		deposit.Id = utils.CreateUUID()
	}
	if deposit.Status == "" {
		deposit.Status = domainwallet.DepositStatusPending
	}
	if deposit.Method == "" {
		deposit.Method = domainwallet.DepositMethodPaymentGateway
	}
	if len(deposit.Metadata) == 0 {
		deposit.Metadata = utils.EmptyJSON()
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := r.ensureWalletTx(tx.WithContext(ctx), deposit.UserID); err != nil {
			return err
		}
		return tx.Create(&deposit).Error
	})
	return deposit, err
}

func (r *repo) CreateManualTopup(ctx context.Context, deposit domainwallet.DepositRequest, description string) (domainwallet.DepositRequest, error) {
	if deposit.Id == "" {
		deposit.Id = utils.CreateUUID()
	}
	now := time.Now()
	deposit.Status = domainwallet.DepositStatusPaid
	deposit.Method = domainwallet.DepositMethodManualAdmin
	deposit.PaidAt = &now
	deposit.UpdatedAt = &now
	if len(deposit.Metadata) == 0 {
		deposit.Metadata = utils.EmptyJSON()
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&deposit).Error; err != nil {
			return err
		}
		_, err := r.mutateWalletTx(tx.WithContext(ctx), walletMutation{
			UserID:           deposit.UserID,
			DepositRequestID: &deposit.Id,
			Type:             domainwallet.TransactionTypeDeposit,
			Direction:        domainwallet.DirectionCredit,
			Amount:           deposit.Amount,
			Reference:        "deposit:" + deposit.Id,
			Description:      description,
			CreatedBy:        deposit.CreatedBy,
		})
		return err
	})
	return deposit, err
}

func (r *repo) AdjustWallet(ctx context.Context, userID, direction, amount, description, createdBy string) (domainwallet.WalletTransaction, error) {
	var ret domainwallet.WalletTransaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var createdByPtr *string
		if strings.TrimSpace(createdBy) != "" {
			createdByPtr = &createdBy
		}
		var err error
		ret, err = r.mutateWalletTx(tx.WithContext(ctx), walletMutation{
			UserID:      userID,
			Type:        domainwallet.TransactionTypeAdjustment,
			Direction:   direction,
			Amount:      amount,
			Reference:   fmt.Sprintf("adjustment:%s", utils.CreateUUID()),
			Description: description,
			CreatedBy:   createdByPtr,
		})
		return err
	})
	return ret, err
}

func (r *repo) CompleteDepositByPaymentReference(ctx context.Context, provider, paymentReference, amount string, log domainwallet.PaymentGatewayLog) (domainwallet.DepositRequest, error) {
	var deposit domainwallet.DepositRequest
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(log.Payload) == 0 {
			log.Payload = utils.EmptyJSON()
		}
		if log.Id == "" {
			log.Id = utils.CreateUUID()
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("provider = ? AND payment_reference = ?", provider, paymentReference).
			First(&deposit).Error; err != nil {
			return err
		}
		if money.NormalizeOrTrim(deposit.Amount) != money.NormalizeOrTrim(amount) {
			return domainwallet.ErrDepositAmountMismatch
		}
		if deposit.Status == domainwallet.DepositStatusPaid {
			return nil
		}
		if deposit.Status != domainwallet.DepositStatusPending {
			return domainwallet.ErrDepositAlreadyFinal
		}

		now := time.Now()
		deposit.Status = domainwallet.DepositStatusPaid
		deposit.PaidAt = &now
		deposit.UpdatedAt = &now
		if err := tx.Save(&deposit).Error; err != nil {
			return err
		}
		log.DepositRequestID = &deposit.Id
		if err := tx.Model(&domainwallet.PaymentGatewayLog{}).Where("id = ?", log.Id).Update("deposit_request_id", deposit.Id).Error; err != nil {
			return err
		}
		_, err := r.mutateWalletTx(tx.WithContext(ctx), walletMutation{
			UserID:           deposit.UserID,
			DepositRequestID: &deposit.Id,
			Type:             domainwallet.TransactionTypeDeposit,
			Direction:        domainwallet.DirectionCredit,
			Amount:           deposit.Amount,
			Reference:        "deposit:" + deposit.Id,
			Description:      "payment gateway deposit",
		})
		return err
	})
	return deposit, err
}

func (r *repo) UpdateDepositStatusByPaymentReference(ctx context.Context, provider, paymentReference, status string, log domainwallet.PaymentGatewayLog) (domainwallet.DepositRequest, error) {
	var deposit domainwallet.DepositRequest
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(log.Payload) == 0 {
			log.Payload = utils.EmptyJSON()
		}
		if log.Id == "" {
			log.Id = utils.CreateUUID()
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("provider = ? AND payment_reference = ?", provider, paymentReference).
			First(&deposit).Error; err != nil {
			return err
		}

		log.DepositRequestID = &deposit.Id
		if err := tx.Model(&domainwallet.PaymentGatewayLog{}).Where("id = ?", log.Id).Update("deposit_request_id", deposit.Id).Error; err != nil {
			return err
		}
		if deposit.Status == domainwallet.DepositStatusPaid {
			return nil
		}
		if deposit.Status != domainwallet.DepositStatusPending {
			return nil
		}

		deposit.Status = status
		deposit.UpdatedAt = new(time.Now())
		return tx.Save(&deposit).Error
	})
	return deposit, err
}

type walletMutation struct {
	UserID           string
	OrderID          *string
	DepositRequestID *string
	Type             string
	Direction        string
	Amount           string
	Reference        string
	Description      string
	CreatedBy        *string
}

func (r *repo) mutateWalletTx(tx *gorm.DB, req walletMutation) (domainwallet.WalletTransaction, error) {
	wallet, err := r.ensureWalletTx(tx, req.UserID)
	if err != nil {
		return domainwallet.WalletTransaction{}, err
	}
	if !wallet.IsActive {
		return domainwallet.WalletTransaction{}, domainwallet.ErrInactiveWallet
	}

	var locked domainwallet.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", wallet.Id).First(&locked).Error; err != nil {
		return domainwallet.WalletTransaction{}, err
	}

	current, err := money.ParseCents(locked.Balance)
	if err != nil {
		return domainwallet.WalletTransaction{}, err
	}
	amount, err := money.ParseCents(req.Amount)
	if err != nil || amount <= 0 {
		return domainwallet.WalletTransaction{}, domainwallet.ErrInvalidAmount
	}

	var next int64
	switch req.Direction {
	case domainwallet.DirectionCredit:
		next = current + amount
	case domainwallet.DirectionDebit:
		if current < amount {
			return domainwallet.WalletTransaction{}, domainwallet.ErrInsufficientBalance
		}
		next = current - amount
	default:
		return domainwallet.WalletTransaction{}, domainwallet.ErrInvalidDirection
	}

	locked.Balance = money.FormatCents(next)
	locked.UpdatedAt = new(time.Now())
	if err := tx.Save(&locked).Error; err != nil {
		return domainwallet.WalletTransaction{}, err
	}

	walletTx := domainwallet.WalletTransaction{
		Id:               utils.CreateUUID(),
		WalletID:         locked.Id,
		UserID:           locked.UserID,
		OrderID:          req.OrderID,
		DepositRequestID: req.DepositRequestID,
		Type:             req.Type,
		Direction:        req.Direction,
		Amount:           money.FormatCents(amount),
		BalanceBefore:    money.FormatCents(current),
		BalanceAfter:     money.FormatCents(next),
		Reference:        strings.TrimSpace(req.Reference),
		Description:      strings.TrimSpace(req.Description),
		Metadata:         utils.EmptyJSON(),
		CreatedBy:        req.CreatedBy,
	}
	if err := tx.Create(&walletTx).Error; err != nil {
		return domainwallet.WalletTransaction{}, err
	}
	return walletTx, nil
}

func (r *repo) ensureWalletTx(tx *gorm.DB, userID string) (domainwallet.Wallet, error) {
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

func getPaged[T any](query *gorm.DB, params filter.BaseParams, allowedOrder map[string]bool, defaultOrder string) ([]T, int64, error) {
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := strings.TrimSpace(params.OrderBy)
	if !allowedOrder[orderBy] {
		orderBy = strings.Fields(defaultOrder)[0]
	}
	dir := utils.NormalizeUpperKey(params.OrderDirection)
	if dir != "ASC" && dir != "DESC" {
		dir = "DESC"
	}

	var ret []T
	err := query.Order(orderBy + " " + dir).Offset(params.Offset).Limit(params.Limit).Find(&ret).Error
	return ret, total, err
}

func applyStringFilters(query *gorm.DB, filters map[string]interface{}) *gorm.DB {
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}
	return query
}

var _ interfacewallet.RepoWalletInterface = (*repo)(nil)
