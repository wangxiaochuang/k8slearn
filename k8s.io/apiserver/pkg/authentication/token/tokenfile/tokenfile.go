package tokenfile

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"k8s.io/apiserver/pkg/authentication/authenticator"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/klog/v2"
)

type TokenAuthenticator struct {
	tokens map[string]*user.DefaultInfo
}

func New(tokens map[string]*user.DefaultInfo) *TokenAuthenticator {
	return &TokenAuthenticator{
		tokens: tokens,
	}
}

func NewCSV(path string) (*TokenAuthenticator, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	recordNum := 0
	tokens := make(map[string]*user.DefaultInfo)
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(record) < 3 {
			return nil, fmt.Errorf("token file '%s' must have at least 3 columns (token, user name, user uid), found %d", path, len(record))
		}

		recordNum++
		if record[0] == "" {
			klog.Warningf("empty token has been found in token file '%s', record number '%d'", path, recordNum)
			continue
		}

		obj := &user.DefaultInfo{
			Name: record[1],
			UID:  record[2],
		}
		if _, exist := tokens[record[0]]; exist {
			klog.Warningf("duplicate token has been found in token file '%s', record number '%d'", path, recordNum)
		}
		tokens[record[0]] = obj

		if len(record) >= 4 {
			obj.Groups = strings.Split(record[3], ",")
		}
	}

	return &TokenAuthenticator{
		tokens: tokens,
	}, nil
}

func (a *TokenAuthenticator) AuthenticateToken(ctx context.Context, value string) (*authenticator.Response, bool, error) {
	panic("not implemented")
}
