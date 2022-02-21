package argo

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	"greenops.io/workflowtrigger/apikeysmanager"

	"github.com/argoproj/argo-cd/pkg/apiclient/account"
	"google.golang.org/grpc/status"
)

type RbacResource string
type RbacAction string

const (
	ClusterResource     RbacResource = "clusters"
	ApplicationResource RbacResource = "applications"

	SyncAction     RbacAction = "sync"
	CreateAction   RbacAction = "create"
	UpdateAction   RbacAction = "update"
	GetAction      RbacAction = "get"
	DeleteAction   RbacAction = "delete"
	ActionAction   RbacAction = "action"
	OverrideAction RbacAction = "override"
)

var AllowedApiKeysEndpointRegex = []string{
	`\/client\/generateNotification\/.*$`,
	`\/client\/.*\/.*\/generateEvent$`,
}

type ArgoAuthenticatorApi interface {
	CheckRbacPermissions(action RbacAction, resource RbacResource, subresource string) bool
}

func (a *ArgoApiImpl) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.CheckApiKey(r) {
			next.ServeHTTP(w, r)
		}

		token := r.Header.Get("Authorization")
		splitToken := strings.Split(token, "Bearer ")
		token = splitToken[1]
		a.configureArgoClient(token)

		defer func() {
			if err := recover(); err != nil {
				switch err.(type) {
				case string:
					if strings.Contains(err.(string), "code = Unauthenticated") {
						log.Printf("User is unauthenticated")
						http.Error(w, "User is unauthenticated", http.StatusUnauthorized)
					} else {
						log.Printf("Interal error occurred: %s", err.(string))
						http.Error(w, err.(string), http.StatusInternalServerError)
					}
				case error:
					st, ok := status.FromError(err.(error))
					if ok {
						if st.Code() == 16 {
							log.Printf("User is unauthenticated")
							http.Error(w, "User is unauthenticated", http.StatusUnauthorized)
						}
					} else {
						log.Printf("Interal error occurred: %s", err.(error).Error())
						http.Error(w, err.(error).Error(), http.StatusInternalServerError)
					}
				default:
					http.Error(w, err.(string), http.StatusInternalServerError)
				}
			}
			return
		}()
		a.CheckRbacPermissions(SyncAction, ClusterResource, "abc")
		next.ServeHTTP(w, r)
	})
}

func (a *ArgoApiImpl) CheckRbacPermissions(action RbacAction, resource RbacResource, subresource string) bool {
	closer, client, err := a.configuredClient.NewAccountClient()
	if err != nil {
		log.Fatalf("account client could not be made for Argo: %s", err)
	}
	defer closer.Close()

	canI, err := client.CanI(context.TODO(), &account.CanIRequest{
		Resource:    string(resource),
		Action:      string(action),
		Subresource: subresource,
	})
	if err != nil {
		panic(fmt.Sprintf("Request to Argo server failed: %s", err))
	}
	return canI.Value == "yes"
}

func (a *ArgoApiImpl) CheckApiKey(r *http.Request) bool {
	var pathAllowed bool
	for _, regex := range AllowedApiKeysEndpointRegex {
		matched, err := regexp.MatchString(regex, r.URL.Path)
		if err != nil {
			return false
		}
		if matched {
			pathAllowed = true
		}
	}
	if !pathAllowed {
		return false
	}

	apikeyStrings := r.Header[apikeysmanager.ApiKeyHeaderName]
	if apikeyStrings == nil || len(apikeyStrings) < 1 {
		return false
	}
	verified, err := a.apikeysManager.VerifyRequest(apikeyStrings[0])
	if err != nil {
		log.Println("failed to verify request with apikey, error: ", err.Error())
		return false
	}
	return verified
}
