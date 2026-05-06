package update

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/update/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"

	"github.com/stretchr/testify/require"
)

// TODO: добавить еще тест-кейсы
func TestUpdateHandler(t *testing.T) {
	cases := []struct {
		name      string
		oldAlias  string
		newAlias  string
		respError string
		mockError error
	}{
		{
			name:     "Success",
			oldAlias: "ml",
			newAlias: "masters",
		},
		{
			name:      "Empty old alias",
			oldAlias:  "",
			newAlias:  "some_new_alias",
			respError: "Поле OldAlias обязательное для заполнения",
		},
		{
			name:      "Empty new alias",
			oldAlias:  "ml",
			newAlias:  "",
			respError: "Поле NewAlias обязательное для заполнения",
		},
		{
			name:      "Update Alias Error",
			oldAlias:  "masters",
			newAlias:  "ml",
			respError: "Не удалось обновить alias",
			mockError: errors.New("unexpected error"),
		},
		{
			name:     "Incorrect old alias",
			oldAlias: "incorrect_old_alias",
			newAlias: "some_new_alias",
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlUpdaterMock := mocks.NewUrlUpdater(t)

			if tc.respError == "" || tc.mockError != nil {
				urlUpdaterMock.On("UpdateURL", tc.oldAlias, tc.newAlias).
					Return(tc.mockError).
					Once()
			}

			handler := New(slogdiscard.NewDiscardLogger(), urlUpdaterMock)

			input := fmt.Sprintf(`{"oldAlias": "%s", "newAlias": "%s"}`, tc.oldAlias, tc.newAlias)

			req, err := http.NewRequest(http.MethodPut, "/update", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			var resp Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

			// TODO: add more checks (проверить что было написано в status, alias и т.д.)
		})
	}
}
