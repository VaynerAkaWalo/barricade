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

func TestClientDeleteHandlerHappyPath(t *testing.T) {
	module := setupClientModule(t)

	result, err := module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
		ClientType:  ClientTypeConfidential,
	})
	assert.NoError(t, err)

	handler := ClientHttpHandler{ClientService: module.service}

	ctx := context.WithValue(context.Background(), xhttp.UserId, TEST_CLIENT_OWNER_ID)
	req := httptest.NewRequest("DELETE", "/v1/oauth2/clients/"+string(result.Client.Id), nil).WithContext(ctx)
	req.SetPathValue("id", string(result.Client.Id))
	rec := httptest.NewRecorder()

	err = handler.Delete(rec, req)
	assert.NoError(t, err)
	assert.Equal(t, 204, rec.Code)

	_, err = module.repository.FindById(context.Background(), result.Client.Id)
	assert.ErrorIs(t, err, ErrClientNotFound)
}

func TestClientDeleteUnauthorizedWhenNoUserId(t *testing.T) {
	module := setupClientModule(t)

	result, err := module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
		ClientType:  ClientTypeConfidential,
	})
	assert.NoError(t, err)

	handler := ClientHttpHandler{ClientService: module.service}

	req := httptest.NewRequest("DELETE", "/v1/oauth2/clients/"+string(result.Client.Id), nil)
	req.SetPathValue("id", string(result.Client.Id))
	rec := httptest.NewRecorder()

	err = handler.Delete(rec, req)
	assert.Error(t, err)
}

func TestClientDeleteForbiddenWhenOwnerMismatch(t *testing.T) {
	module := setupClientModule(t)

	result, err := module.service.Register(context.Background(), RegisterClientParams{
		OwnerId:     TEST_CLIENT_OWNER_ID,
		Name:        TEST_CLIENT_NAME,
		Domain:      TEST_CLIENT_DOMAIN,
		RedirectURI: TEST_CLIENT_REDIRECT_URI,
		ClientType:  ClientTypeConfidential,
	})
	assert.NoError(t, err)

	handler := ClientHttpHandler{ClientService: module.service}

	ctx := context.WithValue(context.Background(), xhttp.UserId, "other-owner")
	req := httptest.NewRequest("DELETE", "/v1/oauth2/clients/"+string(result.Client.Id), nil).WithContext(ctx)
	req.SetPathValue("id", string(result.Client.Id))
	rec := httptest.NewRecorder()

	err = handler.Delete(rec, req)
	assert.Error(t, err)
}
