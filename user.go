package roi

type User struct {
	ID             string
	HashedPassword string
	Name           string
	EngName        string
	Team           string
	Position       string
}

var UserTableFields = []string{
	// id는 자동 추가되는 필드이기 때문에 userid라는 이름을 사용
	"userid STRING UNIQUE NOT NULL CHECK (length(userid) > 0) CHECK (userid NOT LIKE '% %')",
	"hashed_password STRING NOT NULL",
	"name STRING",
	"eng_name STRING",
	"team STRING",
	"position STRING",
}

func (u User) dbKeyValues() []KV {
	kv := []KV{
		{"userid", q(u.ID)},
		{"hashed_password", q(u.HashedPassword)},
		{"name", q(u.Name)},
		{"eng_name", q(u.EngName)},
		{"team", q(u.Team)},
		{"position", q(u.Position)},
	}
	return kv
}
