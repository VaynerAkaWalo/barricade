package oauth2

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/VaynerAkaWalo/go-toolkit/xhttp"
	"github.com/stretchr/testify/assert"
)

func TestClientListEmptyReturnsEmptyList(t *testing.T) {
	module := setupClientModule(t)

	handler := ClientHttpHandler{ClientService: module.service}

	ctx := context.WithValue(context.Background(), xhttp.UserId, TEST_CLIENT_OWNER_ID)
	req := httptest.NewRequest("GET", "/v1/oauth2/clients", nil).WithContext(ctx)
	rec := httptest.NewRecorder()

	err := handler.List(rec, req)
	assert.NoError(t, err)
	assert.Equal(t, 200, rec.Code)
	assert.Equal(t, "[]\n", rec.Body.String())
}
