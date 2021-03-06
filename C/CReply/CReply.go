package CReply

import (
	"gopkg.in/macaron.v1"
	"github.com/go-macaron/session"
	"github.com/EyciaZhou/usth-golang/M/usth"
	"github.com/EyciaZhou/msghub-http/C"
	"github.com/wendal/errors"
	"net/http"
)

func ApiRouterGroup(m *macaron.Macaron) {
	m.Get("/Logout", Logout)

	m.Post("/Score", GetScore)

	m.Group("/Reply", func() {
		m.Get("/course/:classname/:limit/:lstid/:lstti", getReplys)
		m.Get("/:id", getReply)

		m.Post("/course/:classname", addReply)
		m.Post("/course/:classname/:id/reply", addReply)

		m.Get("/:id/digg", getDiggCount)
		m.Post("/:id/digg/add", diggAdd)
	})

	m.Get("/head/getToken", getHeadToken)
	m.Post("/head/callback", usth.HeadStore.Callback)
}

func getHeadToken(ctx *macaron.Context, f session.Store) {
	username := f.Get("api_usernamels");
	if username == nil || username == "" {
		ctx.JSON(http.StatusOK, C.PackError(nil, errors.New("没有登陆或者登录过期")))
		return
	}
	token := usth.HeadStore.MakeupUploadToken(username.(string))
	ctx.JSON(http.StatusOK, C.Pack(token))
}

func getDiggCount(ctx *macaron.Context, f session.Store) {
	reply_id := ctx.Params(":id")
	ctx.JSON(200, C.PackError(usth.DBReply.GetDigg(reply_id)))
}

func diggAdd(ctx *macaron.Context, f session.Store) {
	reply_id := ctx.Params(":id")
	if f.Get("api_username") == nil || f.Get("api_username") == "" {
		ctx.JSON(200, C.PackError(nil, errors.New("没有登陆或者登录过期")))
		return
	}
	err := usth.DBReply.Digg(reply_id, f.Get("api_username").(string))
	if err == nil {
		getDiggCount(ctx, f)
	} else {
		ctx.JSON(200, C.PackError(nil, err))
	}
}

func addReply(ctx *macaron.Context, f session.Store) {
	classname := ctx.Params(":classname")
	content := ctx.Query("content")
	if f.Get("api_username") == nil || f.Get("api_username") == "" {
		ctx.JSON(200, C.PackError(nil, errors.New("没有登陆或者登录过期")))
		return
	}
	username := f.Get("api_username").(string)
	author_name, err := usth.DBInfo.GetName(username)
	if err != nil {
		ctx.JSON(200, C.PackError(nil, err))
		return
	}

	id := ctx.Params(":id")
	if id == "" {
		ctx.JSON(200, C.PackError(usth.DBReply.Reply(author_name, username, content, classname)))
	} else {
		ctx.JSON(200, C.PackError(usth.DBReply.ReplyWithRef(author_name, username, content, classname, id)))
	}
}

func getReply(ctx *macaron.Context, f session.Store) {
	id := ctx.Params(":id")
	username := f.Get("api_username")
	if (username == nil) {
		username = ""
	}

	if (username == "") {
		ctx.JSON(200, C.PackError(usth.DBReply.GetReply(id)))
	} else {
		ctx.JSON(200, C.PackError(usth.DBReply.GetReplyFrom(username.(string), id)))
	}
}

func getReplys(ctx *macaron.Context, f session.Store) {
	classname, limit, lstti, lstid := ctx.Params(":classname"), ctx.ParamsInt(":limit"), ctx.ParamsInt64(":lstti"), ctx.Params(":lstid")
	username := f.Get("api_username")
	if (username == nil) {
		username = ""
	}

	if limit > 20 || limit <= 0 {
		limit = 20
	}
	if lstti < 0 {
		ctx.JSON(200, C.PackError(usth.DBReply.GetReplyFirstPage(username.(string), classname, limit)))
		return
	}
	ctx.JSON(200, C.PackError(usth.DBReply.GetReplyPageFlip(username.(string), classname, limit, lstti, lstid)))
}

func Logout(ctx *macaron.Context, f session.Store) {
	f.Delete("api_username")
}

func GetScore(ctx *macaron.Context, f session.Store) {
	username, password, _type := ctx.Query("username"), ctx.Query("password"), ctx.Query("type")

	if f.Get("api_username") != username {
		f.Delete("api_username")
	}

	resp, err := usth.DBScore.Get(username, password, _type)
	if err == nil {
		//logined
		f.Set("api_username", username)
	}

	ctx.JSON(200, C.PackError(resp, err))
}