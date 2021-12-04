package controllers

import (
	"context"
	"fmt"
	"log"

	"net/http"
	"time"

	"github.com/Phelickz/go-jwt-auth/database"
	"github.com/Phelickz/go-jwt-auth/helpers"
	"github.com/Phelickz/go-jwt-auth/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// creating or opening a user collection
var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

// creating a validator instance
var validate = validator.New()

// function to get a user
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		//getting the user_id from params
		userId := c.Param("user_id")

		//checking user Id
		err := helpers.MatchUserTypeToUid(c, userId)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		//user object from models
		var user models.User

		//finding user
		mongoErr := userCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
		defer cancel()
		if mongoErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": mongoErr.Error()})
			return
		}

		c.JSON(http.StatusOK, user)

	}
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		//creating variable user
		var user models.User

		//setting timeout
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		// fmt.Println("Here 1")
		defer cancel()

		//binding json
		err := c.BindJSON(&user)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// fmt.Println("Here 2 with", &user)

		//validating the user data with the validation requirements in the user struct
		validationErr := validate.Struct(&user)
		fmt.Println(validationErr)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			c.Abort()
			return
		}
		// fmt.Println("Here 3")

		// getting the user count because I want to show the count of the user too
		//using mongo func

		//using count to check if the user email exists already
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		// fmt.Println("Here 4")

		defer cancel()

		//hash password
		hashedPassword := HashPassword(*user.Password)
		user.Password = &hashedPassword

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		}
		// fmt.Println("Here 5")

		//checking if count is greater 0
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Email address already exists"})
			return
		}

		//creating created at
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		//generating tokens
		token, refreshToken, err := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, *&user.User_id)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		}
		user.Token = &token
		user.Refresh_token = &refreshToken

		//inserting user in db
		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "User item was not created"})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		//creating variable user
		var user models.User
		var foundUser models.User

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		defer cancel()

		err := c.BindJSON(&user)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//check if user exists
		dbErr := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()

		if dbErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": dbErr})
		}

		//verifyPassword
		correctPassword, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()

		if correctPassword == false {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}

		//generate tokens
		loginToken, refreshToken, tokenErr := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)

		if tokenErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		//update tokens in database
		helpers.UpdateAllTokens(loginToken, refreshToken, foundUser.User_id)

		foundUser.Token = &loginToken
		foundUser.Refresh_token = &refreshToken

		c.JSON(http.StatusOK, foundUser)

	}
}

func GetAllUsers() {

}

func HashPassword(password string) string {
	byte, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(byte)

}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("Password is incorrect")
		check = false
	}

	return check, msg
}
