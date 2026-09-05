package store

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=255"`
	Until  string   `json:"until" validate:"max=255"`
	Since  string   `json:"since" validate:"max=255"`
}

func (p *PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	qs := r.URL.Query()

	// 1. Limit (avec valeur par défaut par exemple 10)
	limit := qs.Get("limit")
	log.Println("Limit from query string:", limit)
	if limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return PaginatedFeedQuery{}, err
		}
		p.Limit = l
	} else if p.Limit == 0 {
		p.Limit = 10 // Valeur par défaut si non spécifié
	}

	// 2. Offset
	offset := qs.Get("offset")
	if offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return PaginatedFeedQuery{}, err
		}
		p.Offset = o
	}

	// 3. Sort (avec valeur par défaut "desc")
	sort := qs.Get("sort")
	if sort != "" {
		p.Sort = sort
	} else if p.Sort == "" {
		p.Sort = "desc" // Valeur par défaut si non spécifié
	}

	// 4. Tags
	tags := qs.Get("tags")
	if tags != "" {
		p.Tags = strings.Split(tags, ",")
	}

	// 5. Search
	search := qs.Get("search")
	if search != "" {
		p.Search = search
	}

	// 6. since
	since := qs.Get("since")
	if since != "" {
		p.Since = parseTime(since)
	}

	return *p, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}
	return t.Format(time.DateTime)
}
