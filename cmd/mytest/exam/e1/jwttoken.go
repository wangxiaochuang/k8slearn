package e1

import (
	"time"

	"gopkg.in/square/go-jose.v2/jwt"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/kubernetes/pkg/serviceaccount"
	"k8s.io/utils/wxc"
)

type Aaa struct {
	Name string `json:"name"`
}

func main() {
	sk, _ := keyutil.PrivateKeyFromFile("cert-dir/kube-serviceaccount.key")

	g, _ := serviceaccount.JWTTokenGenerator("https://kubernetes.default.svc", sk)
	then := time.Now().Add(-2 * time.Hour)
	sc := &jwt.Claims{
		Subject:   "wxc",
		Audience:  jwt.Audience([]string{"api"}),
		IssuedAt:  jwt.NewNumericDate(then),
		NotBefore: jwt.NewNumericDate(then),
		Expiry:    jwt.NewNumericDate(then.Add(time.Duration(60*60) * time.Second)),
	}
	wxc.P(g.GenerateToken(sc, Aaa{"wxc"}))
}
