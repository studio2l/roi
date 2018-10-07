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
	// id는 어느 테이블에나 꼭 들어가야 하는 항목이다.
	"id UUID PRIMARY KEY DEFAULT gen_random_uuid()",
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
