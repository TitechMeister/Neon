package setup

import "github.com/labstack/echo"

type Sencor interface {
	// センサーの名前を取得する
	GetSencorName() string
	// Echoサーバ経由のリクエストでデータを取得する
	GetData(c echo.Context) error
	// Echoサーバ経由のリクエストでデータをログに記録する
	PostData(c echo.Context) error
	// serialサーバを叩いてログ
	// Echoサーバ経由のリクエストでデータの履歴を取得する
	GetHistory(c echo.Context) error
}

type Neon struct {
	Sencors []*Sencor
}
