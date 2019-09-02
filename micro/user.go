package main


const (
	missingSource = iota
	missingDest
	Payload
	Pic
	Cmd
	Arg
)




type Commands map[string]Decoder


type Carrier int

const (
	WEBSOCKET = iota + 1
	KAFKA
	NSQ
)




type Decoder func(env *Envelope) error
type Modules map[string]Module


type UserCommands int

type UserDb map[string]USER


type USER interface {
	Uname() string
	DisplayName() string
	Secret() string
	Uid() string
	Keys() string
	Type() string
	Flags() UserFlags
	SetFlags(f UserFlags)
	SetId(id string)
	InitKeys()
}



type User struct {
	Name    string `json:"username"`
	Displayname  string   `json:"displayname"`
	Password string `json:"password"`
	Key string `json:"keys"`
	Stamp int64 `json:"stamp"`
	Flag UserFlags `json:"flag"`
	Id string `json:"id"`
}


func (u *User) Uname() string {
	return u.Name
}

func (u *User) Keys() string {
	return u.Key
}

func (u *User) Flags() UserFlags {
	return u.Flag
}



func (u *User) DisplayName() string {
	return u.Displayname
}

func (u *User) Secret() string {
	return u.Password
}


func (u *User) SetFlags(f UserFlags) {
	  u.Flag = f
	  u.Stamp = stamp()
}

func (u *User) InitKeys() {
	u.Key = getNewKeysString()
}

func (u *User) Uid() string {
	return u.Id
}

func (u *User) SetId(id string)  {
	u.Id = id
	u.Stamp = stamp()
}



type UserFlags uint8

const (
	ONLINE UserFlags = 1 << iota
	DELETED
	ACTIVE
	HUMAN
	ROBOT
)

func UserSetFlag(u USER,flag UserFlags) {
	u.SetFlags(u.Flags() | flag)
}
func UserClearFlag(u USER,flag UserFlags) {
	u.SetFlags(u.Flags() &^ flag)
}

func UserToggleFlag(u USER,flag UserFlags){
	u.SetFlags(u.Flags() ^ flag)
}

func HasFlag(u USER,flag UserFlags) bool    {
	return u.Flags() & flag != 0
}


func UserActivate(u USER)  {
	if u == nil {
		panic("UserActivate: can not operate on nil pointer ")
	}
	UserSetFlag(u,ACTIVE)
	u.SetId(getId("USER"))
	u.InitKeys()
}










