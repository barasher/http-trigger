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
		{"ok", "../testdata/scripts/script_0.sh", "0", "here is the body (0)\n"},
		{"ko", "../testdata/scripts/script_1.sh", "1", "here is the body (1)\n"},
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

func TestLoadConf(t *testing.T) {
	nominalConf := ServerConf{
		Port:     42,
		Commands: map[string]string{"cmdKey1": "cmd1", "cmdKey2": "cmd2"},
	}

	var tcs = []struct {
		id      string
		inFile  string
		expIsOk bool
		expConf ServerConf
	}{
		{"nominal", "../testdata/conf/nominal.json", true, nominalConf},
		{"nonExisting", "nonExisting.json", false, ServerConf{}},
		{"unparsable", "../testdata/conf/unparsable.json", false, ServerConf{}},
	}

	for _, tc := range tcs {
		t.Run(tc.id, func(t *testing.T) {
			c, err := LoadConfiguration(tc.inFile)
			if tc.expIsOk {
				assert.Nil(t, err)
				assert.Equal(t, tc.expConf, c)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
