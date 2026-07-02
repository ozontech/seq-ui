package profiles

import (
	"context"
	"sync"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/service"
)

type Profiles struct {
	idByName map[string]int64 // map UserName->UserProfileId
	mx       sync.RWMutex

	service service.Service
}

func New(svc service.Service) *Profiles {
	return &Profiles{
		idByName: make(map[string]int64),
		service:  svc,
	}
}

func (p *Profiles) GeIDFromContext(ctx context.Context) (int64, error) {
	userName, err := types.GetUserKey(ctx)
	if err != nil {
		return 0, err
	}

	id, err := p.GetID(userName)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (p *Profiles) GetID(userName string) (int64, error) {
	p.mx.RLock()
	id, ok := p.idByName[userName]
	p.mx.RUnlock()
	if ok {
		return id, nil
	}

	p.mx.Lock()
	defer p.mx.Unlock()
	id, ok = p.idByName[userName]
	if !ok {
		userProfile, err := p.service.GetOrCreateUserProfile(context.Background(), types.GetOrCreateUserProfileRequest{
			UserName: userName,
		})
		if err != nil {
			return 0, err
		}

		id = userProfile.ID
		p.idByName[userName] = id
	}

	return id, nil
}

func (p *Profiles) SetID(userName string, userProfileID int64) {
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
