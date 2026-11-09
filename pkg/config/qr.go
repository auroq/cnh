package config

import "github.com/spf13/viper"

const KeyQRBaseURL = "qr.base_url"

var qrDefaults = map[string]any{
	KeyQRBaseURL: "https://widedoor.cnh.auroq.com",
}

// QR is what the code renderer needs: the public address the codes open.
type QR interface {
	BaseURL() string
}

func (c *Configuration) QR() QR { return qr{c.v} }

type qr struct{ v *viper.Viper }

func (q qr) BaseURL() string { return q.v.GetString(KeyQRBaseURL) }
