package controllers

import (
	"context"
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
)

// creating or opening a user collection
var userCollection *mongo.Collection = (*mongo.Collection)(database.OpenCollection(database.Client, "user"))

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

		//binding json
		err := c.BindJSON(&user)
		defer cancel()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//validating the user data with the validation requirements in the user struct
		validationErr := validate.Struct(user)

		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// getting the user count because I want to show the count of the user too
		//using mongo func

		//using count to check if the user email exists already
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		}

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

func Login() {}

func GetAllUsers() {

}

func HashPassword() {

}

func VerifyPassword() {}
