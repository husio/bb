package main

import (
	"math"
	"net/url"
	"strconv"
)

type Paginator struct {
	page          int
	pageSize      int
	entitiesCount int
	pagesCount    int
}

func NewPaginator(q url.Values, pageSize int, entities int) *Paginator {
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	return &Paginator{
		page:          page,
		pageSize:      pageSize,
		entitiesCount: entities,
		pagesCount:    int(math.Ceil(float64(entities) / float64(pageSize))),
	}
}

func (p *Paginator) CurrentPage() int {
	return p.page
}

func (p *Paginator) PageCount() int {
	return p.pagesCount
}

func (p *Paginator) IsFirst() bool {
	return p.page == 1
}

func (p *Paginator) FirstPage() int {
	return 1
}

func (p *Paginator) IsLast() bool {
	return p.page == p.pagesCount
}

func (p *Paginator) LastPage() int {
	return p.pagesCount
}

func (p *Paginator) HasNext() bool {
	return p.entitiesCount > (p.page * p.pageSize)
}

func (p *Paginator) NextPage() int {
	return p.page + 1
}

func (p *Paginator) HasPrev() bool {
	return p.page > 1
}

func (p *Paginator) PrevPage() int {
	return p.page - 1
}

func (p *Paginator) Offset() uint {
	return uint((p.page - 1) * p.pageSize)
}

func (p *Paginator) Limit() uint {
	return uint(p.pageSize)
}
