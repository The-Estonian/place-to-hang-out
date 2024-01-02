package dbconnections

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"forum/structs"
	"forum/validateData"
)

// Open and Close database connection. Returns database connection.
func DbConnection() *sql.DB {
	db, err := sql.Open("sqlite3", "./database/forum.db")
	validateData.CheckErr(err)
	return db
}

// Register new user, add user values to users database. Returns true if user/email is database.
func RegisterUser(username, email, password string) (bool, bool) {
	db := DbConnection()
	usernameCheck := CheckValueFromDB("username", username)
	emailCheck := CheckValueFromDB("email", email)
	if !usernameCheck && !emailCheck {
		_, err := db.Exec("INSERT INTO users(username, password, email) VALUES(?, ?, ?)", username, password, email)
		validateData.CheckErr(err)
		fmt.Println("New user added to the DB")
		SetAccessRight(GetID(username), "2")
		fmt.Println("Access granted to user", GetID(username))
	}
	defer db.Close()
	return usernameCheck, emailCheck
}

// Returns True if user inserted credentials are in database.
func LoginUser(username, password string) bool {
	getPassword := CheckPassword(username)
	return getPassword == password
}

// Deletes Cookie from session database
func LogoutUser(hash string) {
	db := DbConnection()
	_, err := db.Exec("DELETE FROM session WHERE hash=?", hash)
	if err != nil {
		fmt.Println("LogoutUser")
		fmt.Println("Error code:", err)
	}
	defer db.Close()
}

// Applies Cookie in session database
func ApplyHash(user, hash string) {
	db := DbConnection()
	_, err := db.Exec("INSERT OR REPLACE INTO session(user, hash) VALUES(?, ?)", user, hash)
	defer db.Close()
	validateData.CheckErr(err)
}

// Returns ID if username exists in the users database
func GetID(username string) string {
	db := DbConnection()
	query := db.QueryRow("SELECT id FROM users WHERE username=?", username).Scan(&username)
	defer db.Close()
	if query != nil {
		fmt.Println("Didn't find username with that name to return ID")
		fmt.Println("Error code: ", query)
	}
	return username
}

// Returns All user info by user hash
func GetUserInfo(id string) structs.User {
	db := DbConnection()
	var userInfo structs.User
	var dump string
	query := db.QueryRow("SELECT * FROM users WHERE id=?", id).Scan(&userInfo.Id, &userInfo.Username, &dump, &userInfo.Email)
	dump = ""
	defer db.Close()
	if query != nil {
		fmt.Println("Didn't find userId with that id")
		fmt.Println("Error code: ", query)
	}
	return userInfo
}

// Returns the users access rights
func GetAccessRight(id string) structs.AccessRights {
	db := DbConnection()
	var userAccess structs.AccessRights
	query := db.QueryRow("SELECT user_access FROM user_access WHERE user=?", id).Scan(&userAccess.AccessRight)
	defer db.Close()
	if query != nil {
		fmt.Println("Didn't find user with that id")
		fmt.Println("Error code: ", query)
	}
	return userAccess
}

// Sets the users access rights
func SetAccessRight(user string, access string) {
	db := DbConnection()
	_, err := db.Exec("INSERT INTO user_access (user, user_access) VALUES(?, ?)", user, access)
	if err != nil {
		fmt.Println("SetAccessRight")
		fmt.Println("Error code: ", err)
	}
	defer db.Close()
}

// Returns UserID from session database
func CheckHash(hash string) string {
	db := DbConnection()
	var user string
	var date string
	query := db.QueryRow("SELECT user, date FROM session WHERE hash=?", hash).Scan(&user, &date)
	defer db.Close()
	if query == nil {
		layout := "2006-01-02 15:04:05"
		hashDate, _ := time.Parse(layout, date)
		currTime := time.Now().Add(time.Minute * -10)
		if currTime.Sub(hashDate) > 0 {
			fmt.Println("Hash expired, Executing delete")
			LogoutUser(hash)
			return "1"
		}
	}
	if query != nil {
		return "1"
	}
	return user
}

// Returns True if hash in database
func HashInDatabase(hash string) bool {
	db := DbConnection()
	var user string
	query := db.QueryRow("SELECT user FROM session WHERE hash=?", hash).Scan(&user)
	defer db.Close()
	if query != nil {
		fmt.Println("HashInDatabase: didn't find user with that hash!")
		fmt.Println("Error code: ", query)
		return false
	}
	return true
}

// Return True if value is in users database column
func CheckValueFromDB(column string, valueToCheck string) bool {
	db := DbConnection()
	newUsername := db.QueryRow("SELECT "+column+" FROM users WHERE "+column+"=?", valueToCheck).Scan(&valueToCheck)
	defer db.Close()
	trigger := false
	if newUsername == nil {
		fmt.Println("Username already exists!")
		trigger = true
	}
	return trigger
}

// Returns password from users based on username. Hopefully encrypted.
func CheckPassword(username string) string {
	db := DbConnection()
	var returnPassword string
	err := db.QueryRow("SELECT password FROM users WHERE username=?", username).Scan(&returnPassword)
	defer db.Close()
	if err != nil {
		fmt.Println("User does not exist")
		return ""
	}
	return returnPassword
}

// Returns all rows in an array of structs from posts database
func GetAllPosts(data string) []structs.Post {
	var allPosts []structs.Post
	if len(data) > 0 {
		allPosts = append(allPosts, GetOnePost(data))
		return allPosts
	}
	db := DbConnection()
	rows, _ := db.Query("SELECT posts.id, title, posts.user, posts.post, created, username, SUM(post_like) FROM posts INNER JOIN users ON posts.user = users.id INNER JOIN post_likes ON posts.id = post_likes.post GROUP BY posts.id, title, posts.user, posts.post, created, username")
	defer db.Close()
	for rows.Next() {
		var post structs.Post
		if err := rows.Scan(&post.Id, &post.Title, &post.User, &post.Post, &post.Created, &post.Username, &post.LikeRating); err != nil {
			fmt.Println("Closing connection")
			rows.Close()
			return allPosts
		}
		layout := "2006-01-02 15:04:05"
		postDate, _ := time.Parse(layout, post.Created)
		post.Created = time.Since(postDate).Truncate(time.Second).String()
		allPosts = append(allPosts, post)
	}
	defer rows.Close()
	return allPosts
}

// Returns a struct that contains data from one row in post database
func GetOnePost(data string) structs.Post {
	db := DbConnection()
	posts := db.QueryRow("SELECT * FROM posts WHERE id=?", data)
	defer db.Close()
	var post structs.Post
	if err := posts.Scan(&post.Id, &post.Title, &post.User, &post.Post, &post.Created); err != nil {
		fmt.Println(err)
	}
	return post
}

// Inserts into posts and post_category_list user inserted data
func InsertMessage(r *http.Request, userId string) {
	r.ParseForm()
	userForm := r.Form
	db := DbConnection()
	var inputTitle string
	var inputMessage string
	var catArray []string

	for key, value := range userForm {
		if key == "title" {
			inputTitle = value[0]
		} else if key == "message" {
			inputMessage = value[0]
		} else {
			catArray = append(catArray, key)
		}
	}

	var data int
	err := db.QueryRow("INSERT INTO posts (title, user, post) VALUES (?, ?, ?) RETURNING id", inputTitle, userId, inputMessage).Scan(&data)
	if err != nil {
		fmt.Println(err)
	}

	_, err = db.Exec("INSERT INTO post_likes (post, user, post_like) VALUES (?, ?, ?)", data, "1", "0")
	if err != nil {
		fmt.Println(err)
	}

	for _, v := range catArray {
		_, err = db.Exec("INSERT INTO post_category_list (post_category, post_id) VALUES (?, ?)", v, data)
		if err != nil {
			fmt.Println(err)
		}
	}
	defer db.Close()
}

// Inserts into comments database user inserted comment and commentator userID
func InsertComment(postId string, commentatorId string, comment string) {
	db := DbConnection()
	_, err := db.Exec("INSERT INTO comments (post_id, user, comment) VALUES (?, ?, ?)", postId, commentatorId, comment)
	defer db.Close()
	if err != nil {
		fmt.Println(err)
	}
}

// Returns all comments by post_id
func GetAllComments(data string) []structs.Comment {
	db := DbConnection()
	var allComments []structs.Comment
	allCommentsFromData, _ := db.Query("SELECT * FROM comments WHERE post_id=?", data)
	defer db.Close()
	for allCommentsFromData.Next() {
		var comments structs.Comment
		if err := allCommentsFromData.Scan(&comments.Id, &comments.PostId, &comments.UserId, &comments.Comment, &comments.Created); err != nil {
			fmt.Println(err)
		}
		allComments = append(allComments, comments)
	}
	defer allCommentsFromData.Close()
	return allComments

}

func GetAllCategories() []structs.Category {
	db := DbConnection()
	var allCategories []structs.Category
	data, _ := db.Query("SELECT * FROM category")
	defer db.Close()
	for data.Next() {
		var category structs.Category
		if err := data.Scan(&category.Id, &category.Category); err != nil {
			fmt.Println(err)
			return allCategories
		}
		allCategories = append(allCategories, category)
	}
	defer data.Close()
	return allCategories
}

// func GetPostLikes() {
// 	db := DbConnection()
// 	data, err := db.Query("SELECT user, COUNT(DISTINCT post_like) FROM post_likes")
// 	if err != nil {
// 		fmt.Println("ERROR", err)
// 	}
// 	for data.Next() {
// 		var stream string
// 		var stream2 string
// 		var stream3 string
// 		data.Scan(&stream, &stream2)
// 		fmt.Println("STREAM", stream)
// 		fmt.Println("STREAM2", stream2)
// 		fmt.Println("STREAM3", stream3)
// 	}
// 	defer db.Close()
// }

func SetPostLikes(userId, postId, like string) {
	db := DbConnection()
	var postLike structs.PostLikes
	db.QueryRow("SELECT * FROM post_likes WHERE post=? AND user=?", postId, userId).Scan(&postLike.Id, &postLike.Post, &postLike.User, &postLike.PostLike)
	if postLike.Id == "" {
		_, err := db.Exec("INSERT INTO post_likes (post, user, post_like) VALUES (?, ?, ?)", postId, userId, like)
		if err != nil {
			fmt.Println("New: Error inserting like to table")
		}
	} else {
		_, err := db.Exec("REPLACE INTO post_likes (post, user, post_like) VALUES (?, ?, ?)", postId, userId, like)
		if err != nil {
			fmt.Println("Exists: Error inserting like to table")
		}
	}
	defer db.Close()
}

func GetCommentLikes() {

}

func SetCommentLikes() {

}

func GetMegaDataValues(r *http.Request, handler string) structs.MegaData {
	var userId string
	cookie, err := r.Cookie("UserCookie")
	if err != nil {
		userId = "1"
	} else {
		userId = CheckHash(cookie.Value)
	}

	postId := r.URL.Query().Get("PostId")
	m := structs.MegaData{
		User:   GetUserInfo(userId),
		Access: GetAccessRight(userId),
	}
	if handler == "Forum" {
		m.AllPosts = GetAllPosts(postId)
	}
	if handler == "Post" {
		m.Categories = GetAllCategories()
	}
	if handler == "PostContent" {
		m.AllPosts = GetAllPosts(postId)
		m.AllComments = GetAllComments(postId)
	}
	return m
}
