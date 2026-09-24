import { request } from './api.svelte'

export type DayViews = { day: string; views: number }
export type PageViews = { slug: string; title: string; views: number }
export type SearchStat = { query: string; count: number; results: number }
export type Insights = {
  days: number
  views_total: number
  views_by_day: DayViews[]
  top_pages: PageViews[]
  top_searches: SearchStat[]
  zero_result_searches: SearchStat[]
  feedback: { helpful: number; not_helpful: number }
}

export const insightsApi = {
  insights: (project: string, days = 30) =>
    request<Insights>('GET', `/api/admin/projects/${encodeURIComponent(project)}/insights?days=${days}`),
}
