package config

type Password struct {
	BCryptCost int `env:"PASSWORD_BCRYPT_COST" envDefault:"12"`
}
