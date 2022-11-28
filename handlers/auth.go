package handlers

import (
	"context"
	"encoding/json"
	"log"
	"moniqq/models"
	"net/http"
	"time"

	"github.com/alexedwards/scs/mongodbstore"
	"github.com/alexedwards/scs/v2"
	"github.com/casbin/casbin/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var SessionManager *scs.SessionManager
var authEnforcer *casbin.Enforcer

func sessionAndAuth() {
	var err error
	authEnforcer, err = casbin.NewEnforcer("./auth_model.conf", "./policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	SessionManager = scs.New()
	SessionManager.IdleTimeout = 3 * time.Hour
	SessionManager.Store = mongodbstore.New(sdb)
	SessionManager.Cookie.Secure = true
}

// Function for authorization
func Authorizer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		role := SessionManager.GetString(r.Context(), "role")
		if role == "" {
			role = "anonymous"
		}
		// if it's an isp, check if the user still exists
		if role == "isp" {
			id, _ := primitive.ObjectIDFromHex(SessionManager.GetString(r.Context(), "id"))
			cursor := collUsers.FindOne(context.Background(), bson.M{"_id": id})
			if cursor.Err() != nil {
				http.Error(w, "FORBIDDEN", http.StatusForbidden)
				log.Println("ERROR: ", cursor.Err())
				return
			}
		}
		// casbin enforce
		res, err := authEnforcer.Enforce(role, r.URL.Path, r.Method)
		if err != nil {
			http.Error(w, "ERROR", http.StatusInternalServerError)
			log.Println("ERROR: ", err)
			return
		}
		if res {
			next.ServeHTTP(w, r)
		}
		if !res {
			http.Error(w, "FORBIDDEN", http.StatusForbidden)
			log.Println("ERROR: unauthorized")
			return
		}
	}
	return http.HandlerFunc(fn)
}

// Function of the endpoint for sign in
func Login(w http.ResponseWriter, r *http.Request) {
	var foundUser models.User
	var authData models.Identity

	username := r.PostFormValue("username")
	pass := r.PostFormValue("password")

	cursor := collUsers.FindOne(context.TODO(), bson.M{"username": username})
	if cursor.Err() != nil {
		http.Error(w, "Username atau Password salah", http.StatusBadRequest)
		return
	}
	if err := cursor.Decode(&foundUser); err != nil {
		http.Error(w, "ERROR", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	passIsValid, msg := verifyPassword(pass, foundUser.Password)
	if !passIsValid {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	SessionManager.Put(r.Context(), "id", foundUser.ID.Hex())
	SessionManager.Put(r.Context(), "role", foundUser.Role)

	if foundUser.Role == "isp" {
		SessionManager.Put(r.Context(), "isp_name", foundUser.Isp_Name)
		authData.Isp_Name = foundUser.Isp_Name
	}

	log.Println(
		"User with role:", foundUser.Role, "(UserID:"+foundUser.ID.Hex()+")", "is logged in",
	)

	authData.Role = foundUser.Role

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(authData)
}

// Function of the endpoint for sign out
func Logout(w http.ResponseWriter, r *http.Request) {
	id := SessionManager.GetString(r.Context(), "id")
	role := SessionManager.GetString(r.Context(), "role")
	if err := SessionManager.Destroy(r.Context()); err != nil {
		http.Error(w, "ERROR", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	log.Println("User with role:", role, "(UserID:"+id+")", "has logged out")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("LOGOUT SUCCESS"))
}

// Function of the endpoint for check the role of user
func WhoAmI(w http.ResponseWriter, r *http.Request) {
	var iam models.Identity

	iam.Role = SessionManager.GetString(r.Context(), "role")
	if iam.Role == "" {
		iam.Role = "anonymous"
	} else if iam.Role == "isp" {
		iam.Isp_Name = SessionManager.GetString(r.Context(), "isp_name")
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(iam)
}

// HELPER
func verifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "Username atau Password salah"
		check = false
	}

	return check, msg
}
