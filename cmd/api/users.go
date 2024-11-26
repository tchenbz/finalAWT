package main

import (
    "errors"
    "net/http"

    "github.com/tchenbz/test3AWT/internal/data"
    "github.com/tchenbz/test3AWT/internal/validator"
)

func (a *applicationDependencies)registerUserHandler(w http.ResponseWriter, r *http.Request) {
// Get the passed in data from the request body and store in a temporary struct
   var incomingData struct {
       Username string  `json:"username"`
   	    Email  string  `json:"email"`
   	    Password  string  `json:"password"`
   }
   err := a.readJSON(w, r, &incomingData)
   if err != nil {
       a.badRequestResponse(w, r, err)
       return
   }
// we will add the password later after we have hashed it
   user := &data.User{
       	Username: incomingData.Username,
       	Email: incomingData.Email,
       	Activated: false,
   }
// hash the password and store it along with the cleartext version
err = user.Password.Set(incomingData.Password)
if err != nil {
	a.serverErrorResponse(w, r, err)
	return
}
// Perform validation for the User
v := validator.New()

data.ValidateUser(v, user)
if !v.IsEmpty() {
	a.failedValidationResponse(w, r, v.Errors)
	return
}
err = a.userModel.Insert(user)  
if err != nil {
	  switch {
		 case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			a.failedValidationResponse(w, r, v.Errors)
		 default:
			a.serverErrorResponse(w, r, err)
	  }
	 return
  }
  data := envelope {
	"user": user,
}

a.background(func() {
	err = a.mailer.Send(user.Email, "user_welcome.tmpl", user)
	if err != nil {
	   a.logger.Error(err.Error())
	}
  })
  

// Status code 201 resource created
err = a.writeJSON(w, http.StatusCreated, data, nil)
if err != nil {
	a.serverErrorResponse(w, r, err)
	return
}
}
