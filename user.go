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

func (u User) dbKeyValues() []KV {
	kv := []KV{
		{"userid", q(u.ID)},
		{"hashed_password", q(u.HashedPassword)},
		{"kor_name", q(u.KorName)},
		{"name", q(u.Name)},
		{"team", q(u.Team)},
		{"position", q(u.Position)},
		{"email", q(u.Email)},
		{"phone_number", q(u.PhoneNumber)},
		{"entry_date", q(u.EntryDate)},
	}
	return kv
}
