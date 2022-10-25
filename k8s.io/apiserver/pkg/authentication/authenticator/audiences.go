package authenticator

import "context"

type Audiences []string

type key int

const (
	// audiencesKey is the context key for request audiences.
	audiencesKey key = iota
)

func WithAudiences(ctx context.Context, auds Audiences) context.Context {
	return context.WithValue(ctx, audiencesKey, auds)
}

func AudiencesFrom(ctx context.Context) (Audiences, bool) {
	auds, ok := ctx.Value(audiencesKey).(Audiences)
	return auds, ok
}

func (a Audiences) Has(taud string) bool {
	for _, aud := range a {
		if aud == taud {
			return true
		}
	}
	return false
}

func (a Audiences) Intersect(tauds Audiences) Audiences {
	selected := Audiences{}
	for _, taud := range tauds {
		if a.Has(taud) {
			selected = append(selected, taud)
		}
	}
	return selected
}
