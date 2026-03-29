package repository

type Repository interface {
	UserRepository
	OrderRepository
	WithdrawalRepository
}
