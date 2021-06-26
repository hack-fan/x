package xecho

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Paginator can be embedded to request struct
// You can check the input by mod validator
// If the value is wrong, It would be change to default
type Paginator struct {
	Page     int `query:"page" validate:"omitempty,min=1"`
	PageSize int `query:"page_size" validate:"omitempty,min=5,max=50"`
}

func (p *Paginator) check() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 5 || p.PageSize > 50 {
		p.PageSize = 10
	}
}

// Offset will used by db query
func (p Paginator) Offset() int {
	p.check()
	return p.PageSize * (p.Page - 1)
}

// Apply paginator to gorm query
func (p Paginator) Apply(tx *gorm.DB) *gorm.DB {
	offset := p.Offset()
	return tx.Offset(offset).Limit(p.PageSize)
}

// AddHeader to echo resp
func (p Paginator) AddHeader(c echo.Context, total int) {
	p.check()
	c.Response().Header().Set("X-Total-Count", strconv.Itoa(total))
	c.Response().Header().Set("X-Page-Current", strconv.Itoa(p.Page))
	c.Response().Header().Set("X-Page-Size", strconv.Itoa(p.PageSize))
}
