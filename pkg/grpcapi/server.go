package grpcapi

import (
	"andreyladmj/gAuth/pkg/grpcapi/userpb"
	"andreyladmj/gAuth/pkg/models"
	"andreyladmj/gAuth/pkg/models/mysql"
	"context"
	"fmt"
	"log"
)

type GRPCServer struct{
	UserModel *mysql.UserModel
	ErrorLog *log.Logger
	InfoLog *log.Logger
}

func (s *GRPCServer) GetUser(ctx context.Context, req *userpb.AuthRequest) (*userpb.AuthResponse, error) {
	fmt.Printf("GetUser function was invoked with %v\n", req)
	token := req.GetToken()
	var userObj *userpb.User

	user, err := s.UserModel.GetUserByToken(token)
	errorCode := 0
	errorText := ""
	userObj = nil

	if err != nil {
		if err == models.ErrNoRecord {
			errorCode = 404
			errorText = "Not Found"
		} else {
			errorCode = 500
			errorText = "Internal Server Error"
			s.ErrorLog.Printf("GRPC GetUserByToken error %v", err)
		}
	} else {
		userObj = &userpb.User{
			Name:    user.Name,
			Email:   user.Email,
			Picture: user.Picture,
			Gender:  user.Gender.String,
			Locale:  user.Locale,
			Created: user.Created.Format("2006-01-02T15:04:05"),
		}
	}

	fmt.Println("err", err)
	fmt.Println("errorCode", errorCode)
	fmt.Println("errorText", errorText)
	fmt.Println("user", user)
	fmt.Println("userObj", userObj)

	res := &userpb.AuthResponse{
		User: userObj,
		Error: errorText,
		Status: int32(errorCode),
	}

	return res, nil
}
