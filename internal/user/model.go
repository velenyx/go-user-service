package user

type User struct {
	ID           string `json:"id" bson:"_id,omitempty"`
	Email        string `json:"email" bson:"email,omitempty"`
	Username     string `json:"username" bson:"username,omitempty"`
	PasswordHash string `json:"-" bson:"password,omitempty"`
}

type CreateUserDTO struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}
