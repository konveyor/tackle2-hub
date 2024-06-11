package auth

import "testing"

func TestParseToken(t *testing.T) {
	type args struct {
		requestToken string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Valid Bearer token",
			args:    args{requestToken: "Bearer AAAAAAAAAAAAAAAAAAAAMLheAAAAAAA0%2BuSeid%2BULvsea4JtiGRiSDSJSI%3DEUifiRBkKG5E2XzMDjRfl76ZC9Ub0wnz4XsNiRVBChTYbJcE3F"},
			want:    "AAAAAAAAAAAAAAAAAAAAMLheAAAAAAA0%2BuSeid%2BULvsea4JtiGRiSDSJSI%3DEUifiRBkKG5E2XzMDjRfl76ZC9Ub0wnz4XsNiRVBChTYbJcE3F",
			wantErr: false,
		},
		{
			name: "Empty Bearer token",
			args: args{
				requestToken: "Bearer ",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Empty Bearer token no whitespace",
			args: args{
				requestToken: "Bearer",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Empty request Token",
			args: args{
				requestToken: "",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Misspelled Bearer",
			args: args{
				requestToken: "Bearr AAAAAAAAAAAAAAAAAAAAMLheAAAAAAA0%2BuSeid%2BULvsea4JtiGRiSDSJSI%3DEUifiRBkKG5E2XzMDjRfl76ZC9Ub0wnz4XsNiRVBChTYbJcE3F",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseToken(tt.args.requestToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseToken() = %v, want %v", got, tt.want)
			}
		})
	}
}
