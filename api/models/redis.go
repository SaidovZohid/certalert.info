package models

type UserRedis struct {
	FirstName        string
	LastName         string
	Email            string
	Password         string
	IsUserAgreeTerms bool
}
