package migrations

import (
	"dev.jevido/jevidocs/services/api/app/facades"
)

// M20260924000020CreateAnalyticsTables adds daily page view and search
// counters for the admin's Insights tab.
type M20260924000020CreateAnalyticsTables struct{}

func (r *M20260924000020CreateAnalyticsTables) Signature() string {
	return "20260924000020_create_analytics_tables"
}

func (r *M20260924000020CreateAnalyticsTables) Up() error {
	s := facades.Schema()
	if !s.HasTable("page_views") {
		if err := s.Sql(`CREATE TABLE page_views (
			id bigserial PRIMARY KEY,
			project_id bigint NOT NULL,
			slug varchar(255) NOT NULL,
			day date NOT NULL,
			views integer NOT NULL DEFAULT 0,
			UNIQUE (project_id, slug, day))`); err != nil {
			return err
		}
	}
	if !s.HasTable("search_queries") {
		if err := s.Sql(`CREATE TABLE search_queries (
			id bigserial PRIMARY KEY,
			project_id bigint NOT NULL,
			query varchar(100) NOT NULL,
			day date NOT NULL,
			count integer NOT NULL DEFAULT 0,
			results integer NOT NULL DEFAULT 0,
			UNIQUE (project_id, query, day))`); err != nil {
			return err
		}
	}
	return nil
}

func (r *M20260924000020CreateAnalyticsTables) Down() error {
	if err := facades.Schema().DropIfExists("page_views"); err != nil {
		return err
	}
	return facades.Schema().DropIfExists("search_queries")
}
