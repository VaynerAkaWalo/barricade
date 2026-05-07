package identity

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/stretchr/testify/assert"
)

func TestIdentityListEmptyReturnsEmptyList(t *testing.T) {
	module := setupModule(t)

	handler := HttpHandler{Service: module.service}

	ctx := context.WithValue(context.Background(), xhttp.UserId, "test-user")
	req := httptest.NewRequest("GET", "/v1/identities", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	err := handler.List(rec, req)
	assert.NoError(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "[]\n", rec.Body.String())
}
