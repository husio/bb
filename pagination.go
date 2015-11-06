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

func (p *Paginator) Pages() []int {
	pages := make([]int, p.pagesCount)
	for i := range pages {
		pages[i] = i + 1
	}
	return pages
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

func (p *Paginator) PageSize() uint {
	return PageSize
}

type SimplePaginator struct {
	Current int
	Next    int
	now     int
}

func NewSimplePaginator(now time.Time) *SimplePaginator {
	unix := int(now.Unix())
	return &SimplePaginator{
		now:     unix,
		Current: unix,
		Next:    0,
	}
}

func (p SimplePaginator) CurrentPage() int {
	return p.Current
}

func (p SimplePaginator) IsFirst() bool {
	return p.Current == p.now
}

func (p SimplePaginator) HasNext() bool {
	return p.Next != 0
}

func (p SimplePaginator) NextPage() int {
	return p.Next
}

func (p SimplePaginator) Limit() uint {
	return uint(PageSize)
}

func (p *SimplePaginator) PageSize() uint {
	return PageSize
}
