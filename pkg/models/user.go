package models

// Address Model
type Address struct {
	Model
	Street     string `json:"street"`
	PostalCode string `gorm:"column:postal_code" json:"postalCode"`
	City       string `json:"city"`
	State      string `json:"state"`
	Country    string `json:"country"`
	Landline   string `json:"landline"`
	Mobile     string `json:"mobile"`
	UserID     uint32 `json:"userId"`
}

type Role struct {
	Model
	Name string `json:"name"`
}

// User Model
type User struct {
	Model
	UserName   string  `gorm:"column:username" json:"username"`
	BusinessID uint32  `json:"businessId"`
	Address    Address `json:"address"`
	Account    Account `json:"account"`
	Roles      []Role  `gorm:"many2many:user_roles;" json:"roles"`
}

// Auth Model
type Auth struct {
	Token        string
	RefreshToken string
}
