package auth

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestParseToken(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	tests := []struct {
		name    string
		request *Request
		want    string
		wantErr bool
	}{
		{
			name:    "Valid Bearer token",
			request: &Request{Token: "Bearer BChTYbJcE3F"},
			want:    "BChTYbJcE3F",
			wantErr: false,
		},
		{
			name: "Empty Bearer token",
			request: &Request{
				Token: "Bearer ",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Empty Bearer token no whitespace",
			request: &Request{
				Token: "Bearer",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Empty request Token",
			request: &Request{
				Token: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Misspelled Bearer",
			request: &Request{
				Token: "Bearr BChTYbJcE3F",
			},
			want:    "",
			wantErr: true,
		},
	}
	builtin := Builtin{}
	for _, test := range tests {
		token, err := builtin.parseToken(test.request)
		g.Expect(test.wantErr).To(gomega.Equal(err != nil))
		g.Expect(test.want).To(gomega.Equal(token))
	}
}
