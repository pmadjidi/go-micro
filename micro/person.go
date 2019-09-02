package main


type PERSON interface {
	USER
	GivenName() string
	SirName() string
	Email() string
	Location() string
	Avatar() string

}

type Person struct {
	*User
	Givenname    string   `json:"givenname"`
	Sirname   string    `json:"sirname"`
	Email string   `json:"email"`
	Location string   `json:"location"`
	Avatar	string   `json:"avatar"`
}

func (p *Person) GivenUname() string {
	return p.Givenname
}

func (p *Person) SirName() string {
	return p.Sirname
}

func (p *Person) Mail() string {
	return p.Email
}

func (p *Person) Place() string {
	return p.Location
}


func (p *Person) Type() string {
	return "PERSON"
}







