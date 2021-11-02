package pairing

type Config struct {
	PairingUrl  string `split_words:"true"`
	PairingTTL  int    `split_words:"true"`
	CompanyName string `split_words:"true"`
	CompanyLogo string `split_words:"true"`
}
