package order

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserServiceRequest = errors.New("user service request failed")
)

type UserClient interface {
	GetUser(id int64) (UserDTO, error)
}

type HTTPUserClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewUserClient(baseURL string) *HTTPUserClient {
	return &HTTPUserClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *HTTPUserClient) GetUser(id int64) (UserDTO, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return UserDTO{}, fmt.Errorf("%w: %v", ErrUserServiceRequest, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return UserDTO{}, ErrUserNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return UserDTO{}, fmt.Errorf("%w: status %d", ErrUserServiceRequest, resp.StatusCode)
	}

	var user UserDTO
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return UserDTO{}, fmt.Errorf("%w: invalid json: %v", ErrUserServiceRequest, err)
	}

	return user, nil
}
