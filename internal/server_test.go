package internal

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestExecHandler(t *testing.T) {
	var tcs = []struct {
		cmdKey        string
		cmd           string
		expReturnCode string
		expBody       string
	}{
		{"ok", "../testdata/script_0.sh", "0", "here is the body (0)\n"},
		{"ko", "../testdata/script_1.sh", "1", "here is the body (1)\n"},
		{"not_found", "non_existing.sh", "-1", ""},
	}

	for _, tc := range tcs {
		t.Run(tc.cmdKey, func(t *testing.T) {
			h := execHandler{
				cmdKey: tc.cmdKey,
				cmd:    tc.cmd,
			}
			path := fmt.Sprintf("/exec/%s", tc.cmdKey)
			req, err := http.NewRequest(http.MethodPost, path, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(path, h).Methods(http.MethodPost)
			router.ServeHTTP(rr, req)

			assert.Equal(t, 200, rr.Code)
			assert.Equal(t, tc.expReturnCode, rr.HeaderMap.Get(headerExitCode))
			assert.NotEqual(t, "", rr.HeaderMap.Get(headerDuration))
			assert.Equal(t, tc.expBody, rr.Body.String())
		})
	}
}

/*func TestLoadConf(t *testing.T) {
	var tcs = []struct {
		cmdKey        string
		cmd           string
		expReturnCode string
	}{
		{"ok", "../testdata/script_0.sh", "0"},
		{"ko", "../testdata/script_1.sh", "1"},
		{"not_found", "non_existing.sh", "-1"},
	}

	for _, tc := range tcs {
		t.Run(tc.cmdKey, func(t *testing.T) {
		}
	}
}*/
