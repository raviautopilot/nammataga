package model

type User struct {
	ID         string `json:"id" bson:"_id,omitempty"`
	Email      string `json:"email" bson:"email"`
	Password   string `json:"-" bson:"password"`
	FirstLogin bool   `json:"firstLogin" bson:"first_login"`
}
