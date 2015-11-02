package main

import (
	"math"
	"net/url"
	"strconv"
	"time"
)

const PageSize = 25

type Paginator struct {
	page          int
	entitiesCount int
	pagesCount    int
}

func NewPaginator(q url.Values, entities int) *Paginator {
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	return &Paginator{
		page:          page,
		entitiesCount: entities,
		pagesCount:    int(math.Ceil(float64(entities) / float64(PageSize))),
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
	return p.entitiesCount > (p.page * PageSize)
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
	return uint((p.page - 1) * PageSize)
}

func (p *Paginator) Limit() uint {
	return uint(PageSize)
}

type SimplePagination struct {
	Current int
	Next    int
	now     int
}

func NewSimplePagination(now time.Time) *SimplePagination {
	unix := int(now.Unix())
	return &SimplePagination{
		now:     unix,
		Current: unix,
		Next:    0,
	}
}

func (p SimplePagination) CurrentPage() int {
	return p.Current
}

func (p SimplePagination) IsFirst() bool {
	return p.Current == p.now
}

func (p SimplePagination) HasNext() bool {
	return p.Next != 0
}

func (p SimplePagination) NextPage() int {
	return p.Next
}

func (p SimplePagination) Limit() uint {
	return uint(PageSize)
}
