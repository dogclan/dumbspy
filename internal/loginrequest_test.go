package internal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGamespyLoginRequest_Validate(t *testing.T) {
	type test struct {
		name                string
		prepareLoginRequest func(req *GamespyLoginRequest)
		wantErrContains     string
	}

	tests := []test{
		{
			name:                "passes for valid request",
			prepareLoginRequest: func(req *GamespyLoginRequest) {},
		},
		{
			name: "fails for non-zero string login",
			prepareLoginRequest: func(req *GamespyLoginRequest) {
				req.Login = ToPointer("some-string")
			},
			wantErrContains: "validation for 'Login' failed on the 'len' tag",
		},
		{
			name: "fails for wrong length challenge",
			prepareLoginRequest: func(req *GamespyLoginRequest) {
				req.Challenge = strings.Repeat("a", 15)
			},
			wantErrContains: "validation for 'Challenge' failed on the 'len' tag",
		},
		{
			name: "fails for non-numeric port",
			prepareLoginRequest: func(req *GamespyLoginRequest) {
				req.Port = "not-numeric"
			},
			wantErrContains: "validation for 'Port' failed on the 'numeric' tag",
		},
		{
			name: "fails for non-numeric product id",
			prepareLoginRequest: func(req *GamespyLoginRequest) {
				req.ProductID = "not-numeric"
			},
			wantErrContains: "validation for 'ProductID' failed on the 'numeric' tag",
		},
		{
			name: "fails for zero-string game name",
			prepareLoginRequest: func(req *GamespyLoginRequest) {
				req.GameName = ""
			},
			wantErrContains: "validation for 'GameName' failed on the 'min' tag",
		},
		{
			name: "fails for non-numeric namespace id",
			prepareLoginRequest: func(req *GamespyLoginRequest) {
				req.NamespaceID = "not-numeric"
			},
			wantErrContains: "validation for 'NamespaceID' failed on the 'numeric' tag",
		},
		{
			name: "fails for non-numeric sdk revision",
			prepareLoginRequest: func(req *GamespyLoginRequest) {
				req.SDKRevision = "not-numeric"
			},
			wantErrContains: "validation for 'SDKRevision' failed on the 'numeric' tag",
		},
		{
			name: "fails for non-numeric id",
			prepareLoginRequest: func(req *GamespyLoginRequest) {
				req.ID = "not-numeric"
			},
			wantErrContains: "validation for 'ID' failed on the 'numeric' tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// GIVEN
			req := &GamespyLoginRequest{
				Login:       ToPointer(""),
				Challenge:   "YJk5UFExKBwn0PEpOpinWHsRCDcfejyJ",
				UniqueNick:  "some-nick",
				Response:    "1c5a1eb9e75006ec317bd6d8a2c09969",
				Port:        "2475",
				ProductID:   "10439",
				GameName:    "battlefield2",
				NamespaceID: "12",
				SDKRevision: "3",
				ID:          "1",
			}
			tt.prepareLoginRequest(req)

			// WHEN
			err := req.Validate()

			// THEN
			if tt.wantErrContains != "" {
				assert.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
