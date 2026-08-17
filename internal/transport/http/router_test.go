package httptransport

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicationRequiresRoleAndRevision(t *testing.T) {
	engine := New(Services{Ready: func() error { return nil }})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/standards/quality/versions/4/publications", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("anonymous publication status=%d body=%s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodPost, "/api/v1/standards/quality/versions/4/publications", nil)
	request.Header.Set("X-Actor-Role", "template_approver")
	response = httptest.NewRecorder()
	engine.ServeHTTP(response, request)
	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing revision status=%d body=%s", response.Code, response.Body.String())
	}
}
