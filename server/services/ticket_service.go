package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"gowitcase/db"
	"gowitcase/errors"
	"gowitcase/models"
	"log"
	"math"
	"strconv"
	"time"
)

type TicketService struct {
	DB    db.DatabaseInterface
	Cache db.RedisInterface
}

func NewTicketService(db db.DatabaseInterface, cache db.RedisInterface) *TicketService {
	return &TicketService{DB: db, Cache: cache}
}

func (s *TicketService) CreateTicket(ticket *models.Ticket) error {
	if ticket == nil {
		return fmt.Errorf("ticket is nil")
	}

	err := s.ValidateTicket(*ticket)
	if err != nil {
		return err
	}

	err = s.DB.QueryRow(
		"INSERT INTO ticket (name, description, allocation) VALUES ($1, $2, $3) RETURNING id",
		ticket.Name, ticket.Description, ticket.Allocation,
	).Scan(&ticket.ID)

	if err != nil {
		return err
	}

	return nil
}

func (s *TicketService) GetTicket(id int) (*models.Ticket, error) {
	ticket, err := s.getCacheTicket(id)
	if err == nil && ticket != nil {
		return ticket, nil
	}

	ticket = &models.Ticket{}

	err = s.DB.QueryRow(
		"SELECT id, name, description, allocation FROM ticket WHERE id = $1",
		id,
	).Scan(&ticket.ID, &ticket.Name, &ticket.Description, &ticket.Allocation)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewRestError(fmt.Sprintf("Ticket %d not found", id), 404)
		}
		return nil, err
	}

	err = s.cacheTicket(ticket)
	if err != nil {
		log.Printf("Failed to cache ticket: %v", err)
	}

	return ticket, nil
}

func (s *TicketService) PurchaseTicket(ticketID int, quantity int) error {

	if quantity <= 0 || quantity > math.MaxInt32 {
		return errors.NewRestError("Quantity must be a positive number within the valid range", 400)
	}

	tx, err := s.DB.BeginTransaction()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return fmt.Errorf("failed to start transaction: %v", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	ticket := &models.Ticket{}
	err = tx.QueryRow(
		"SELECT id, name, description, allocation FROM ticket WHERE id = $1 FOR UPDATE",
		ticketID,
	).Scan(&ticket.ID, &ticket.Name, &ticket.Description, &ticket.Allocation)

	if err != nil {
		if err == sql.ErrNoRows {
			err = errors.NewRestError(fmt.Sprintf("Ticket %d not found", ticketID), 404)
			return err
		}

		return fmt.Errorf("failed to get ticket: %v", err)
	}

	if ticket.Allocation == 0 {
		err = errors.NewRestError("Ticket is sold out", 400)
		return err
	}

	if ticket.Allocation < quantity {
		err = errors.NewRestError("Not enough tickets available", 400)
		return err
	}

	ticket.Allocation -= quantity

	_, err = tx.Exec(
		"UPDATE ticket SET allocation = $1 WHERE id = $2",
		ticket.Allocation, ticket.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update ticket: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	err = s.invalidateCache(ticketID)
	if err != nil {
		log.Printf("Failed to invalidate cache: %v for ticker: %d", err, ticketID)
	}

	return nil
}

func (s *TicketService) ValidateTicket(ticket models.Ticket) error {
	if ticket.Name == "" {
		return errors.NewRestError("Field 'name' is required", 400)
	}

	if len(ticket.Name) > 255 {
		return errors.NewRestError("Field 'name' must be less than 255 characters", 400)
	}

	if ticket.Allocation <= 0 {
		return errors.NewRestError("Field 'allocation' must be greater than 0", 400)
	}

	if ticket.Allocation > math.MaxInt32 {
		return errors.NewRestError("Allocation is too large", 400)
	}

	return nil
}

func (s *TicketService) cacheTicket(ticket *models.Ticket) error {
	ticketBytes, err := json.Marshal(ticket)
	if err != nil {
		return fmt.Errorf("failed to marshal ticket: %v", err)
	}
	return s.Cache.Set(s.getCacheKey(ticket.ID), string(ticketBytes), 5*time.Minute)
}

func (s *TicketService) invalidateCache(ticketID int) error {
	return s.Cache.Del(s.getCacheKey(ticketID))
}

func (s *TicketService) getCacheTicket(ticketID int) (*models.Ticket, error) {
	ticket := &models.Ticket{}
	ticketJSON, err := s.Cache.Get(s.getCacheKey(ticketID))
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(ticketJSON), &ticket)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (s *TicketService) getCacheKey(ticketID int) string {
	return models.TicketCachePrefix + strconv.Itoa(ticketID)
}
