package repositorywallet

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	domainwallet "3za-digital/internal/domain/wallet"
	interfacewallet "3za-digital/internal/interfaces/wallet"
	repositorygeneric "3za-digital/internal/repositories/generic"
	"3za-digital/pkg/filter"
	"3za-digital/pkg/money"
	"3za-digital/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type repo struct {
	db *gorm.DB
}

var userSummaryColumns = []string{"id", "name", "email", "phone", "role", "avatar_url"}

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
	return repositorygeneric.New[domainwallet.WalletTransaction](r.db).GetAll(ctx, params, repositorygeneric.QueryOptions{
		BaseQuery: func(query *gorm.DB) *gorm.DB {
			return query.Where("user_id = ?", userID)
		},
		AllowedFilters:      []string{"type", "direction", "reference"},
		FilterSanitizer:     filter.WhitelistStringFilter,
		AllowedOrderColumns: []string{"created_at", "amount", "type", "direction"},
		DefaultOrders:       []string{"created_at DESC"},
	})
}

func (r *repo) GetWallets(ctx context.Context, params filter.BaseParams) ([]domainwallet.Wallet, int64, error) {
	return repositorygeneric.New[domainwallet.Wallet](r.db).GetAll(ctx, params, repositorygeneric.QueryOptions{
		BaseQuery: func(query *gorm.DB) *gorm.DB {
			return query.Preload("User", func(db *gorm.DB) *gorm.DB {
				return db.Select(userSummaryColumns)
			})
		},
		AllowedFilters:      []string{"user_id", "currency", "is_active"},
		FilterSanitizer:     filter.WhitelistStringFilter,
		AllowedOrderColumns: []string{"created_at", "updated_at", "balance", "user_id"},
		DefaultOrders:       []string{"created_at DESC"},
	})
}

func (r *repo) GetDeposits(ctx context.Context, userID string, params filter.BaseParams) ([]domainwallet.DepositRequest, int64, error) {
	return repositorygeneric.New[domainwallet.DepositRequest](r.db).GetAll(ctx, params, repositorygeneric.QueryOptions{
		BaseQuery: func(query *gorm.DB) *gorm.DB {
			if strings.TrimSpace(userID) != "" {
				return query.Where("user_id = ?", userID)
			}
			return query.Preload("User", func(db *gorm.DB) *gorm.DB {
				return db.Select(userSummaryColumns)
			})
		},
		AllowedFilters:      []string{"user_id", "status", "method", "provider", "payment_reference"},
		FilterSanitizer:     filter.WhitelistStringFilter,
		AllowedOrderColumns: []string{"created_at", "updated_at", "amount", "status"},
		DefaultOrders:       []string{"created_at DESC"},
	})
}

func (r *repo) GetDepositByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	var deposit domainwallet.DepositRequest
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&deposit).Error
	return deposit, err
}

func (r *repo) GetDepositWithUserByID(ctx context.Context, id string) (domainwallet.DepositRequest, error) {
	var deposit domainwallet.DepositRequest
	err := r.db.WithContext(ctx).
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select(userSummaryColumns)
		}).
		Where("id = ?", id).
		First(&deposit).Error
	return deposit, err
}

func (r *repo) CreateDepositRequest(ctx context.Context, deposit domainwallet.DepositRequest) (domainwallet.DepositRequest, error) {
	if deposit.Id == "" {
		deposit.Id = utils.CreateUUID()
	}
	if deposit.Status == "" {
		deposit.Status = domainwallet.DepositStatusPending
	}
	if deposit.Method == "" {
		deposit.Method = domainwallet.DepositMethodManualAdmin
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

func (r *repo) CreateManualTopup(ctx context.Context, deposit domainwallet.DepositRequest, description string, mainBalanceLimit string) (domainwallet.DepositRequest, error) {
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
		if _, err := r.ensureWalletTx(tx.WithContext(ctx), deposit.UserID); err != nil {
			return err
		}
		if err := r.assertMainBalanceLimitTx(tx.WithContext(ctx), deposit.Amount, mainBalanceLimit); err != nil {
			return err
		}
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

func (r *repo) ApproveManualTopup(ctx context.Context, deposit domainwallet.DepositRequest, description string, mainBalanceLimit string) (domainwallet.DepositRequest, error) {
	now := time.Now()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing domainwallet.DepositRequest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", deposit.Id).First(&existing).Error; err != nil {
			return err
		}
		if existing.UserID != deposit.UserID || money.NormalizeOrTrim(existing.Amount) != money.NormalizeOrTrim(deposit.Amount) {
			return domainwallet.ErrDepositAmountMismatch
		}
		if existing.Status != domainwallet.DepositStatusPending {
			return domainwallet.ErrDepositAlreadyFinal
		}
		if _, err := r.ensureWalletTx(tx.WithContext(ctx), existing.UserID); err != nil {
			return err
		}
		if err := r.assertMainBalanceLimitTx(tx.WithContext(ctx), existing.Amount, mainBalanceLimit); err != nil {
			return err
		}

		existing.Status = domainwallet.DepositStatusPaid
		existing.PaidAt = &now
		existing.UpdatedAt = &now
		existing.CreatedBy = deposit.CreatedBy
		if len(existing.Metadata) == 0 {
			existing.Metadata = utils.EmptyJSON()
		}
		if err := tx.Save(&existing).Error; err != nil {
			return err
		}
		deposit = existing
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

func (r *repo) CancelDeposit(ctx context.Context, depositID string, actorID string, reason string) (domainwallet.DepositRequest, error) {
	var deposit domainwallet.DepositRequest
	now := time.Now()
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", depositID).First(&deposit).Error; err != nil {
			return err
		}
		if deposit.Status != domainwallet.DepositStatusPending {
			return domainwallet.ErrDepositAlreadyFinal
		}

		deposit.Status = domainwallet.DepositStatusCancelled
		deposit.UpdatedAt = &now
		if strings.TrimSpace(actorID) != "" {
			deposit.CreatedBy = &actorID
		}
		deposit.Metadata = mergeDepositMetadataJSON(deposit.Metadata, map[string]string{
			"cancelled_by":  actorID,
			"cancelled_at":  now.Format(time.RFC3339),
			"cancel_reason": strings.TrimSpace(reason),
		})
		return tx.Save(&deposit).Error
	})
	return deposit, err
}

func (r *repo) AdjustWallet(ctx context.Context, userID, direction, amount, description, createdBy string, mainBalanceLimit string) (domainwallet.WalletTransaction, error) {
	var ret domainwallet.WalletTransaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var createdByPtr *string
		if strings.TrimSpace(createdBy) != "" {
			createdByPtr = &createdBy
		}
		if direction == domainwallet.DirectionCredit {
			if _, err := r.ensureWalletTx(tx.WithContext(ctx), userID); err != nil {
				return err
			}
			if err := r.assertMainBalanceLimitTx(tx.WithContext(ctx), amount, mainBalanceLimit); err != nil {
				return err
			}
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
		if money.NormalizeOrTrim(paymentExpectedAmount(deposit)) != money.NormalizeOrTrim(amount) {
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

		now := time.Now()
		deposit.Status = status
		deposit.UpdatedAt = &now
		return tx.Save(&deposit).Error
	})
	return deposit, err
}

func mergeDepositMetadataJSON(raw json.RawMessage, values map[string]string) json.RawMessage {
	metadata := map[string]string{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &metadata)
	}
	for key, value := range values {
		metadata[key] = value
	}
	return utils.MustJSON(metadata)
}

func paymentExpectedAmount(deposit domainwallet.DepositRequest) string {
	if len(deposit.Metadata) == 0 {
		return deposit.Amount
	}
	var metadata map[string]string
	if err := json.Unmarshal(deposit.Metadata, &metadata); err != nil {
		return deposit.Amount
	}
	if payableAmount := strings.TrimSpace(metadata["payable_amount"]); payableAmount != "" {
		return payableAmount
	}
	return deposit.Amount
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

func (r *repo) assertMainBalanceLimitTx(tx *gorm.DB, creditAmount string, mainBalanceLimit string) error {
	if strings.TrimSpace(mainBalanceLimit) == "" {
		return nil
	}
	amountCents, err := money.ParseCents(creditAmount)
	if err != nil || amountCents <= 0 {
		return domainwallet.ErrInvalidAmount
	}
	limitCents, err := money.ParseCents(mainBalanceLimit)
	if err != nil || limitCents < 0 {
		return domainwallet.ErrMainBalanceUnavailable
	}

	var wallets []domainwallet.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("is_active = ?", true).Find(&wallets).Error; err != nil {
		return err
	}
	totalCents := int64(0)
	for _, wallet := range wallets {
		walletCents, err := money.ParseCents(wallet.Balance)
		if err != nil {
			return err
		}
		lockedCents, err := money.ParseCents(wallet.LockedBalance)
		if err != nil {
			return err
		}
		totalCents += walletCents + lockedCents
	}
	if totalCents+amountCents > limitCents {
		return domainwallet.ErrInsufficientMainBalance
	}
	return nil
}

var _ interfacewallet.RepoWalletInterface = (*repo)(nil)
