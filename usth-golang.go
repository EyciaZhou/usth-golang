package main

import (
	"gopkg.in/macaron.v1"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/redis"
	"github.com/EyciaZhou/usth-golang/C/CReply"
	"github.com/wendal/errors"
)

func main() {
	errors.AddStack = false
	m := macaron.Classic()
	m.Use(macaron.Renderer())
	m.Use(session.Sessioner(session.Options{
		Provider:"redis",
		ProviderConfig:"addr=127.0.0.1:6379",

			// 用于存放会话 ID 的 Cookie 名称，默认为 "MacaronSession"
			CookieName:     "MacaronSession",
			// Cookie 储存路径，默认为 "/"
			CookiePath:     "/",
			// GC 执行时间间隔，默认为 3600 秒
			Gclifetime:     3600,
			// 最大生存时间，默认和 GC 执行时间间隔相同
			Maxlifetime:    3600,
			// 仅限使用 HTTPS，默认为 false
			Secure:         false,
			// Cookie 生存时间，默认为 0 秒
			CookieLifeTime: 3600,
			// Cookie 储存域名，默认为空
			Domain:         "",
			// 会话 ID 长度，默认为 16 位
			IDLength:       16,
			// 配置分区名称，默认为 "session"
			Section:        "session",
	}))

	CReply.ApiRouterGroup(m)

	m.Run()
}
