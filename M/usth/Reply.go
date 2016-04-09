package usth
import (
	"github.com/Sirupsen/logrus"
	"time"
	"database/sql"
	"strconv"
)

type Reply struct {
	Time       int64  `json:"time"`
	AuthorName string `json:"authorName"` //name of post
	StuId      string `json:"stuId"`
	Content    string `json:"content"` //length of 150
	ClassName  string `json:"className"`
	Digg       int    `json:"digg"`

	Id string `json:"id"`

	RefId       string `json:"refId"`
	RefedAuthorId string `json:"refAuthorId`
	RefedAuthor string `json:"refedAuthor"` //cache
	RefedContent string `json:"refedContent"` //cache

	Digged bool `json:"digged"`
}

type Dbreply struct{}
var DBReply = &Dbreply{}

var (
	TIME_DIGG = "点赞时出错"
	TIME_GET_DIGG = "获取点赞数量时出错"
	TIME_GET_REPLY = "获取评论时出错"
	TIME_REPLY = "回复时出错"
)

type scanner interface {
	Scan(dest ...interface{}) error
}

func (p *Dbreply) scanfOne(row scanner) (*Reply, error) {
	var (
		NullRefId sql.NullString
		NullRefAuthor sql.NullString
		NullRefAuthorId sql.NullString
		NullRefContent sql.NullString
	)

	r := &Reply{}
	err := row.Scan(
		&r.Id, &r.Time, &r.AuthorName, &r.StuId, &r.Content, &r.ClassName,
		&r.Digg, &NullRefId, &NullRefAuthorId, &NullRefAuthor, &NullRefContent,
	)

	if err != nil {
		return nil, err
	}

	r.RefId = NullRefId.String
	r.RefedAuthor = NullRefAuthor.String
	r.RefedAuthorId = NullRefAuthorId.String
	r.RefedContent = NullRefContent.String

	return r, nil
}

func (p *Dbreply) scanfOneWithDigg(row scanner) (*Reply, error) {
	var (
		NullRefId sql.NullString
		NullRefAuthor sql.NullString
		NullRefAuthorId sql.NullString
		NullRefContent sql.NullString

		NullDigged sql.NullString
	)

	r := &Reply{}
	err := row.Scan(
		&r.Id, &r.Time, &r.AuthorName, &r.StuId, &r.Content, &r.ClassName,
		&r.Digg, &NullRefId, &NullRefAuthorId, &NullRefAuthor, &NullRefContent, &NullDigged,
	)

	if err != nil {
		return nil, err
	}

	r.RefId = NullRefId.String
	r.RefedAuthor = NullRefAuthor.String
	r.RefedAuthorId = NullRefAuthorId.String
	r.RefedContent = NullRefContent.String
	r.Digged = NullDigged.Valid

	return r, nil
}

func (p *Dbreply) GetReplyFirstPage(FromId string, ClassName string, Limit int) ([]*Reply, error) {
	if (Limit > 50) {
		Limit = 50
	}

	var (
		rows *sql.Rows
		err error
		scan func (row scanner) (*Reply, error)
	)

	if FromId != "" {
		rows, err = db.Query(`
			SELECT
					id, _time, author_name, _reply.stu_id, content, class_name, digg, refid, ref_author_id, ref_author, ref_content, b.stu_id
				FROM _reply
				LEFT JOIN
						(SELECT reply_id, stu_id FROM diggs WHERE stu_id=?) AS b
					ON _reply.id=b.reply_id
				WHERE class_name=?
				ORDER BY _time DESC
				LIMIT ?
			`, FromId, ClassName, Limit)

		scan = p.scanfOneWithDigg

	} else {
		rows, err = db.Query(`
			SELECT
					id, _time, author_name, stu_id, content, class_name, digg, refid, ref_author_id, ref_author, ref_content
				FROM _reply
				WHERE class_name=?
				ORDER BY _time DESC
				LIMIT ?
			`, ClassName, Limit)

		scan = p.scanfOne
	}
	if err != nil {
		return nil, newError(TIME_GET_REPLY, "获取评论失败"+err.Error())
	}
	defer rows.Close()

	result := make([]*Reply, Limit)

	var i int
	for i = 0; rows.Next(); i++ {
		one_row, err := scan(rows)
		if err != nil {
			return nil, newError(TIME_GET_REPLY, "获取评论失败 : " + err.Error())
		}
		result[i] = one_row
	}
	if err = rows.Err(); err != nil {
		return nil, newError(TIME_GET_REPLY, "获取评论失败 : " + err.Error())
	}

	return result[:i], nil
}

func (p *Dbreply) GetReplyPageFlip(FromId string, ClassName string, Limit int, lstti int64, lstid string) ([]*Reply, error) {
	if (Limit > 50) {
		Limit = 50
	}

	var (
		rows *sql.Rows
		err error
		scan func (row scanner) (*Reply, error)
	)

	if FromId != "" {

		rows, err = db.Query(`
			SELECT
					id, _time, author_name, _reply.stu_id, content, class_name, digg, refid, ref_author_id, ref_author, ref_content, b.stu_id
				FROM _reply
				LEFT JOIN
						(SELECT reply_id, stu_id FROM diggs WHERE stu_id=?) AS b
					ON _reply.id=b.reply_id
				WHERE class_name=? AND ? >= _time
				ORDER BY _time DESC
				LIMIT ?
			`, FromId, ClassName, lstti, Limit + 1)

		scan = p.scanfOneWithDigg

	} else {

		rows, err = db.Query(`
			SELECT
					id, _time, author_name, stu_id, content, class_name, digg, refid, ref_author_id, ref_author, ref_content
				FROM _reply
				WHERE class_name=? AND ? >= _time
				ORDER BY _time DESC
				LIMIT ?
			`, ClassName, lstti, Limit + 1)

		scan = p.scanfOne


	}

	if err != nil {
		return nil, newError(TIME_GET_REPLY, "获取评论失败" + err.Error())
	}

	defer rows.Close()

	result := make([]*Reply, Limit+1)

	var i int
	for i = 0; rows.Next(); i++ {
		one_row, err := scan(rows)
		if err != nil {
			return nil, newError(TIME_GET_REPLY, "获取评论失败 : " + err.Error())
		}

		if one_row.Id == lstid {
			i--
			continue
		}

		result[i] = one_row
	}
	if err = rows.Err(); err != nil {
		return nil, newError(TIME_GET_REPLY, "获取评论失败 : " + err.Error())
	}

	return result[:i], nil
}


func (p *Dbreply) GetReplyFrom(FromId string, Id string) (*Reply, error) {
	row := db.QueryRow(`
		SELECT
				id, _time, author_name, _reply.stu_id, content, class_name, digg, refid, ref_author_id, ref_author, ref_content, b.stu_id
			FROM _reply
			LEFT JOIN
						(SELECT reply_id, stu_id FROM diggs WHERE stu_id=?) AS b
					ON _reply.id=b.reply_id
			WHERE id=?
			LIMIT 1
	`, Id)

	r, err := p.scanfOneWithDigg(row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, newError(TIME_GET_REPLY, err.Error())
		}
		return nil, newError(TIME_GET_REPLY, "没有该评论")
	}

	return r, nil
}

func (p *Dbreply) GetReply(Id string) (*Reply, error) {
	row := db.QueryRow(`
		SELECT
				id, _time, author_name, stu_id, content, class_name, digg, refid, ref_author_id, ref_author, ref_content
			FROM _reply
			WHERE id=?
			LIMIT 1
	`, Id)

	r, err := p.scanfOne(row)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, newError(TIME_GET_REPLY, err.Error())
		}
		return nil, newError(TIME_GET_REPLY, "没有该评论")
	}

	return r, nil
}

func (p *Dbreply) RemoveReply(Id string, IdOfBy string) {
	//TODO
}

func (p *Dbreply) Reply(AuthorName string, StuId string, Content string, ClassName string) (_Id string, _err error) {
	return p.ReplyWithRef(AuthorName, StuId, Content, ClassName, "")
}

func (p *Dbreply) ReplyWithRef(AuthorName string, StuId string, Content string, ClassName string, RefId string) (_Id string, _err error) {
	ti := time.Now().Unix();
	if (len(Content) > 200) {
		return
	}

	var (
		result sql.Result
		err error
	)

	if RefId == "" {
		result, err = db.Exec(`
			INSERT INTO
					_reply(_time, author_name, stu_id, content, class_name, digg)
				VALUES
					(?,?,?,?,?,0)
		`, ti, AuthorName, StuId, Content, ClassName)
	} else {
		ref_reply, e := p.GetReply(RefId)
		if e != nil {
			return "", e
		}
		result, err = db.Exec(`
		INSERT INTO
					_reply(_time, author_name, stu_id, content, class_name, digg, refid, ref_author_id, ref_author, ref_content)
				VALUES
					(?,?,?,?,?,0,?,?,?,?)
		`, ti, AuthorName, StuId, Content, ClassName, RefId, ref_reply.StuId, ref_reply.AuthorName, ref_reply.Content)
	}
	if err != nil {
		return "", newErrorByError(TIME_REPLY, err)
	}
	id, _ := result.LastInsertId()
	return strconv.FormatInt(id, 10), nil
}

func (p *Dbreply) GetDigg(Id string) (int, error) {
	row := db.QueryRow(`
		SELECT
				digg
			FROM _reply
			WHERE (id = ?)
			LIMIT 1
	`, Id)

	var digg_cnt int
	if err := row.Scan(&digg_cnt); err != nil {
		logrus.Error(TIME_GET_DIGG, err.Error())
		return 0, newErrorByError(TIME_GET_DIGG, err)
	}
	return digg_cnt, nil
}

func (p *Dbreply) Digg(Id string, FromId string) error {
	result, _ := db.Exec(`
		INSERT INTO
				diggs(reply_id, stu_id)
			VALUES(?, ?)
	`, Id, FromId)

	diged, _ := result.RowsAffected()
	if diged == 0 {
		return newError(TIME_DIGG, "重复点赞")
	}

	result, err := db.Exec(`
		UPDATE
			_reply
		SET
			digg = digg + 1
		WHERE (id = ?)
		LIMIT 1
	`, Id)
	if err != nil {
		logrus.Error(TIME_DIGG, err.Error())
		return newErrorByError(TIME_DIGG, err)
	}

	row_cnt, _ := result.RowsAffected()
	if row_cnt != 1 {
		return newError(TIME_DIGG, "点赞失败")
	}

	return nil
}