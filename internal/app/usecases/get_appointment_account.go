package usecases

import (
	"context"
	"fmt"

	"dental-scheduler-backend/internal/domain/entities"
	"dental-scheduler-backend/internal/domain/ports/repositories"
	"dental-scheduler-backend/internal/domain/services"

	"github.com/google/uuid"
)

// GetAppointmentAccountUseCase handles retrieving appointment account information
type GetAppointmentAccountUseCase struct {
	accountingService *services.AccountingService
}

// NewGetAppointmentAccountUseCase creates a new instance
func NewGetAppointmentAccountUseCase(
	accountingService *services.AccountingService,
) *GetAppointmentAccountUseCase {
	return &GetAppointmentAccountUseCase{
		accountingService: accountingService,
	}
}

// GetAccountWithEntriesOutput contains account and entries
type GetAccountWithEntriesOutput struct {
	Account *entities.AppointmentAccount
	Entries []*entities.AppointmentAccountEntry
	Balance *repositories.AccountBalance
}

// Execute retrieves an appointment account with all its entries and balance
func (uc *GetAppointmentAccountUseCase) Execute(ctx context.Context, organizationID, appointmentID string) (*GetAccountWithEntriesOutput, error) {
	// Parse UUIDs
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	apptID, err := uuid.Parse(appointmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid appointment ID: %w", err)
	}

	// Get account
	account, err := uc.accountingService.CreateOrGetAccount(ctx, orgID, apptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Get entries
	entries, err := uc.accountingService.GetEntriesForAccount(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries: %w", err)
	}

	// Calculate balance
	balance, err := uc.accountingService.CalculateBalance(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate balance: %w", err)
	}

	return &GetAccountWithEntriesOutput{
		Account: account,
		Entries: entries,
		Balance: balance,
	}, nil
}

// GetBalanceOnly retrieves just the balance for an appointment
func (uc *GetAppointmentAccountUseCase) GetBalanceOnly(ctx context.Context, organizationID, appointmentID string) (*repositories.AccountBalance, error) {
	// Parse UUIDs
	orgID, err := uuid.Parse(organizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	apptID, err := uuid.Parse(appointmentID)
	if err != nil {
		return nil, fmt.Errorf("invalid appointment ID: %w", err)
	}

	// Get or create account
	account, err := uc.accountingService.CreateOrGetAccount(ctx, orgID, apptID)
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// Calculate balance
	balance, err := uc.accountingService.CalculateBalance(ctx, account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate balance: %w", err)
	}

	return balance, nil
}
