package session

import (
	"fmt"
	"github.com/buildboxapp/app/pkg/model"
	"time"
)

func (s *session) Get(sessionID string) (profile model.ProfileData, err error)  {
	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	profile = s.Registry.M[sessionID].Profile

	return profile, err
}

func (s *session) Delete(sessionID string) (err error)  {
	if sessionID == "" {
		return err
	}

	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	delete(s.Registry.M, sessionID)

	return err
}

func (s *session) Set(token *model.Token) (err error)  {
	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	if s.Registry.M == nil {
		s.Registry.M = map[string]SessionRec{}
	}
	expiration := time.Now().Add(300 * time.Hour)
	profile, err := s.Profile(token)

	if err != nil {
		return err
	}

	if _, found := s.Registry.M[token.Session]; !found {
		var f = SessionRec{}
		f.Profile = profile
		f.DeadTime = expiration.Unix()

		s.Registry.M[token.Session] = f
	}

	return err
}

// генерируем профиль пользователя
func (s *session) Profile(token *model.Token) (profile model.ProfileData, err error)  {
	userID := token.Uid
	objUser, err := s.api.ObjGet(userID)

	if len(objUser.Data) == 0 {
		return profile, fmt.Errorf("%s", "Error. Object user is empty.")
	}

	profile.Uid = objUser.Data[0].Uid
	profile.Country, _ 	= objUser.Data[0].Attr("country", "value")
	profile.City, _ 	= objUser.Data[0].Attr("city", "value")
	profile.Age, _ 		= objUser.Data[0].Attr("age", "value")
	profile.Photo, _ 	= objUser.Data[0].Attr("photo", "value")
	profile.Last_name, _ = objUser.Data[0].Attr("last_name", "value")
	profile.First_name, _ = objUser.Data[0].Attr("first_name", "value")
	profile.Email, _ 	= objUser.Data[0].Attr("email", "value")

	// 2 берем текущий профиль пользователя
	objProfile, err := s.api.ObjGet(token.Profile)
	if len(objProfile.Data) != 0 {
		profile.CurrentProfile = objProfile.Data[0]
	}

	// 3 берем текущую роль пользователя
	objRole, err := s.api.ObjGet(token.Role)
	if len(objRole.Data) != 0 {
		profile.CurrentRole = objRole.Data[0]
	}

	return profile, err
}

// список всех токенов для всех пользователей доступных для сервиса
func (s *session) List() (result map[string]SessionRec)  {
	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	result = s.Registry.M

	return result
}