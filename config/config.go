package config

import (
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	HttpPort                    string
	BaseUrl                     string
	JwtAccessTokenSecretKey     string
	AuthCookieNameCertAlert     string
	Redis                       string
	LocationInfoKey             string
	TelegramApiToken            string
	TelegramBotUsername         string
	SignUPLinkTokenTime         time.Duration
	ForgotPasswordLinkTokenTime time.Duration
	UpdateEmailLinkTokenTime    time.Duration
	PullUpdateDomainInterval    time.Duration
	Postgres                    Postgres
	Google                      Google
	Smtp                        Smtp
}

type Smtp struct {
	Sender   string
	Password string
}

type Google struct {
	Conf *oauth2.Config
}

type Postgres struct {
	Database string
	User     string
	Password string
	Port     string
	Host     string
}

func Load() Config {
	godotenv.Load()

	conf := viper.New()
	conf.AutomaticEnv()

	return Config{
		HttpPort:                    conf.GetString("HTTP_PORT"),
		BaseUrl:                     conf.GetString("BASE_URL"),
		JwtAccessTokenSecretKey:     conf.GetString("JWT_ACCESS_TOKEN_SECRET_KEY"),
		AuthCookieNameCertAlert:     conf.GetString("AUTH_COOKIE_NAME_CERTALERT"),
		SignUPLinkTokenTime:         conf.GetDuration("SIGNUP_TOKEN_LINK_DURATION"),
		ForgotPasswordLinkTokenTime: conf.GetDuration("FORGOT_PASSWORD_TOKEN_LINK_DURATION"),
		UpdateEmailLinkTokenTime:    conf.GetDuration("UPDATE_EMAIL_TOKEN_LINK_DURATION"),
		Redis:                       conf.GetString("REDIS_ADDR"),
		LocationInfoKey:             conf.GetString("LOCATION_INFO_KEY"),
		Postgres: Postgres{
			Host:     conf.GetString("POSTGRES_HOST"),
			Port:     conf.GetString("POSTGRES_PORT"),
			User:     conf.GetString("POSTGRES_USER"),
			Password: conf.GetString("POSTGRES_PASSWORD"),
			Database: conf.GetString("POSTGRES_DATABASE"),
		},
		Google: Google{
			Conf: &oauth2.Config{
				ClientID:     conf.GetString("GOOGLE_CLIENT_ID"),
				ClientSecret: conf.GetString("GOOGLE_SECRET_KEY"),
				RedirectURL:  conf.GetString("GOOGLE_REDIRECT_URI"),
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
					"https://www.googleapis.com/auth/userinfo.profile",
				},
				Endpoint: google.Endpoint,
			},
		},
		Smtp: Smtp{
			Sender:   conf.GetString("SMTP_SENDER"),
			Password: conf.GetString("SMTP_PASSWORD"),
		},
		PullUpdateDomainInterval: conf.GetDuration("PULL_UPDATE_DOMAIN_INTERVAL"),
		TelegramApiToken:         conf.GetString("TELEGRAM_APITOKEN"),
		TelegramBotUsername:      conf.GetString("TELEGRAM_BOT_USERNAME"),
	}
}
