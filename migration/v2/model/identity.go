package model

// Identity represents and identity with a set of credentials.
type Identity struct {
	Model
	Kind        string `gorm:"not null"`
	Name        string `gorm:"not null"`
	Description string
	User        string
	Password    string
	Key         string
	Settings    string
	Proxies     []Proxy `gorm:"constraint:OnDelete:SET NULL"`
}
