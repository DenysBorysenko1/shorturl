package repository

type StatsProvider interface {
	CountURLs() (int, error)
	CountUsers() (int, error)
}
