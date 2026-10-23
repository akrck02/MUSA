// Package projects is the work the artist draws to order: who asked for it, how far it has got
// on a board of stages, when it is due and what it is paid. Delivering a project records its sale.
package projects

import (
	"fmt"
	"time"

	"github.com/Itros97/MUSA/calendar"
	"github.com/Itros97/MUSA/money"
)

// Stage is a column of the projects board.
type Stage string

// The stages, in the board's order. The words are the ones stored in the database.
const (
	ToStart   Stage = "TO_START"
	Drawing   Stage = "DRAWING"
	ToReview  Stage = "TO_REVIEW"
	Delivered Stage = "DELIVERED"
)

// Stages lists the stages in the board's order.
var Stages = []Stage{ToStart, Drawing, ToReview, Delivered}

// Project is one commission.
type Project struct {
	ID         string
	Title      string
	CustomerID string
	Stage      Stage
	Deadline   calendar.Date // zero when there is none
	Price      money.Money
	// VariantID is the digital product the project makes, when it is one of the catalogue's.
	VariantID string
	// SaleID is the sale its delivery recorded; empty until it is delivered.
	SaleID string
}

// DeadlineWords says when a project is due, counted from today, or "" when it has no deadline.
func DeadlineWords(deadline, today calendar.Date) string {
	if deadline.Zero() {
		return ""
	}
	days := daysBetween(today, deadline)
	switch {
	case days == 0:
		return "Vence hoy"
	case days == 1:
		return "Vence mañana"
	case days == -1:
		return "Venció ayer"
	case days < 0:
		return fmt.Sprintf("Venció hace %d días", -days)
	}
	return fmt.Sprintf("Vence en %d días", days)
}

// Late tells whether a project still to deliver is past its deadline.
func (project Project) Late(today calendar.Date) bool {
	return project.Stage != Delivered && !project.Deadline.Zero() && project.Deadline.Before(today)
}

// utc counts days between dates without a summer-time hour in the way.
var utc = time.UTC

// hoursInDay is how many hours a day has, to count days between two dates.
const hoursInDay = 24

func daysBetween(from, to calendar.Date) int {
	return int(to.At(0, utc).Sub(from.At(0, utc)).Hours() / hoursInDay)
}
