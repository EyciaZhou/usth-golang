package usth
import (
	"database/sql"
	"github.com/EyciaZhou/msghub-http/M/MUser"
)

type DbInfo struct{}
var DBInfo = &DbInfo{}

func (p *DbInfo) Update(username string, password string, name string, class string) { //silence //TODO: class
	db.Exec(`
	INSERT INTO
			info (stu_id, pwd, author_name, class)
		VALUES
			(?,?,?,?)
		ON DUPLICATE KEY UPDATE
			stu_id = VALUES(stu_id),
			pwd = VALUES(pwd),
			author_name = VALUES(author_name),
			class = VALUES(class)
	`, username, password, name)
}

func (p *DbInfo) GetUserInfo(stuid string) (*SchoolRollInfo, error) {
	var school_roll_info SchoolRollInfo

	row := db.QueryRow(`
	SELECT stu_id, author_name, class, head
		FROM info
		WHERE stu_id=?
		LIMIT 1`, stuid)

	var head sql.NullString

	err := row.Scan(&school_roll_info.Stu_id, &school_roll_info.Name, &school_roll_info.Class, &head)
	if err != nil {
		return nil, err
	}

	if head.Valid {
		school_roll_info.Head = MUser.HeadStore.GetHead(school_roll_info.Stu_id, head.String)
	}

	return &school_roll_info, nil
}

func (p *DbInfo) GetName(username string) (string, error) {
	row := db.QueryRow(`
	SELECT author_name
		FROM info
		WHERE stu_id=?
		LIMIT 1
	`, username)

	var author_name string

	err := row.Scan(&author_name)
	if err != nil {
		return "", err
	}
	return author_name, nil
}

func (p *DbInfo) MarkHead(username string, value string) error {
	_, err := db.Exec(`UPDATE info
				SET head=?
				WHERE stu_id=?`, value, username)
	return err
}