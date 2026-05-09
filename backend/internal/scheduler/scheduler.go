package scheduler

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
	"meddoc/internal/repository"
)

type Scheduler struct {
	c       *cron.Cron
	tickets *repository.TicketRepo
}

func New(db *pgxpool.Pool) *Scheduler {
	return &Scheduler{
		c:       cron.New(),
		tickets: repository.NewTicketRepo(db),
	}
}

func (s *Scheduler) Start() {
	// Run every day at 08:00
	s.c.AddFunc("0 8 * * *", s.escalateOldTickets)
	s.c.Start()
	log.Println("Scheduler started: auto-escalation at 08:00 daily")
}

func (s *Scheduler) Stop() {
	s.c.Stop()
}

func (s *Scheduler) escalateOldTickets() {
	n, err := s.tickets.EscalateOldTickets(context.Background())
	if err != nil {
		log.Printf("Escalation error: %v", err)
		return
	}
	if n > 0 {
		log.Printf("Auto-escalated %d tickets to HIGH priority", n)
	}
}
