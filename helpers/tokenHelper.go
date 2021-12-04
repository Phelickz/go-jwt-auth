package helpers

import (
	"log"
	"os"
	"time"

	"github.com/Phelickz/go-jwt-auth/database"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/mongo"
)

//creating struct for the jwt model
type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

//accessing collection
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

//getting secretKey
var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstname string, lastname string, userType string, userId string) (token string, refreshToken string, err error) {

	claims := &SignedDetails{
		Email:      email,
		First_name: firstname,
		Last_name:  lastname,
		Uid:        userId,
		User_type:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(), //setting expiring to be 24 hours
		},
	}

	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(700)).Unix(),
		},
	}

	tokenJWT, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refreshTokenJWT, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return tokenJWT, refreshTokenJWT, err

}
