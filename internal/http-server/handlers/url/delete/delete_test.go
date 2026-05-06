package delete

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"

	"github.com/stretchr/testify/require"
)

// TODO: добавить еще тест-кейсы
func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		respError string
		mockError error
	}{
		{
			name:      "Success",
			alias:     "google",
			mockError: nil,
			respError: "",
		},
		{
			name:      "Empty alias",
			alias:     "",
			respError: "Поле alias обязательное для заполнения",
		},
		{
			name:      "Delete Alias Error",
			alias:     "some_alias",
			respError: "Не удалось удалить alias",
			mockError: errors.New("unexpected error"),
		},
		{
			name:  "Incorrect alias",
			alias: "incorrect_alias",
		},
	}
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlDeleterMock.On("DeleteURL", tc.alias).
					Return(tc.mockError).
					Once()
			}

			handler := New(slogdiscard.NewDiscardLogger(), urlDeleterMock)

			req, err := http.NewRequest(http.MethodDelete, "/url/"+tc.alias, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, http.StatusOK)

			body := rr.Body.String()

			var resp Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

			require.True(t, urlDeleterMock.AssertCalled(t, "DeleteURL", tc.alias))

			// TODO: add more checks (проверить что было написано в status, alias и т.д.)
		})
	}
}
