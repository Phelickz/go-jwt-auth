package helpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Phelickz/go-jwt-auth/database"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

func UpdateAllTokens(token string, refreshToken string, userId string) {
	//find and update user

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	// defer cancel()

	// var updateObj primitive.D

	// updateObj = append(updateObj, bson.E{"Token": token})
	// updateObj = append(updateObj, bson.E{"Refresh_token": refreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	// updateObj = append(updateObj, bson.E{"updated_at", Updated_at})

	upsert := true
	// filter := bson.M{"user_id", userId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	_, err := userCollection.UpdateOne(ctx, bson.M{"user_id": userId}, bson.D{{"$set", bson.M{
		"token":         token,
		"refresh_token": refreshToken,
		"updated_at":    Updated_at,
	}}}, &opt)

	defer cancel()
	// fmt.Println(result)

	if err != nil {
		log.Panic(err)
		return
	}
}

func ValidateToken(signedToken string) (claims *SignedDetails, msg string) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&SignedDetails{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(SECRET_KEY), nil
		},
	)

	if err != nil {
		msg = err.Error()
		return
	}

	claims, ok := token.Claims.(*SignedDetails)

	if !ok {
		msg = "Invalid token"
		msg = err.Error()
		return
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		msg = fmt.Sprintf("Token is expired")
		msg = err.Error()
		return
	}

	return claims, msg
}
