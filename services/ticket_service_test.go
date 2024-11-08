package services_test

import (
	"database/sql"
	"gowitcase/mocks"
	"gowitcase/models"
	"gowitcase/services"
	"log"
	"math"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func setupTest(t *testing.T) (*services.TicketService, sqlmock.Sqlmock) {
	mockDB, mock, err := mocks.NewMockDatabase()
	assert.NoError(t, err)

	mockRedis := mocks.NewMockRedis()

	ticketService := services.NewTicketService(mockDB, mockRedis)

	return ticketService, mock
}

func TestCreateTicket_Success(t *testing.T) {
	ticketService, mock := setupTest(t)

	mock.ExpectQuery("INSERT INTO ticket").
		WithArgs("test", "test", 100).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	ticket := &models.Ticket{
		Name:        "test",
		Description: "test",
		Allocation:  100,
	}

	err := ticketService.CreateTicket(ticket)
	assert.NoError(t, err, "failed to create ticket")

	assert.Equal(t, 1, ticket.ID, "expected ticket ID 1")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "unexpected error")
}

func TestCreateTicket_AllocationZero(t *testing.T) {
	ticketService, _ := setupTest(t)

	ticket := &models.Ticket{
		Name:        "zero allocation ticket",
		Description: "zero allocation ticket",
		Allocation:  0,
	}

	err := ticketService.CreateTicket(ticket)
	assert.Error(t, err, "expected error when allocation is zero")
}

func TestCreateTicket_MissingName(t *testing.T) {
	ticketService, _ := setupTest(t)

	ticket := &models.Ticket{
		Description: "Missing name field",
		Allocation:  50,
	}

	err := ticketService.CreateTicket(ticket)
	assert.Error(t, err, "expected error when name is missing")
}

func TestCreateTicket_MissingAllocation(t *testing.T) {
	ticketService, _ := setupTest(t)

	ticket := &models.Ticket{
		Name:        "Ticket with missing allocation",
		Description: "Missing allocation field",
	}

	err := ticketService.CreateTicket(ticket)
	assert.Error(t, err, "expected error when allocation is missing")
}

func TestCreateTicket_NegativeAllocation(t *testing.T) {
	ticketService, _ := setupTest(t)

	ticket := &models.Ticket{
		Name:        "Ticket with negative allocation",
		Description: "Negative allocation field",
		Allocation:  -10,
	}

	err := ticketService.CreateTicket(ticket)
	assert.Error(t, err, "expected error when allocation is negative")
}

func TestCreateTicket_NameTooLong(t *testing.T) {
	ticketService, _ := setupTest(t)

	longName := make([]byte, 300)
	for i := range longName {
		longName[i] = 'a'
	}

	ticket := &models.Ticket{
		Name:        string(longName),
		Description: "Ticket with long name",
		Allocation:  100,
	}

	err := ticketService.CreateTicket(ticket)
	if err != nil {
		log.Printf("Error: %v", err)
	}
	assert.Error(t, err, "expected error when name exceeds maximum length")
}

func TestCreateTicket_AllocationMaxInt(t *testing.T) {
	ticketService, mock := setupTest(t)

	maxInt := math.MaxInt32

	mock.ExpectQuery("INSERT INTO ticket").
		WithArgs("ticket max allocation", "ticket with max allocation", maxInt).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	ticket := &models.Ticket{
		Name:        "ticket max allocation",
		Description: "ticket with max allocation",
		Allocation:  maxInt,
	}

	err := ticketService.CreateTicket(ticket)
	assert.NoError(t, err, "failed to create ticket with maximum allocation")

	assert.Equal(t, 1, ticket.ID, "expected ticket ID 1")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "unexpected error")
}

func TestCreateTicket_ExcessivelyLargeAllocation(t *testing.T) {
	ticketService, _ := setupTest(t)

	excessiveAllocation := math.MaxInt64

	ticket := &models.Ticket{
		Name:        "Ticket with excessive allocation",
		Description: "Ticket with excessive allocation",
		Allocation:  int(excessiveAllocation),
	}

	err := ticketService.CreateTicket(ticket)
	assert.Error(t, err, "expected error when allocation is excessively large")
}

func TestCreateTicket_EmptyRequestBody(t *testing.T) {
	ticketService, _ := setupTest(t)

	var ticket *models.Ticket

	err := ticketService.CreateTicket(ticket)
	assert.Error(t, err, "expected error when ticket is nil")
}

func TestGetTicket_Success(t *testing.T) {
	ticketService, mock := setupTest(t)

	ticketID := 1
	ticket := &models.Ticket{
		ID:          ticketID,
		Name:        "test",
		Description: "test",
		Allocation:  100,
	}

	mock.ExpectQuery("SELECT").
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "allocation"}).AddRow(ticketID, "test", "test", 100))

	returnedTicket, err := ticketService.GetTicket(ticketID)
	assert.NoError(t, err, "failed to get ticket")

	assert.Equal(t, ticket, returnedTicket, "expected ticket to match")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "unexpected error")
}

func TestGetTicket_CacheHit(t *testing.T) {
	ticketService, mock := setupTest(t)

	ticket := &models.Ticket{
		ID:          1,
		Name:        "test",
		Description: "test",
		Allocation:  100,
	}

	mock.ExpectQuery("SELECT").
		WithArgs(ticket.ID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "allocation"}).
			AddRow(ticket.ID, "test", "test", 100))

	returnedTicket, err := ticketService.GetTicket(ticket.ID)
	assert.NoError(t, err, "failed to get ticket")
	assert.Equal(t, ticket, returnedTicket, "expected ticket to match")

	cachedTicket, err := ticketService.GetTicket(ticket.ID)
	assert.NoError(t, err, "failed to get ticket from cache")
	assert.Equal(t, ticket, cachedTicket, "expected ticket to match")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "unexpected error")
}

func TestGetTicket_NotFound(t *testing.T) {
	ticketService, mock := setupTest(t)

	ticketID := 1

	mock.ExpectQuery("SELECT").
		WithArgs(ticketID).
		WillReturnError(sql.ErrNoRows)

	_, err := ticketService.GetTicket(ticketID)
	assert.Error(t, err, "expected error when ticket not found")
}

func TestPurchaseTicket_Success(t *testing.T) {
	ticketService, mock := setupTest(t)

	ticketID := 1
	quantity := 5
	initialAllocation := 100

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT").
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "allocation"}).
			AddRow(ticketID, "test", "test", initialAllocation))

	mock.ExpectExec("UPDATE").
		WithArgs(initialAllocation-quantity, ticketID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := ticketService.PurchaseTicket(ticketID, quantity)
	assert.NoError(t, err, "failed to purchase ticket")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "unfulfilled expectations")
}

func TestPurchaseTicket_NotEnoughRemaining(t *testing.T) {
	ticketService, mock := setupTest(t)

	ticketID := 1
	quantity := 5
	initialAllocation := 2

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT").
		WithArgs(ticketID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "allocation"}).
			AddRow(ticketID, "test", "test", initialAllocation))

	mock.ExpectRollback()

	err := ticketService.PurchaseTicket(ticketID, quantity)
	assert.Error(t, err, "expected error when not enough tickets remaining")
	assert.Equal(t, "Not enough tickets available", err.Error(), "expected error message to match")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "unexpected error")
}

func TestPurchaseTicket_NotFound(t *testing.T) {
	ticketService, mock := setupTest(t)

	ticketID := 1
	quantity := 5

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT").
		WithArgs(ticketID).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	err := ticketService.PurchaseTicket(ticketID, quantity)
	assert.Error(t, err, "expected error when ticket not found")

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err, "unexpected error")
}

func TestPurchaseTicket_InvalidQuantity(t *testing.T) {
	ticketService, _ := setupTest(t)

	ticketID := 1
	quantity := -5

	err := ticketService.PurchaseTicket(ticketID, quantity)
	assert.Error(t, err, "expected error when quantity is negative")
}

func TestPurchaseTicket_ZeroQuantity(t *testing.T) {
	ticketService, _ := setupTest(t)

	ticketID := 1
	quantity := 0

	err := ticketService.PurchaseTicket(ticketID, quantity)
	assert.Error(t, err, "expected error when quantity is zero")
}
