package delete

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/url/delete/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name           string
		alias          string
		respError      string
		mockError      error
		expectMockCall bool // явно указываем, ожидается ли вызов мока
		useRouter      bool // явно указываем, ожидается ли использование роутера
	}{
		{
			name:           "Success",
			alias:          "google",
			mockError:      nil,
			respError:      "",
			expectMockCall: true,
			useRouter:      true,
		},
		{
			name:           "Empty alias",
			alias:          "",
			respError:      "Поле alias обязательное для заполнения",
			expectMockCall: false,
			useRouter:      false, // новое поле
		},
		{
			name:           "Delete Alias Error",
			alias:          "some_alias",
			respError:      "внутренняя ошибка", // исправлено под реальный текст в delete.go
			mockError:      errors.New("unexpected error"),
			expectMockCall: true,
			useRouter:      true,
		},
		{
			name:           "URL Not Found",
			alias:          "missing_alias",
			respError:      "url не найден",
			mockError:      storage.ErrURLNotFound,
			expectMockCall: true,
			useRouter:      true,
		},
		{
			name:           "Incorrect alias",
			alias:          "incorrect_alias",
			mockError:      nil,
			respError:      "",
			expectMockCall: true,
			useRouter:      true,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.expectMockCall {
				urlDeleterMock.On("DeleteURL", tc.alias).
					Return(tc.mockError).
					Once()
			}

			// Используем chi-роутер, чтобы chi.URLParam корректно парсил {alias}
			r := chi.NewRouter()
			r.Delete("/url/{alias}", New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

			var url string
			if tc.alias == "" {
				url = "/url/"
			} else {
				url = "/url/" + tc.alias
			}

			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()

			if tc.useRouter {
				r := chi.NewRouter()
				r.Delete("/url/{alias}", New(slogdiscard.NewDiscardLogger(), urlDeleterMock))
				r.ServeHTTP(rr, req)
			} else {
				handler := New(slogdiscard.NewDiscardLogger(), urlDeleterMock)
				handler.ServeHTTP(rr, req)
			}

			require.Equal(t, http.StatusOK, rr.Code)

			body := rr.Body.String()

			var resp Response
			require.NoError(t, json.Unmarshal([]byte(body), &resp))
			require.Equal(t, tc.respError, resp.Error)

			if tc.expectMockCall {
				urlDeleterMock.AssertCalled(t, "DeleteURL", tc.alias)
			}
		})
	}
}
