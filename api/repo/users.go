package repo

import (
	"errors"
	"fmt"
	"rethink/api/db"
	"rethink/api/models"

	r "github.com/rethinkdb/rethinkdb-go"
)

// UserController struct handles database interactions
type UserController struct {
	session *r.Session
}

// NewUserController initializes the UserController with a GORM DB instance
func NewUserController(session *r.Session) *UserController {
	return &UserController{session: session}
}

// Getter function to expose the session
func (uc *UserController) GetSession() *r.Session {
	return uc.session
}

// AddUser inserts a new user into the database
func (uc *UserController) AddUser(user models.AppUser) (models.AppUser, error) {
	// Check if user already exists
	cursor, err := r.DB("taipan").Table("users").Filter(r.Row.Field("email").Eq(user.Email)).Run(uc.session)
	if err != nil {
		return models.AppUser{}, err
	}
	defer cursor.Close()

	// Hash the password before saving
	user.Password = (user.Password)

	// Save user
	_, err = r.Table("users").Insert(user).RunWrite(uc.session)
	if err != nil {
		return models.AppUser{}, err
	}

	return user, nil
}

// GetAllUsers fetches all users from the database
func (uc *UserController) GetAllUsers() ([]models.AppUser, error) {
	cursor, err := r.Table("users").Run(uc.session)
	if err != nil {
		return nil, err
	}
	defer cursor.Close()

	var users []models.AppUser
	if err := cursor.All(&users); err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserByID fetches a user from the database by ID
func (uc *UserController) GetUserByEmail(Email string) (*models.AppUser, error) {
	var user models.AppUser

	err := r.DB("taipan").Table("users").Filter(r.Row.Field("Email").Eq(Email)).ReadOne(&user, db.DB)
	if err != nil {
		fmt.Println("Error fetching from DB : ", err)
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates user information
func (uc *UserController) UpdateUser(Email string, updatedUser models.AppUser) error {

	// Fetch the existing user
	existingUser, err := uc.GetUserByEmail(Email)
	if err != nil {
		return err // Return error if user is not found
	}

	// Ensure `createdat` is not overwritten by preserving its original value
	updatedUser.CreatedAt = existingUser.CreatedAt
	updatedUser.Userid = existingUser.Userid

	_, err = r.Table("users").Filter(r.Row.Field("Email").Eq(Email)).Update(updatedUser, r.UpdateOpts{NonAtomic: true}).RunWrite(uc.session)
	if err != nil {
		return err
	}
	return nil
}

// DeleteUser removes a user from the database
func (uc *UserController) DeleteUser(Email string) error {
	// Check if the user exists
	cursor, err := r.Table("users").Filter(r.Row.Field("Email").Eq(Email)).Run(uc.session)
	if err != nil {
		return fmt.Errorf("error fetching user: %v", err)
	}
	defer cursor.Close()

	if cursor.IsNil() {
		return fmt.Errorf("user not found")
	}

	//delete
	res, err := r.Table("users").Filter(r.Row.Field("Email").Eq(Email)).Delete().RunWrite(uc.session)
	if err != nil {
		return err
	}

	// Check if any rows were actually deleted
	if res.Deleted == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// GetUserRoleByEmail fetches the role of a user by email
func (uc *UserController) GetUserRoleByEmail(Email string) (string, error) {

	var user models.AppUser
	cursor, err := r.Table("users").
		Filter(r.Row.Field("Email").Eq(Email)).
		Pluck("Role").
		Run(uc.session)

	if err != nil {
		return "", err
	}
	defer cursor.Close()

	if cursor.Next(&user) {
		return user.Role, nil
	}

	return "", errors.New("user role not found")
}

// HasPermission checks if a role has the required permission
func (uc *UserController) HasPermission(role, permission string) (bool, error) {

	cursor, err := r.Table("access").
		Filter(r.Row.Field("Role").Eq(role).And(r.Row.Field("privilege").Eq(permission))).
		Count().
		Run(uc.session)

	if err != nil {
		return false, err
	}
	defer cursor.Close()

	var permissioncount int

	if cursor.Next(&permissioncount) && permissioncount > 0 {
		return true, nil
	}

	return false, nil
}
