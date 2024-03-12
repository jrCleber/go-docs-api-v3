package guards

import "net/http"

func ApplyGuards(guards []any) []func(http.Handler) http.Handler {
	var g []func(http.Handler) http.Handler

	for _, guard := range guards {
		switch v := guard.(type) {
		case *AuthGuard:
			g = append(g, v.CanActivate())
		}
	}

	return g
}
