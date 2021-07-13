package session

import (
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
	profile, err := s.iam.ProfileGet(token.Session)

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

// список всех токенов для всех пользователей доступных для сервиса
func (s *session) List() (result map[string]SessionRec)  {
	s.Registry.Mx.Lock()
	defer s.Registry.Mx.Unlock()

	result = s.Registry.M

	return result
}