package setup

import (
	"github.com/TitechMeister/Neon/port"
	"github.com/labstack/echo"
)

type Sencor interface {
	// センサーの名前を取得する
	GetSencorName() string
	// Echoサーバ経由のリクエストでデータを取得する
	GetData(c echo.Context) error
	// Echoサーバ経由のリクエストでデータをログに記録する
	PostData(c echo.Context) error
	// serialサーバを叩いてログ記録
	LogData() error
	// serialサーバを叩く頻度を返す(周波数)
	GetLogFrequency() int
	// Echoサーバ経由のリクエストでデータの履歴を取得する
	GetHistory(c echo.Context) error
}

type Neon struct {
	Sencors []*Sencor
	Port    *port.Port
}
