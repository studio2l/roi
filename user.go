package roi

// User는 사용자와 관련된 정보이다.
type User struct {
	ID          string
	KorName     string
	Name        string
	Team        string
	Position    string
	Email       string
	PhoneNumber string
	EntryDate   string
}

var UserTableFields = []string{
	"uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid()",
	"id STRING UNIQUE NOT NULL CHECK (length(id) > 0) CHECK (id NOT LIKE '% %')",
	"kor_name STRING NOT NULL",
	"name STRING NOT NULL",
	"team STRING NOT NULL",
	"position STRING NOT NULL",
	"email STRING NOT NULL",
	"phone_number STRING NOT NULL",
	"entry_date STRING NOT NULL",
	// hashed_password는 DB에서만 관리하고 User에는 들어가지는 않는다.
	"hashed_password STRING NOT NULL",
}

// NewUserMap은 새로운 사용자를 생성할 때 쓰는 맵이다.
func NewUserMap(id, hashed_password string) *ordMap {
	o := newOrdMap()
	o.Set("id", id)
	o.Set("kor_name", "")
	o.Set("name", "")
	o.Set("team", "")
	o.Set("position", "")
	o.Set("email", "")
	o.Set("phone_number", "")
	o.Set("entry_date", "")
	o.Set("hashed_password", hashed_password)
	return o
}

// ordMapFromUser는 유저 정보를 OrdMap에 담는다.
func ordMapFromUser(u *User) *ordMap {
	o := newOrdMap()
	o.Set("id", u.ID)
	o.Set("kor_name", u.KorName)
	o.Set("name", u.Name)
	o.Set("team", u.Team)
	o.Set("position", u.Position)
	o.Set("email", u.Email)
	o.Set("phone_number", u.PhoneNumber)
	o.Set("entry_date", u.EntryDate)
	return o
}
