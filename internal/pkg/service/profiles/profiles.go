package profiles

import (
	"context"
	"sync"

	"github.com/ozontech/seq-ui/internal/app/types"
)

type userProfileService interface {
	GetOrCreateUserProfile(context.Context, types.GetOrCreateUserProfileRequest) (types.UserProfile, error)
}

var profile *profiles

func InitProfiles(svc userProfileService) {
	profile = &profiles{
		idByName: make(map[string]int64),
		service:  svc,
	}
}

type profiles struct {
	idByName map[string]int64 // map UserName->UserProfileId
	mx       sync.RWMutex

	service userProfileService
}

func GetIDFromContext(ctx context.Context) (int64, error) {
	return profile.getIDFromContext(ctx)
}

func SetID(userName string, userProfileID int64) {
	profile.setID(userName, userProfileID)
}

func (p *profiles) getIDFromContext(ctx context.Context) (int64, error) {
	userName, err := types.GetUserKey(ctx)
	if err != nil {
		return 0, err
	}

	id, err := p.getID(userName)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *profiles) getID(userName string) (int64, error) {
	p.mx.RLock()
	id, ok := p.idByName[userName]
	p.mx.RUnlock()
	if ok {
		return id, nil
	}

	userProfile, err := p.service.GetOrCreateUserProfile(context.Background(), types.GetOrCreateUserProfileRequest{UserName: userName})
	if err != nil {
		return 0, err
	}

	p.mx.Lock()
	p.idByName[userName] = userProfile.ID
	p.mx.Unlock()

	return userProfile.ID, nil
}

func (p *profiles) setID(userName string, userProfileID int64) {
	p.mx.RLock()
	_, ok := p.idByName[userName]
	p.mx.RUnlock()
	if ok {
		return
	}

	p.mx.Lock()
	defer p.mx.Unlock()
	_, ok = p.idByName[userName]
	if !ok {
		p.idByName[userName] = userProfileID
	}
}
