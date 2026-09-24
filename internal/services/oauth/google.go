package oauth

import (
	"IAM-server/internal/connections"
	"IAM-server/internal/models"
	"context"
	"fmt"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type Service struct {
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func GoogleOauthService(ctx context.Context) (*Service, error) {
	provider, err := oidc.NewProvider(ctx, "https://accounts.google.com")
	if err != nil {
		return nil, err
	}
	return &Service{
		oauth: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
			Endpoint:     provider.Endpoint(),
		},
		verifier: provider.Verifier(
			&oidc.Config{
				ClientID: os.Getenv("GOOGLE_CLIENT_ID"),
			}),
	}, nil
}

func (s *Service) GetAuthCodeURL(state string) string {
	return s.oauth.AuthCodeURL(state)
}

func (s *Service) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.oauth.Exchange(ctx, code)
}

func (s *Service) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return s.verifier.Verify(ctx, rawIDToken)
}

// FindOrCreateByGoogle looks up a customer by email.
// If the customer exists, it updates the name and photo URL from the Google claims and returns the customer.
// If the customer does not exist, it creates a new one using the provided Google profile information.
func FindOrCreateByGoogle(ctx context.Context, email, name, picture string) (*models.Customer, error) {
	var customer models.Customer
	err := connections.DB.WithContext(ctx).Where("email = ?", email).First(&customer).Error
	if err == nil {
		// Existing user — update profile fields from Google if needed
		updates := map[string]any{}
		if customer.Name == "" && name != "" {
			updates["name"] = name
		}
		if (customer.PhotoURL == nil || *customer.PhotoURL == "") && picture != "" {
			updates["photo_url"] = picture
		}
		if len(updates) > 0 {
			connections.DB.WithContext(ctx).Model(&customer).Updates(updates)
		}
		return &customer, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to look up user: %w", err)
	}

	// New user — create
	customer = models.Customer{
		Name:     name,
		Email:    email,
		PhotoURL: &picture,
	}
	if err := connections.DB.WithContext(ctx).Create(&customer).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &customer, nil
}
