package controllers

import (
	"errors"
	"time"

	"aahframework.org/aah.v0"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/keiwi/api/app/models"
	"github.com/keiwi/utils"
	storageModel "github.com/keiwi/utils/models"
	utilNats "github.com/keiwi/utils/nats"
	"github.com/nats-io/go-nats"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

type UsersController struct {
	*aah.Context
}

func (a *UsersController) UserSignup(signup models.User) {
	if signup.Username == "" || signup.Password == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Missing username or password"})
		return
	}

	inter := a.Get("nats")
	if inter == nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	n, ok := inter.(*nats.Conn)
	if !ok {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	hasUsername := utils.HasOptions{
		Filter: utils.Filter{"username": signup.Username},
	}

	data, err := bson.MarshalJSON(hasUsername)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	has, err := utilNats.HasUser(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}
	if has {
		a.Reply().BadRequest().JSON(models.Response{Message: "Username already exists"})
		return
	}

	hasEmail := utils.HasOptions{
		Filter: utils.Filter{"email": signup.Email},
	}

	data, err = bson.MarshalJSON(hasEmail)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	has, err = utilNats.HasUser(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}
	if has {
		a.Reply().BadRequest().JSON(models.Response{Message: "A user with this email already exists"})
		return
	}

	passwordHash := HashPassword(signup.Username, signup.Password)

	user := storageModel.User{
		Username: signup.Username,
		Email:    signup.Email,
		Password: passwordHash,
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	data, err = bson.MarshalJSON(user)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	err = utilNats.CreateUser(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	jsontoken := GetJSONToken(&user)
	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully signed up", Data: jsontoken})
}

func (a *UsersController) UserLogin(login models.User) {
	if login.Username == "" || login.Password == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Missing username or password"})
		return
	}

	inter := a.Get("nats")
	if inter == nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	n, ok := inter.(*nats.Conn)
	if !ok {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	findUser := utils.FindOptions{
		Filter: utils.Filter{"username": login.Username},
		Sort:   utils.Sort{"-created_at"},
		Limit:  1,
	}

	data, err := bson.MarshalJSON(findUser)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	users, err := utilNats.FindUser(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	var user storageModel.User
	if len(users) <= 0 {
		findUser := utils.FindOptions{
			Filter: utils.Filter{"email": login.Email},
			Sort:   utils.Sort{"-created_at"},
			Limit:  1,
		}

		data, err := bson.MarshalJSON(findUser)
		if err != nil {
			a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
			return
		}

		users, err = utilNats.FindUser(n, data)
		if err != nil {
			a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
			return
		}

		if len(users) <= 0 {
			a.Reply().BadRequest().JSON(models.Response{Message: "Username or email not found"})
			return
		} else {
			user = users[0]
		}
	} else {
		user = users[0]
	}

	if !CheckPassword(user.Password, login.Password) {
		a.Reply().BadRequest().JSON(models.Response{Message: "Bad password"})
		return
	}

	jsontoken := GetJSONToken(&user)
	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully logged in", Data: jsontoken})
}

// UserInfo - example to get
func (a *UsersController) UserInfo() {
	inter := a.Get("nats")
	if inter == nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	n, ok := inter.(*nats.Conn)
	if !ok {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	user, err := GetUserFromContext(n, a.Context)
	if err != nil {
		a.Reply().BadRequest().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Data: user})
}

// signinKey set up a global string for our secret
var signinKey = []byte("kdsadsndadsafs")

// JwtMiddleware handler for jwt tokens
var JwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return signinKey, nil
	},
	UserProperty:  "user",
	Debug:         false,
	SigningMethod: jwt.SigningMethodHS256,
})

// GetToken create a jwt token with user claims
func GetToken(user *storageModel.User) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["uuid"] = user.ID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	signedToken, _ := token.SignedString(signinKey)
	return signedToken
}

// GetJSONToken create a JSON token string
func GetJSONToken(user *storageModel.User) string {
	token := GetToken(user)
	jsontoken := "{\"id_token\": \"" + token + "\"}"
	return jsontoken
}

// GetUserClaimsFromContext return "user" claims as a map from request
func GetUserClaimsFromContext(context *aah.Context) map[string]interface{} {
	token := context.Get("user").(*jwt.Token)
	claims := token.Claims.(jwt.MapClaims)
	return claims
}

// HashPassword - Hash the password (takes a username as well, it can be used for salting).
func HashPassword(username, password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("Permissions: bcrypt password hashing unsuccessful")
	}
	return string(hash)
}

// CheckPassword - compare a hashed password with a possible plaintext equivalent
func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}

// GetUserFromContext - return User reference from header token
func GetUserFromContext(n *nats.Conn, context *aah.Context) (*storageModel.User, error) {
	userclaims := GetUserClaimsFromContext(context)

	find := utils.FindOptions{
		Filter: utils.Filter{"_id": bson.ObjectIdHex(userclaims["uuid"].(string))},
		Sort:   utils.Sort{"-created_at"},
		Limit:  1,
	}

	data, err := bson.MarshalJSON(find)
	if err != nil {
		return nil, err
	}

	users, err := utilNats.FindUser(n, data)
	if err != nil {
		return nil, err
	}
	if len(users) <= 0 {
		return nil, errors.New("invalid user")
	}

	return &users[0], nil
}
