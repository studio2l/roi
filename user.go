package roi

type User struct {
	ID             string
	HashedPassword string
	KorName        string
	Name           string
	Team           string
	Position       string
	Email          string
	PhoneNumber    string
	EntryDate      string
}

var UserTableFields = []string{
	// id는 자동 추가되는 필드이기 때문에 userid라는 이름을 사용
	"userid STRING UNIQUE NOT NULL CHECK (length(userid) > 0) CHECK (userid NOT LIKE '% %')",
	"hashed_password STRING NOT NULL",
	"kor_name STRING",
	"name STRING",
	"team STRING",
	"position STRING",
	"email STRING",
	"phone_number STRING",
	"entry_date STRING",
}

func (u User) toOrdMap() *ordMap {
	o := newOrdMap()
	o.Set("userid", u.ID)
	o.Set("hashed_password", u.HashedPassword)
	o.Set("kor_name", u.KorName)
	o.Set("name", u.Name)
	o.Set("team", u.Team)
	o.Set("position", u.Position)
	o.Set("email", u.Email)
	o.Set("phone_number", u.PhoneNumber)
	o.Set("entry_date", u.EntryDate)
	return o
}
