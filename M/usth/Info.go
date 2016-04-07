package usth

type DbInfo struct{}
var DBInfo = &DbInfo {}

func (p *DbInfo) Update(username string, password string, name string) { //silence
	db.Exec(`
	INSERT INTO
			info (stu_id, pwd, author_name)
		VALUES
			(?,?,?)
		ON DUPLICATE KEY UPDATE
			stu_id = VALUES(stu_id),
			pwd = VALUES(pwd),
			author_name = VALUES(author_name)
	`, username, password, name)
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