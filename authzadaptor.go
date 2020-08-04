package authzadaptor

import (
	"crypto/ecdsa"
	"crypto/tls"
	fmt "fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	strings "strings"
	"time"

	"github.com/gogo/googleapis/google/rpc"
	"golang.org/x/net/context"

	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"istio.io/api/mixer/adapter/model/v1beta1"
	"istio.io/istio/userkey/config"
)

type AuthZAdaptor struct {
	URLToPublicKeyDict map[string]*ecdsa.PublicKey
}

func (authZAdaptor AuthZAdaptor) HandleAuthzadaptor(_ context.Context, req *HandleAuthzadaptorRequest) (*HandleAuthzadaptorResponse, error) {
	config := &config.Params{}
	if err := config.Unmarshal(req.AdapterConfig.Value); err != nil {
		return nil, err
	}

	if req.Instance.Key == "unknown" {
		log.Warnf("skip unknown header: %v", req.Instance.Key)
		return &HandleAuthzadaptorResponse{
			Result: &v1beta1.CheckResult{
				ValidDuration: config.ValidDuration,
				// Status: rpc.Status{Code: int32(rpc.PERMISSION_DENIED)},
			},
		}, nil
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(req.Instance.Key, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		url := constructURLFromTokenHeader(token.Header)
		if val, ok := authZAdaptor.URLToPublicKeyDict[url]; ok {
			return val, nil
		}

		publicKey, err := fetchPublicKeyFromURL(url)
		if err != nil {
			log.Warnf("Can not fetch public key from %s, error: %s", url, err)
			return nil, err
		}

		pKey, err := convertKey(publicKey)
		if err != nil {
			log.Warnf("Can not convert pem key to *ecdsa.PublicKey %s, error: %s", publicKey, err)
			return nil, err
		}

		authZAdaptor.URLToPublicKeyDict[url] = pKey
		return pKey, nil
	})

	if err != nil {
		log.Warnf("Can not verify header: %v, err: %v", req.Instance.Key, err)
		return &HandleAuthzadaptorResponse{
			Result: &v1beta1.CheckResult{
				Status: rpc.Status{Code: int32(rpc.PERMISSION_DENIED)},
			},
		}, nil
	}

	email, ok := claims["email"]
	if !ok {
		log.Errorf("Email doesn't exist in user's claim %v. This is not supported", claims)
		return &HandleAuthzadaptorResponse{
			Result: &v1beta1.CheckResult{
				Status: rpc.Status{Code: int32(rpc.PERMISSION_DENIED)},
			},
		}, nil
	}

	// Only verify this field if it exists.
	if val, ok := claims["email_verified"]; ok {
		isEmailVerified := false

		switch v := val.(type) {
		case bool:
			isEmailVerified = v
		case string:
			b, err := strconv.ParseBool(v)
			if err != nil {
				log.Warnf("Can not parse email_verified %v to bool", val)
			} else {
				isEmailVerified = b
			}
		default:
			log.Warnf("Unknown email_verified type %v, value %v", reflect.TypeOf(v), v)
		}

		if !isEmailVerified {
			log.Errorf("Email is not verified %v", isEmailVerified)
			return &HandleAuthzadaptorResponse{
				Result: &v1beta1.CheckResult{
					Status: rpc.Status{Code: int32(rpc.PERMISSION_DENIED)},
				},
			}, nil
		}
	}

	log.Infof("Return email %v", email)
	return &HandleAuthzadaptorResponse{
		Result: &v1beta1.CheckResult{ValidDuration: config.ValidDuration},
		Output: &OutputMsg{Email: email.(string)},
	}, nil
}

// constructURLFromTokenHeader retrieve region and kid from headers and construct url to fetch public key
func constructURLFromTokenHeader(header map[string]interface{}) string {
	kid := header["kid"]
	signer := header["signer"].(string)
	region := strings.SplitN(signer, ":", -1)[3]
	return fmt.Sprintf("https://public-keys.auth.elb.%s.amazonaws.com/%s", region, kid)
}

// https client to get public key pem from a give url
func fetchPublicKeyFromURL(url string) ([]byte, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   15 * time.Second,
		Transport: tr,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	return body, err
}

// convertKey transform pem public key to *ecdsa.PublicKey
func convertKey(publicKey []byte) (*ecdsa.PublicKey, error) {
	key, err := jwt.ParseECPublicKeyFromPEM(publicKey)
	if err != nil {
		return nil, err
	}

	return key, nil
}
