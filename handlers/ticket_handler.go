package handlers

import (
	customErrors "gowitcase/errors"
	"gowitcase/models"
	"gowitcase/services"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TicketHandler struct {
	TicketService *services.TicketService
}

func NewTicketHandler(ticketService *services.TicketService) *TicketHandler {
	return &TicketHandler{TicketService: ticketService}
}

func (h *TicketHandler) CreateTicket(ctx *gin.Context) {

	ticket := &models.Ticket{}
	if err := ctx.ShouldBindJSON(ticket); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err := h.TicketService.CreateTicket(ticket)
	if err != nil {
		if restErr, ok := err.(customErrors.RestError); ok {
			ctx.JSON(restErr.Status, gin.H{"error": restErr.Message})
			return
		}

		log.Printf("Failed to create ticket: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	ctx.JSON(http.StatusCreated, ticket)
}

func (h *TicketHandler) GetTicket(ctx *gin.Context) {
	id := ctx.Param("id")
	ticketID, err := strconv.Atoi(id)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	ticket, err := h.TicketService.GetTicket(ticketID)
	if err != nil {
		if restErr, ok := err.(customErrors.RestError); ok {
			ctx.JSON(restErr.Status, gin.H{"error": restErr.Message})
			return
		}

		log.Printf("Failed to get ticket: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong, please try again."})
		return
	}

	ctx.JSON(200, ticket)
}

func (h *TicketHandler) PurchaseTicket(ctx *gin.Context) {
	ticketID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ticket ID"})
		return
	}

	purchaseRequest := &models.PurchaseRequest{}
	if err := ctx.ShouldBindJSON(purchaseRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	err = h.TicketService.PurchaseTicket(ticketID, purchaseRequest.Quantity)
	if err != nil {
		if restErr, ok := err.(customErrors.RestError); ok {
			ctx.JSON(restErr.Status, gin.H{"error": restErr.Message})
			return
		}

		log.Printf("Failed to purchase ticket: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to purchase ticket"})
		return
	}

	ctx.Status(http.StatusOK)
}
