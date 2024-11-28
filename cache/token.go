package cache

import (
	"github.com/patrickmn/go-cache"
	"github.com/prok05/ecom/service/alpha"
	"log"
	"time"
)

type TokenCache struct {
	cache  *cache.Cache
	expiry time.Duration
}

func NewTokenCache(expiry time.Duration) *TokenCache {
	c := cache.New(expiry, expiry)
	return &TokenCache{
		cache:  c,
		expiry: expiry,
	}
}

func (tc *TokenCache) GetToken() (string, error) {
	token, found := tc.cache.Get("alpha_token")
	if found {
		return token.(string), nil
	}

	newToken, err := alpha.GetAlphaToken()
	if err != nil {
		return "", err
	}

	tc.cache.Set("alpha_token", newToken, tc.expiry)
	log.Println("Token restored")
	return newToken, nil
}
